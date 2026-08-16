package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func ExecuteSequence(seq *Sequence) {
	if seq == nil || len(seq.Nodes) == 0 {
		return
	}

	var lastSuccess bool = true // Initial state

	for i, node := range seq.Nodes {
		// Determine if we should execute based on previous success and current Op
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
}

// executePipeline returns true if successful (exit code 0), false otherwise
func executePipeline(pipeline *Pipeline) bool {
	if pipeline == nil || len(pipeline.Commands) == 0 {
		return true
	}

	// Handle Built-ins (like cd)
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
	}

	var cmds []*exec.Cmd

	for _, astCmd := range pipeline.Commands {
		if len(astCmd.Args) == 0 {
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

		cmd := exec.Command(expandedArgs[0], expandedArgs[1:]...)
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
		
		// We only care about the exit status of the LAST command in the pipeline
		if i == len(cmds)-1 {
			if err != nil {
				success = false
			}
		}
	}

	return success
}
