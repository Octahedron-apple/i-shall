package tests

import (
	"reflect"
	"testing"
	"i-shall/shell"
)

func TestParser(t *testing.T) {
	tests := []struct {
		input    string
		expected *shell.Sequence
	}{
		{
			input: "ls -l",
			expected: &shell.Sequence{
				Nodes: []*shell.SequenceNode{
					{
						Op: shell.OpNone,
						Pipeline: &shell.Pipeline{
							Commands: []*shell.Command{
								{Args: []string{"ls", "-l"}},
							},
						},
					},
				},
			},
		},
		{
			input: "cat < input.txt | grep a > output.txt",
			expected: &shell.Sequence{
				Nodes: []*shell.SequenceNode{
					{
						Op: shell.OpNone,
						Pipeline: &shell.Pipeline{
							Commands: []*shell.Command{
								{Args: []string{"cat"}, RedirectIn: "input.txt"},
								{Args: []string{"grep", "a"}, RedirectOut: "output.txt", RedirectAppend: false},
							},
						},
					},
				},
			},
		},
		{
			input: "build && run || echo failed",
			expected: &shell.Sequence{
				Nodes: []*shell.SequenceNode{
					{
						Op: shell.OpNone,
						Pipeline: &shell.Pipeline{
							Commands: []*shell.Command{
								{Args: []string{"build"}},
							},
						},
					},
					{
						Op: shell.OpAnd,
						Pipeline: &shell.Pipeline{
							Commands: []*shell.Command{
								{Args: []string{"run"}},
							},
						},
					},
					{
						Op: shell.OpOr,
						Pipeline: &shell.Pipeline{
							Commands: []*shell.Command{
								{Args: []string{"echo", "failed"}},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		lexer := shell.NewLexer(tt.input)
		parser := shell.NewParser(lexer)
		seq := parser.ParseSequence()

		if !reflect.DeepEqual(seq, tt.expected) {
			t.Errorf("For input %q, expected %+v, got %+v", tt.input, tt.expected, seq)
		}
	}
}
