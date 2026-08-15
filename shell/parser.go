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

func (p *Parser) ParsePipeline() *Pipeline {
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
	if p.current.Type == TokenEOF || p.current.Type == TokenPipe {
		return nil
	}

	cmd := &Command{}

	for p.current.Type != TokenEOF && p.current.Type != TokenPipe {
		switch p.current.Type {
		case TokenWord:
			cmd.Args = append(cmd.Args, p.current.Value)
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
