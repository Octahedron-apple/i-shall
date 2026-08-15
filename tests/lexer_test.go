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
				{shell.TokenWord, "ls"},
				{shell.TokenWord, "-l"},
				{shell.TokenEOF, ""},
			},
		},
		{
			input: "echo 'hello world' | grep a",
			expected: []shell.Token{
				{shell.TokenWord, "echo"},
				{shell.TokenWord, "hello world"},
				{shell.TokenPipe, "|"},
				{shell.TokenWord, "grep"},
				{shell.TokenWord, "a"},
				{shell.TokenEOF, ""},
			},
		},
		{
			input: "cat<file.txt>out.txt>>log.txt",
			expected: []shell.Token{
				{shell.TokenWord, "cat"},
				{shell.TokenRedirectIn, "<"},
				{shell.TokenWord, "file.txt"},
				{shell.TokenRedirectOut, ">"},
				{shell.TokenWord, "out.txt"},
				{shell.TokenRedirectAppend, ">>"},
				{shell.TokenWord, "log.txt"},
				{shell.TokenEOF, ""},
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
