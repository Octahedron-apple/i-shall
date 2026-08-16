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
				{shell.TokenWord, "ls", false},
				{shell.TokenWord, "-l", false},
				{shell.TokenEOF, "", false},
			},
		},
		{
			input: "echo 'hello world' | grep a",
			expected: []shell.Token{
				{shell.TokenWord, "echo", false},
				{shell.TokenWord, "hello world", false},
				{shell.TokenPipe, "|", false},
				{shell.TokenWord, "grep", false},
				{shell.TokenWord, "a", false},
				{shell.TokenEOF, "", false},
			},
		},
		{
			input: "cat<file.txt>out.txt>>log.txt",
			expected: []shell.Token{
				{shell.TokenWord, "cat", false},
				{shell.TokenRedirectIn, "<", false},
				{shell.TokenWord, "file.txt", false},
				{shell.TokenRedirectOut, ">", false},
				{shell.TokenWord, "out.txt", false},
				{shell.TokenRedirectAppend, ">>", false},
				{shell.TokenWord, "log.txt", false},
				{shell.TokenEOF, "", false},
			},
		},
		{
			input: "ls *.go 'not*.go'",
			expected: []shell.Token{
				{shell.TokenWord, "ls", false},
				{shell.TokenWord, "*.go", true},
				{shell.TokenWord, "not*.go", false},
				{shell.TokenEOF, "", false},
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
			if tok.Type != tt.expected[i].Type || tok.Value != tt.expected[i].Value {
				t.Errorf("For input %q at token %d, expected {%v, %q}, got {%v, %q}",
					tt.input, i, tt.expected[i].Type, tt.expected[i].Value, tok.Type, tok.Value)
			}
		}
	}
}
