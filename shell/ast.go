package shell

type Command struct {
	Args           []string
	RedirectIn     string
	RedirectOut    string
	RedirectAppend bool
}

type Pipeline struct {
	Commands []*Command
}
