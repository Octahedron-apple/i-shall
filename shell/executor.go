package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func ExecuteScript(script *Script) {
	if script == nil {
		return
	}
	for _, stmt := range script.Statements {
		ExecuteStatement(stmt)
	}
}

func ExecuteStatement(stmt Statement) bool {
	if stmt == nil {
		return true
	}

	switch s := stmt.(type) {
	case *Sequence:
		return ExecuteSequence(s)
	case *IfControl:
		success := ExecuteSequence(s.Condition)
		if success {
			ExecuteScript(s.Body)
			return true // or return the exit code of Body
		}

		for _, elif := range s.Elifs {
			if ExecuteSequence(elif.Condition) {
				ExecuteScript(elif.Body)
				return true
			}
		}

		if s.ElseBody != nil {
			ExecuteScript(s.ElseBody)
			return true
		}
		return true
	case *WhileControl:
		for {
			if !ExecuteSequence(s.Condition) {
				break
			}
			ExecuteScript(s.Body)
		}
		return true
	}
	return false
}

func ExecuteSequence(seq *Sequence) bool {
	if seq == nil || len(seq.Nodes) == 0 {
		return true
	}

	var lastSuccess bool = true

	for i, node := range seq.Nodes {
		if i > 0 {
			if node.Op == OpAnd && !lastSuccess {
				continue
			}
			if node.Op == OpOr && lastSuccess {
				continue
			}
		}

		lastSuccess = executePipeline(node.Pipeline)
	}
	
	return lastSuccess
}

// executePipeline returns true if successful (exit code 0), false otherwise
func executePipeline(pipeline *Pipeline) bool {
	if pipeline == nil || len(pipeline.Commands) == 0 {
		return true
	}

	// Handle Built-ins
	firstCmd := pipeline.Commands[0]
	if len(firstCmd.Args) > 0 && firstCmd.Args[0].Value == "cd" {
		var targetDir string
		if len(firstCmd.Args) > 1 {
			targetDir = firstCmd.Args[1].Value
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error getting home directory:", err)
				return false
			}
			targetDir = home
		}

		err := os.Chdir(targetDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return false
		}
		return true
	} else if len(firstCmd.Args) > 0 && firstCmd.Args[0].Value == "source" {
		if len(firstCmd.Args) < 2 {
			fmt.Fprintln(os.Stderr, "source: missing argument")
			return false
		}
		
		fileToSource := firstCmd.Args[1].Value
		if firstCmd.Args[1].IsGlobbable {
			matches, _ := filepath.Glob(fileToSource)
			if len(matches) > 0 {
				fileToSource = matches[0]
			}
		}

		content, err := os.ReadFile(fileToSource)
		if err != nil {
			fmt.Fprintln(os.Stderr, "source:", err)
			return false
		}
		
		lexer := NewLexer(string(content))
		parser := NewParser(lexer)
		script := parser.ParseScript()
		ExecuteScript(script)
		
		return true
	}

	var cmds []*exec.Cmd

	for _, astCmd := range pipeline.Commands {
		if !astCmd.IsSubshell && len(astCmd.Args) == 0 {
			continue
		}
		
		var expandedArgs []string
		for _, arg := range astCmd.Args {
			if arg.IsGlobbable {
				matches, err := filepath.Glob(arg.Value)
				if err == nil && len(matches) > 0 {
					expandedArgs = append(expandedArgs, matches...)
				} else {
					expandedArgs = append(expandedArgs, arg.Value)
				}
			} else {
				expandedArgs = append(expandedArgs, arg.Value)
			}
		}

		var cmd *exec.Cmd
		if astCmd.IsSubshell {
			cmd = exec.Command(os.Args[0], "-c", astCmd.SubshellString)
		} else {
			cmd = exec.Command(expandedArgs[0], expandedArgs[1:]...)
		}
		cmd.Stderr = os.Stderr

		if astCmd.RedirectIn != "" {
			file, err := os.Open(astCmd.RedirectIn)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error opening input file:", err)
				return false
			}
			cmd.Stdin = file
		}

		if astCmd.RedirectOut != "" {
			flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
			if astCmd.RedirectAppend {
				flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
			}
			file, err := os.OpenFile(astCmd.RedirectOut, flags, 0644)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error opening output file:", err)
				return false
			}
			cmd.Stdout = file
		}

		cmds = append(cmds, cmd)
	}

	if len(cmds) == 0 {
		return true
	}

	for i := 0; i < len(cmds)-1; i++ {
		stdout, err := cmds[i].StdoutPipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error setting up pipe:", err)
			return false
		}
		cmds[i+1].Stdin = stdout
	}

	if cmds[len(cmds)-1].Stdout == nil {
		cmds[len(cmds)-1].Stdout = os.Stdout
	}
	
	if cmds[0].Stdin == nil {
		cmds[0].Stdin = os.Stdin
	}

	for _, cmd := range cmds {
		err := cmd.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error starting command:", err)
			return false
		}
	}

	var success bool = true
	for i, cmd := range cmds {
		err := cmd.Wait()
		if i == len(cmds)-1 {
			if err != nil {
				success = false
			}
		}
	}

	return success
}
