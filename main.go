package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strings"

	"i-shall/shell"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	currentUser, err := user.Current()
	username := "unknown"
	if err == nil {
		username = currentUser.Username
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	for {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "unknown"
		}

		fmt.Printf("%s@%s@ishall:%s> ", username, hostname, cwd)
		
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

		lexer := shell.NewLexer(input)
		parser := shell.NewParser(lexer)
		pipeline := parser.ParsePipeline()

		shell.ExecutePipeline(pipeline)
	}
}
