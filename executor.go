package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// parseArgs handles quote parsing (single and double quotes)
func parseArgs(commandString string) []string {
	var args []string
	var currentArg strings.Builder
	var inSingleQuote bool
	var inDoubleQuote bool

	for i := 0; i < len(commandString); i++ {
		char := commandString[i]

		switch char {
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			} else {
				currentArg.WriteByte(char)
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			} else {
				currentArg.WriteByte(char)
			}
		case ' ', '\t':
			if inSingleQuote || inDoubleQuote {
				currentArg.WriteByte(char)
			} else if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
		default:
			currentArg.WriteByte(char)
		}
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args
}

// executePipeline chains commands with pipes and executes them
func executePipeline(commands []string) {
	if len(commands) == 0 || commands[0] == "" {
		return
	}

	var cmds []*exec.Cmd

	for _, cmdStr := range commands {
		args := parseArgs(cmdStr)
		if len(args) == 0 {
			continue
		}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stderr = os.Stderr
		cmds = append(cmds, cmd)
	}

	if len(cmds) == 0 {
		return
	}

	// Setup pipes
	for i := 0; i < len(cmds)-1; i++ {
		stdout, err := cmds[i].StdoutPipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error setting up pipe:", err)
			return
		}
		cmds[i+1].Stdin = stdout
	}

	// The last command's output goes to standard output
	cmds[len(cmds)-1].Stdout = os.Stdout

	// Start all commands
	for _, cmd := range cmds {
		err := cmd.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error starting command:", err)
			return
		}
	}

	// Wait for all commands to finish
	for _, cmd := range cmds {
		err := cmd.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Command finished with error:", err)
		}
	}
}
