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

// Statement interface allows Script to hold multiple types of structures
type Statement interface {
	isStatement()
}

type Sequence struct {
	Nodes []*SequenceNode
}
func (s *Sequence) isStatement() {}

type ElifBlock struct {
	Condition *Sequence
	Body      *Script
}

type IfControl struct {
	Condition *Sequence
	Body      *Script
	Elifs     []*ElifBlock
	ElseBody  *Script
}
func (s *IfControl) isStatement() {}

type WhileControl struct {
	Condition *Sequence
	Body      *Script
}
func (s *WhileControl) isStatement() {}

type Script struct {
	Statements []Statement
}
