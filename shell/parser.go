package shell

type Parser struct {
	lexer   *Lexer
	current Token
}

func NewParser(lexer *Lexer) *Parser {
	p := &Parser{lexer: lexer}
	p.advance()
	return p
}

func (p *Parser) advance() {
	p.current = p.lexer.NextToken()
}

func (p *Parser) ParseSequence() *Sequence {
	seq := &Sequence{}

	if p.current.Type == TokenEOF {
		return seq
	}

	var currentOp Operator = OpNone

	for p.current.Type != TokenEOF {
		pipeline := p.parsePipeline()
		if pipeline != nil && len(pipeline.Commands) > 0 {
			seq.Nodes = append(seq.Nodes, &SequenceNode{
				Op:       currentOp,
				Pipeline: pipeline,
			})
		}

		if p.current.Type == TokenAnd {
			currentOp = OpAnd
			p.advance()
		} else if p.current.Type == TokenOr {
			currentOp = OpOr
			p.advance()
		} else {
			break
		}
	}

	return seq
}

func (p *Parser) parsePipeline() *Pipeline {
	pipeline := &Pipeline{}
	
	if p.current.Type == TokenEOF {
		return pipeline
	}

	for {
		cmd := p.parseCommand()
		if cmd != nil {
			pipeline.Commands = append(pipeline.Commands, cmd)
		}

		if p.current.Type == TokenPipe {
			p.advance()
			continue
		}
		
		break
	}

	return pipeline
}

func (p *Parser) parseCommand() *Command {
	if p.current.Type == TokenEOF || p.current.Type == TokenPipe || p.current.Type == TokenAnd || p.current.Type == TokenOr {
		return nil
	}

	cmd := &Command{}

	for p.current.Type != TokenEOF && p.current.Type != TokenPipe && p.current.Type != TokenAnd && p.current.Type != TokenOr {
		switch p.current.Type {
		case TokenWord:
			cmd.Args = append(cmd.Args, Arg{Value: p.current.Value, IsGlobbable: p.current.IsGlobbable})
			p.advance()
		case TokenRedirectIn:
			p.advance()
			if p.current.Type == TokenWord {
				cmd.RedirectIn = p.current.Value
				p.advance()
			}
		case TokenRedirectOut:
			p.advance()
			if p.current.Type == TokenWord {
				cmd.RedirectOut = p.current.Value
				cmd.RedirectAppend = false
				p.advance()
			}
		case TokenRedirectAppend:
			p.advance()
			if p.current.Type == TokenWord {
				cmd.RedirectOut = p.current.Value
				cmd.RedirectAppend = true
				p.advance()
			}
		default:
			p.advance()
		}
	}

	return cmd
}
