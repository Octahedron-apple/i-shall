package shell

type Operator int

const (
	OpNone Operator = iota
	OpAnd
	OpOr
)

type Arg struct {
	Value       string
	IsGlobbable bool
}

type Command struct {
	Args           []Arg
	RedirectIn     string
	RedirectOut    string
	RedirectAppend bool
	IsSubshell     bool
	SubshellString string
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
