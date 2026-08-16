package main

import (
	"flag"
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
	cFlag := flag.String("c", "", "Command to execute")
	flag.Parse()

	if *cFlag != "" {
		lexer := shell.NewLexer(*cFlag)
		parser := shell.NewParser(lexer)
		script, parseErr := parser.ParseScript()
		if parseErr == nil {
			shell.ExecuteScript(script)
		}
		return
	}

	currentUser, err := user.Current()
	username := "unknown"
	if err == nil {
		username = currentUser.Username
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		rcPath := filepath.Join(homeDir, ".ishallrc")
		if _, err := os.Stat(rcPath); err == nil {
			lexer := shell.NewLexer("source " + rcPath)
			parser := shell.NewParser(lexer)
			script, parseErr := parser.ParseScript()
			if parseErr == nil {
				shell.ExecuteScript(script)
			}
		}
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

	var buffer strings.Builder

	for {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "unknown"
		}

		if buffer.Len() > 0 {
			rl.SetPrompt("> ")
		} else {
			prompt := fmt.Sprintf("\033[32m%s@%s\033[0m@ishall:\033[34m%s\033[0m> ", username, hostname, cwd)
			rl.SetPrompt(prompt)
		}

		input, err := rl.Readline()
		if err != nil { // EOF or Ctrl-C
			break
		}

		input = strings.TrimSpace(input)
		if input == "" && buffer.Len() == 0 {
			continue
		}
		if input == "exit" && buffer.Len() == 0 {
			break
		}

		if buffer.Len() > 0 {
			buffer.WriteString("\n")
		}
		buffer.WriteString(input)

		lexer := shell.NewLexer(buffer.String())
		parser := shell.NewParser(lexer)
		script, parseErr := parser.ParseScript()

		if parseErr == shell.ErrIncomplete {
			continue
		}

		if parseErr != nil {
			fmt.Fprintln(os.Stderr, "Syntax error:", parseErr)
		} else {
			shell.ExecuteScript(script)
		}
		
		buffer.Reset()
	}
}
