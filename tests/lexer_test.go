package tests

import (
	"testing"
	"i-shall/shell"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		input    string
		expected []shell.Token
	}{
		{
			input: "ls -l",
			expected: []shell.Token{
				{Type: shell.TokenWord, Value: "ls"},
				{Type: shell.TokenWord, Value: "-l"},
				{Type: shell.TokenEOF, Value: ""},
			},
		},
		{
			input: "echo 'hello world' | grep a",
			expected: []shell.Token{
				{Type: shell.TokenWord, Value: "echo"},
				{Type: shell.TokenWord, Value: "hello world", IsSingleQuoted: true},
				{Type: shell.TokenPipe, Value: "|"},
				{Type: shell.TokenWord, Value: "grep"},
				{Type: shell.TokenWord, Value: "a"},
				{Type: shell.TokenEOF, Value: ""},
			},
		},
		{
			input: "cat<file.txt>out.txt>>log.txt",
			expected: []shell.Token{
				{Type: shell.TokenWord, Value: "cat"},
				{Type: shell.TokenRedirectIn, Value: "<"},
				{Type: shell.TokenWord, Value: "file.txt"},
				{Type: shell.TokenRedirectOut, Value: ">"},
				{Type: shell.TokenWord, Value: "out.txt"},
				{Type: shell.TokenRedirectAppend, Value: ">>"},
				{Type: shell.TokenWord, Value: "log.txt"},
				{Type: shell.TokenEOF, Value: ""},
			},
		},
		{
			input: "ls *.go 'not*.go'",
			expected: []shell.Token{
				{Type: shell.TokenWord, Value: "ls"},
				{Type: shell.TokenWord, Value: "*.go", IsGlobbable: true},
				{Type: shell.TokenWord, Value: "not*.go", IsSingleQuoted: true},
				{Type: shell.TokenEOF, Value: ""},
			},
		},
		{
			input: "(ls -l)",
			expected: []shell.Token{
				{Type: shell.TokenLParen, Value: "("},
				{Type: shell.TokenWord, Value: "ls"},
				{Type: shell.TokenWord, Value: "-l"},
				{Type: shell.TokenRParen, Value: ")"},
				{Type: shell.TokenEOF, Value: ""},
			},
		},
	}

	for _, tt := range tests {
		lexer := shell.NewLexer(tt.input)
		var tokens []shell.Token
		for {
			tok := lexer.NextToken()
			tokens = append(tokens, tok)
			if tok.Type == shell.TokenEOF {
				break
			}
		}

		if len(tokens) != len(tt.expected) {
			t.Fatalf("For input %q, expected %d tokens, got %d", tt.input, len(tt.expected), len(tokens))
		}

		for i, tok := range tokens {
			if tok.Type != tt.expected[i].Type || tok.Value != tt.expected[i].Value || tok.IsGlobbable != tt.expected[i].IsGlobbable || tok.IsSingleQuoted != tt.expected[i].IsSingleQuoted {
				t.Errorf("For input %q at token %d, expected %+v, got %+v",
					tt.input, i, tt.expected[i], tok)
			}
		}
	}
}
