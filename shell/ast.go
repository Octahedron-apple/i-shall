package shell

type Operator int

const (
	OpNone Operator = iota
	OpAnd
	OpOr
)

type Command struct {
	Args           []string
	RedirectIn     string
	RedirectOut    string
	RedirectAppend bool
}

type Pipeline struct {
	Commands []*Command
}

type SequenceNode struct {
	Op       Operator
	Pipeline *Pipeline
}

type Sequence struct {
	Nodes []*SequenceNode
}
