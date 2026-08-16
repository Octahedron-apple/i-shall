package shell

import "strings"

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
	if p.current.Type == TokenEOF || p.current.Type == TokenPipe || p.current.Type == TokenAnd || p.current.Type == TokenOr || p.current.Type == TokenRParen {
		return nil
	}

	cmd := &Command{}

	if p.current.Type == TokenLParen {
		p.advance()
		cmd.IsSubshell = true
		
		var subshellTokens []string
		parenCount := 1
		
		for p.current.Type != TokenEOF {
			if p.current.Type == TokenLParen {
				parenCount++
			} else if p.current.Type == TokenRParen {
				parenCount--
				if parenCount == 0 {
					p.advance()
					break
				}
			}
			
			val := p.current.Value
			if p.current.Type == TokenWord && strings.ContainsAny(val, " \t|&<>()*?") {
				val = "\"" + val + "\""
			}
			subshellTokens = append(subshellTokens, val)
			p.advance()
		}
		
		cmd.SubshellString = strings.Join(subshellTokens, " ")
	}

	for p.current.Type != TokenEOF && p.current.Type != TokenPipe && p.current.Type != TokenAnd && p.current.Type != TokenOr && p.current.Type != TokenRParen {
		switch p.current.Type {
		case TokenWord:
			if !cmd.IsSubshell {
				cmd.Args = append(cmd.Args, Arg{Value: p.current.Value, IsGlobbable: p.current.IsGlobbable})
			} else {
				// We don't support trailing args after a subshell except redirections
			}
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

	if cmd.IsSubshell || len(cmd.Args) > 0 {
		return cmd
	}
	return nil
}
