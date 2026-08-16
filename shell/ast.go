package shell

type Operator int

const (
	OpNone Operator = iota
	OpAnd
	OpOr
)

type Arg struct {
	Value          string
	IsGlobbable    bool
	IsDoubleQuoted bool
	IsVarRef       bool
	VarName        string
	VarType        string // $ or #
	IsArrayIdx     bool
	ArrayIndex     string
	IsCommandSub   bool
	CommandSub     string
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

type Assignment struct {
	Name     string
	IsArray  bool
	IsExport bool
	Value    Arg
	Values   []Arg
}
func (s *Assignment) isStatement() {}

type ForControl struct {
	Init      *Assignment
	Condition *MathCondition
	Increment *MathAssignment
	Body      *Script
}
func (s *ForControl) isStatement() {}

type MathCondition struct {
	Left     Arg
	Operator string
	Right    Arg
}

type MathAssignment struct {
	Name     string
	Left     Arg
	Operator string
	Right    Arg
}

type FunctionDef struct {
	Name string
	Body *Script
}
func (s *FunctionDef) isStatement() {}

type AliasDef struct {
	Name  string
	Value Arg
}
func (s *AliasDef) isStatement() {}

type Script struct {
	Statements []Statement
}
