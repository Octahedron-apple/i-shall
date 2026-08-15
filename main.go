package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"i-shall/shell"
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

		lexer := shell.NewLexer(input)
		parser := shell.NewParser(lexer)
		pipeline := parser.ParsePipeline()

		shell.ExecutePipeline(pipeline)
	}
}
