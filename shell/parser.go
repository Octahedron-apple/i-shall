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

// ParseScript is the main entry point now. It parses until EOF.
func (p *Parser) ParseScript() *Script {
	return p.parseScriptBlock(TokenEOF)
}

// parseScriptBlock parses statements until it hits the `untilToken` or EOF.
func (p *Parser) parseScriptBlock(untilTokens ...TokenType) *Script {
	script := &Script{}

	for p.current.Type != TokenEOF {
		// Check if we hit an early termination token (e.g. fi, done, elif, else)
		hit := false
		for _, t := range untilTokens {
			if p.current.Type == t {
				hit = true
				break
			}
		}
		if hit {
			break
		}

		if p.current.Type == TokenSemicolon {
			p.advance()
			continue
		}

		if p.current.Type == TokenIf {
			script.Statements = append(script.Statements, p.parseIf())
			continue
		}

		if p.current.Type == TokenWhile {
			script.Statements = append(script.Statements, p.parseWhile())
			continue
		}

		// Otherwise, it's a normal sequence
		seq := p.ParseSequence()
		if seq != nil && len(seq.Nodes) > 0 {
			script.Statements = append(script.Statements, seq)
		}
	}

	return script
}

func (p *Parser) parseIf() *IfControl {
	p.advance() // consume 'if'
	
	ifCtrl := &IfControl{}
	ifCtrl.Condition = p.ParseSequence() // parses until ';'
	
	if p.current.Type == TokenSemicolon {
		p.advance()
	}

	// Parse body until 'elif', 'else', or 'fi'
	ifCtrl.Body = p.parseScriptBlock(TokenElif, TokenElse, TokenFi)

	// Parse elifs
	for p.current.Type == TokenElif {
		p.advance() // consume 'elif'
		elifBlock := &ElifBlock{}
		elifBlock.Condition = p.ParseSequence()
		if p.current.Type == TokenSemicolon {
			p.advance()
		}
		elifBlock.Body = p.parseScriptBlock(TokenElif, TokenElse, TokenFi)
		ifCtrl.Elifs = append(ifCtrl.Elifs, elifBlock)
	}

	// Parse else
	if p.current.Type == TokenElse {
		p.advance() // consume 'else'
		ifCtrl.ElseBody = p.parseScriptBlock(TokenFi)
	}

	if p.current.Type == TokenFi {
		p.advance() // consume 'fi'
	}

	return ifCtrl
}

func (p *Parser) parseWhile() *WhileControl {
	p.advance() // consume 'while'

	whileCtrl := &WhileControl{}
	whileCtrl.Condition = p.ParseSequence()

	if p.current.Type == TokenSemicolon {
		p.advance()
	}

	whileCtrl.Body = p.parseScriptBlock(TokenDone)

	if p.current.Type == TokenDone {
		p.advance() // consume 'done'
	}

	return whileCtrl
}

func (p *Parser) ParseSequence() *Sequence {
	seq := &Sequence{}

	if p.current.Type == TokenEOF {
		return seq
	}

	var currentOp Operator = OpNone

	for p.current.Type != TokenEOF && p.current.Type != TokenSemicolon && p.current.Type != TokenFi && p.current.Type != TokenDone && p.current.Type != TokenElif && p.current.Type != TokenElse {
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
	if p.current.Type == TokenEOF || p.current.Type == TokenPipe || p.current.Type == TokenAnd || p.current.Type == TokenOr || p.current.Type == TokenRParen || p.current.Type == TokenSemicolon || p.current.Type == TokenFi || p.current.Type == TokenDone || p.current.Type == TokenElif || p.current.Type == TokenElse {
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
			if p.current.Type == TokenWord && strings.ContainsAny(val, " \t|&<>()*?;") {
				val = "\"" + val + "\""
			}
			subshellTokens = append(subshellTokens, val)
			p.advance()
		}

		cmd.SubshellString = strings.Join(subshellTokens, " ")
	}

	for p.current.Type != TokenEOF && p.current.Type != TokenPipe && p.current.Type != TokenAnd && p.current.Type != TokenOr && p.current.Type != TokenRParen && p.current.Type != TokenSemicolon && p.current.Type != TokenFi && p.current.Type != TokenDone && p.current.Type != TokenElif && p.current.Type != TokenElse {
		switch p.current.Type {
		case TokenWord:
			if !cmd.IsSubshell {
				cmd.Args = append(cmd.Args, Arg{Value: p.current.Value, IsGlobbable: p.current.IsGlobbable})
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
