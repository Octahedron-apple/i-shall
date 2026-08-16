package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type VarValue struct {
	StringValue string
	NumberValue float64
	ArrayValue  []VarValue
	IsNumber    bool
	IsArray     bool
}

var Env = make(map[string]VarValue)

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
	case *Assignment:
		executeAssignment(s)
		return true
	case *Sequence:
		return ExecuteSequence(s)
	case *IfControl:
		success := ExecuteSequence(s.Condition)
		if success {
			ExecuteScript(s.Body)
			return true
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
	case *ForControl:
		executeAssignment(s.Init)
		for {
			if !executeMathCondition(s.Condition) {
				break
			}
			ExecuteScript(s.Body)
			executeMathAssignment(s.Increment)
		}
		return true
	}
	return false
}

func parseValue(arg Arg) VarValue {
	if arg.IsVarRef {
		val := resolveVarRef(arg)
		// Return it as string, let it be reparsed below
		arg.Value = val
	}

	// Try to parse as number
	if num, err := strconv.ParseFloat(arg.Value, 64); err == nil {
		return VarValue{NumberValue: num, IsNumber: true}
	}
	return VarValue{StringValue: arg.Value}
}

func executeAssignment(assign *Assignment) {
	if assign.IsArray {
		var arr []VarValue
		for _, v := range assign.Values {
			arr = append(arr, parseValue(v))
		}
		Env[assign.Name] = VarValue{ArrayValue: arr, IsArray: true}
	} else if assign.Value.Value != "" || assign.Value.IsVarRef {
		val := parseValue(assign.Value)
		Env[assign.Name] = val
		if assign.IsExport {
			if val.IsNumber {
				os.Setenv(assign.Name, strconv.FormatFloat(val.NumberValue, 'f', -1, 64))
			} else {
				os.Setenv(assign.Name, val.StringValue)
			}
		}
	} else if assign.IsExport {
		if val, ok := Env[assign.Name]; ok {
			if val.IsNumber {
				os.Setenv(assign.Name, strconv.FormatFloat(val.NumberValue, 'f', -1, 64))
			} else {
				os.Setenv(assign.Name, val.StringValue)
			}
		}
	}
}

func executeMathCondition(cond *MathCondition) bool {
	left := parseValue(cond.Left).NumberValue
	right := parseValue(cond.Right).NumberValue

	switch cond.Operator {
	case "<": return left < right
	case ">": return left > right
	case "<=": return left <= right
	case ">=": return left >= right
	case "==": return left == right
	case "!=": return left != right
	}
	return false
}

func executeMathAssignment(inc *MathAssignment) {
	left := parseValue(inc.Left).NumberValue
	right := parseValue(inc.Right).NumberValue
	
	var res float64
	switch inc.Operator {
	case "+": res = left + right
	case "-": res = left - right
	case "*": res = left * right
	case "/": 
		if right != 0 { res = left / right }
	}
	
	Env[inc.Name] = VarValue{NumberValue: res, IsNumber: true}
}

func resolveVarRef(arg Arg) string {
	val, ok := Env[arg.VarName]
	if !ok {
		return ""
	}

	if arg.IsArrayIdx {
		idx, err := strconv.Atoi(arg.ArrayIndex)
		if err != nil || !val.IsArray || idx < 0 || idx >= len(val.ArrayValue) {
			return ""
		}
		val = val.ArrayValue[idx]
	}

	if arg.VarType == "$" {
		if val.IsNumber {
			return strconv.FormatFloat(val.NumberValue, 'f', -1, 64)
		}
		return val.StringValue
	} else if arg.VarType == "#" {
		if val.IsNumber {
			return strconv.FormatFloat(val.NumberValue, 'f', -1, 64)
		}
		return val.StringValue
	}

	return ""
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

func executePipeline(pipeline *Pipeline) bool {
	if pipeline == nil || len(pipeline.Commands) == 0 {
		return true
	}

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
		script, err := parser.ParseScript()
		if err != nil {
			fmt.Fprintln(os.Stderr, "source error:", err)
		}
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
			val := arg.Value
			if arg.IsVarRef {
				val = resolveVarRef(arg)
			}

			if arg.IsGlobbable && !arg.IsVarRef {
				matches, err := filepath.Glob(val)
				if err == nil && len(matches) > 0 {
					expandedArgs = append(expandedArgs, matches...)
				} else {
					expandedArgs = append(expandedArgs, val)
				}
			} else {
				expandedArgs = append(expandedArgs, val)
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
