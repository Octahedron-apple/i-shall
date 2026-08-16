package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"i-shall/shell"
	"github.com/chzyer/readline"
)

// FileCompleter implements readline.AutoCompleter
type FileCompleter struct{}

func (f *FileCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	// Simple file auto-completion
	str := string(line[:pos])
	words := strings.Split(str, " ")
	prefix := words[len(words)-1]

	matches, _ := filepath.Glob(prefix + "*")
	for _, match := range matches {
		newLine = append(newLine, []rune(match[len(prefix):]))
	}

	return newLine, len(prefix)
}

func main() {
	currentUser, err := user.Current()
	username := "unknown"
	if err == nil {
		username = currentUser.Username
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	historyFile := filepath.Join(os.TempDir(), "ishall_history")

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "", // We'll set this dynamically
		HistoryFile:     historyFile,
		AutoComplete:    &FileCompleter{},
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "unknown"
		}

		prompt := fmt.Sprintf("\033[32m%s@%s\033[0m@ishall:\033[34m%s\033[0m> ", username, hostname, cwd)
		rl.SetPrompt(prompt)

		input, err := rl.Readline()
		if err != nil { // EOF or Ctrl-C
			break
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
		seq := parser.ParseSequence()

		shell.ExecuteSequence(seq)
	}
}
