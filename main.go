package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("i-shall> ")
		
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if input == "exit" {
			break
		}

		// Split by pipe
		commands := strings.Split(input, "|")
		for i := range commands {
			commands[i] = strings.TrimSpace(commands[i])
		}

		executePipeline(commands)
	}
}
