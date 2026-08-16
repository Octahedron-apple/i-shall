package shell

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type VarValue struct {
	StringValue string
	NumberValue float64
	ArrayValue  []VarValue
	IsNumber    bool
	IsArray     bool
}

var Env = make(map[string]VarValue)
var Funcs = make(map[string]*Script)
var traps = make(map[string]string)
var signalChan = make(chan os.Signal, 1)

func init() {
	go func() {
		for sig := range signalChan {
			if cmdStr, ok := traps[sig.String()]; ok {
				execCommandSub(cmdStr)
			}
		}
	}()
}

func execCommandSub(scriptStr string) string {
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	cmd := exec.Command(exe, "-c", scriptStr)
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "command sub error:", err)
		return ""
	}
	return strings.TrimSuffix(string(out), "\n")
}

func expandTilde(s string) string {
	if strings.HasPrefix(s, "~/") || s == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home + s[1:]
		}
	}
	return s
}

func expandBraces(s string) []string {
	start := strings.Index(s, "{")
	if start == -1 {
		return []string{s}
	}
	end := -1
	open := 0
	for i := start; i < len(s); i++ {
		if s[i] == '{' {
			open++
		} else if s[i] == '}' {
			open--
			if open == 0 {
				end = i
				break
			}
		}
	}
	if end == -1 {
		return []string{s}
	}

	prefix := s[:start]
	suffix := s[end+1:]
	options := strings.Split(s[start+1:end], ",")

	var res []string
	for _, opt := range options {
		res = append(res, prefix+opt+suffix)
	}
	
	var finalRes []string
	for _, r := range res {
		finalRes = append(finalRes, expandBraces(r)...)
	}
	return finalRes
}

func interpolateString(s string) string {
	re := regexp.MustCompile(`([$#])([a-zA-Z_][a-zA-Z0-9_]*)`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		name := match[1:]
		if val, ok := Env[name]; ok {
			if val.IsNumber {
				return strconv.FormatFloat(val.NumberValue, 'f', -1, 64)
			}
			return val.StringValue
		}
		return match
	})
}

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
	case *FunctionDef:
		Funcs[s.Name] = s.Body
		return true
	case *AliasDef:
		val := parseValue(s.Value).StringValue
		lexer := NewLexer(val + " $args")
		parser := NewParser(lexer)
		script, err := parser.ParseScript()
		if err != nil {
			fmt.Fprintln(os.Stderr, "alias error:", err)
			return false
		}
		Funcs[s.Name] = script
		return true
	}
	return false
}

func parseValue(arg Arg) VarValue {
	if arg.IsCommandSub {
		arg.Value = execCommandSub(arg.CommandSub)
	}

	if arg.IsVarRef {
		vals := resolveVarRefMulti(arg)
		if len(vals) > 0 {
			arg.Value = vals[0]
		}
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
	} else if assign.Value.Value != "" || assign.Value.IsVarRef || assign.Value.IsCommandSub {
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

func resolveVarRefMulti(arg Arg) []string {
	val, ok := Env[arg.VarName]
	if !ok {
		return nil
	}

	if arg.IsArrayIdx {
		idx, err := strconv.Atoi(arg.ArrayIndex)
		if err != nil || !val.IsArray || idx < 0 || idx >= len(val.ArrayValue) {
			return nil
		}
		val = val.ArrayValue[idx]
	}

	if val.IsArray && !arg.IsArrayIdx {
		var res []string
		for _, v := range val.ArrayValue {
			if v.IsNumber {
				res = append(res, strconv.FormatFloat(v.NumberValue, 'f', -1, 64))
			} else {
				res = append(res, v.StringValue)
			}
		}
		return res
	}

	if arg.VarType == "$" || arg.VarType == "#" {
		if val.IsNumber {
			return []string{strconv.FormatFloat(val.NumberValue, 'f', -1, 64)}
		}
		return []string{val.StringValue}
	}

	return nil
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
			if arg.IsCommandSub {
				out := execCommandSub(arg.CommandSub)
				expandedArgs = append(expandedArgs, out)
				continue
			}

			if arg.IsVarRef {
				expandedArgs = append(expandedArgs, resolveVarRefMulti(arg)...)
			} else if arg.IsGlobbable {
				matches, err := filepath.Glob(arg.Value)
				if err == nil && len(matches) > 0 {
					expandedArgs = append(expandedArgs, matches...)
				} else {
					expandedArgs = append(expandedArgs, arg.Value)
				}
			} else {
				val := expandTilde(arg.Value)
				if arg.IsDoubleQuoted {
					val = interpolateString(val)
				}
				braced := expandBraces(val)
				expandedArgs = append(expandedArgs, braced...)
			}
		}

		if len(expandedArgs) == 0 && !astCmd.IsSubshell {
			continue
		}

		if !astCmd.IsSubshell && expandedArgs[0] == "trap" {
			if len(expandedArgs) < 3 {
				fmt.Fprintln(os.Stderr, "usage: trap <command> <signal>")
				return true
			}
			cmd := expandedArgs[1]
			sigName := expandedArgs[2]
			
			var sig os.Signal
			if sigName == "SIGINT" {
				sig = os.Interrupt
			}
			
			if sig != nil {
				signal.Notify(signalChan, sig)
				traps[sigName] = cmd
			}
			return true
		}

		// Function Execution Check
		if !astCmd.IsSubshell {
			if fnBody, ok := Funcs[expandedArgs[0]]; ok {
				// Save old args
				oldArgs, hasOldArgs := Env["args"]
				
				// Set new $args array
				var newArgs []VarValue
				for _, a := range expandedArgs[1:] {
					newArgs = append(newArgs, VarValue{StringValue: a})
				}
				Env["args"] = VarValue{ArrayValue: newArgs, IsArray: true}
				
				// Execute function body
				ExecuteScript(fnBody)
				
				// Restore old args
				if hasOldArgs {
					Env["args"] = oldArgs
				} else {
					delete(Env, "args")
				}
				return true // Functions do not currently participate properly in pipes (simplification)
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
