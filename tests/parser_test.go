package tests

import (
	"reflect"
	"testing"
	"i-shall/shell"
)

func TestParser(t *testing.T) {
	tests := []struct {
		input    string
		expected *shell.Pipeline
	}{
		{
			input: "ls -l",
			expected: &shell.Pipeline{
				Commands: []*shell.Command{
					{Args: []string{"ls", "-l"}},
				},
			},
		},
		{
			input: "cat < input.txt | grep a > output.txt",
			expected: &shell.Pipeline{
				Commands: []*shell.Command{
					{Args: []string{"cat"}, RedirectIn: "input.txt"},
					{Args: []string{"grep", "a"}, RedirectOut: "output.txt", RedirectAppend: false},
				},
			},
		},
	}

	for _, tt := range tests {
		lexer := shell.NewLexer(tt.input)
		parser := shell.NewParser(lexer)
		pipeline := parser.ParsePipeline()

		if !reflect.DeepEqual(pipeline, tt.expected) {
			t.Errorf("For input %q, expected %+v, got %+v", tt.input, tt.expected, pipeline)
		}
	}
}
