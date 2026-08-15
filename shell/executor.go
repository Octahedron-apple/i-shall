package shell

import (
	"fmt"
	"os"
	"os/exec"
)

func ExecutePipeline(pipeline *Pipeline) {
	if pipeline == nil || len(pipeline.Commands) == 0 {
		return
	}

	var cmds []*exec.Cmd

	for _, astCmd := range pipeline.Commands {
		if len(astCmd.Args) == 0 {
			continue
		}
		
		cmd := exec.Command(astCmd.Args[0], astCmd.Args[1:]...)
		cmd.Stderr = os.Stderr

		if astCmd.RedirectIn != "" {
			file, err := os.Open(astCmd.RedirectIn)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error opening input file:", err)
				return
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
				return
			}
			cmd.Stdout = file
		}

		cmds = append(cmds, cmd)
	}

	if len(cmds) == 0 {
		return
	}

	for i := 0; i < len(cmds)-1; i++ {
		stdout, err := cmds[i].StdoutPipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error setting up pipe:", err)
			return
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
			return
		}
	}

	for _, cmd := range cmds {
		err := cmd.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Command finished with error:", err)
		}
	}
}
