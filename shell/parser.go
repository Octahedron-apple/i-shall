package shell

import (
	"errors"
	"strings"
)

var ErrIncomplete = errors.New("incomplete statement")

type Parser struct {
	lexer   *Lexer
	current Token
	peek    Token
	hasPeek bool
}

func NewParser(lexer *Lexer) *Parser {
	p := &Parser{lexer: lexer}
	p.advance()
	return p
}

func (p *Parser) advance() {
	if p.hasPeek {
		p.current = p.peek
		p.hasPeek = false
	} else {
		p.current = p.lexer.NextToken()
	}
}

func (p *Parser) peekToken() Token {
	if !p.hasPeek {
		p.peek = p.lexer.NextToken()
		p.hasPeek = true
	}
	return p.peek
}

func (p *Parser) ParseScript() (*Script, error) {
	return p.parseScriptBlock(TokenEOF)
}

func (p *Parser) parseScriptBlock(untilTokens ...TokenType) (*Script, error) {
	script := &Script{}

	for p.current.Type != TokenEOF {
		if p.current.Type == TokenIncomplete {
			return script, ErrIncomplete
		}

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
			stmt, err := p.parseIf()
			if err != nil {
				return script, err
			}
			script.Statements = append(script.Statements, stmt)
			continue
		}

		if p.current.Type == TokenWhile {
			stmt, err := p.parseWhile()
			if err != nil {
				return script, err
			}
			script.Statements = append(script.Statements, stmt)
			continue
		}
		
		if p.current.Type == TokenFor {
			stmt, err := p.parseFor()
			if err != nil {
				return script, err
			}
			script.Statements = append(script.Statements, stmt)
			continue
		}

		if p.current.Type == TokenFn {
			stmt, err := p.parseFn()
			if err != nil {
				return script, err
			}
			script.Statements = append(script.Statements, stmt)
			continue
		}

		if p.current.Type == TokenAlias {
			stmt, err := p.parseAlias()
			if err != nil {
				return script, err
			}
			script.Statements = append(script.Statements, stmt)
			continue
		}

		var isExport bool
		if p.current.Type == TokenExport {
			isExport = true
			p.advance()
		}

		// Check for assignment: name = ...
		if p.current.Type == TokenWord && p.peekToken().Type == TokenAssign {
			stmt, err := p.parseAssignment(isExport)
			if err != nil {
				return script, err
			}
			script.Statements = append(script.Statements, stmt)
			continue
		} else if isExport {
			name := p.current.Value
			if strings.HasPrefix(name, "#") || strings.HasPrefix(name, "$") {
				name = name[1:]
			}
			stmt := &Assignment{Name: name, IsExport: true}
			script.Statements = append(script.Statements, stmt)
			p.advance()
			continue
		}

		seq, err := p.ParseSequence()
		if err != nil {
			return script, err
		}
		if seq != nil && len(seq.Nodes) > 0 {
			script.Statements = append(script.Statements, seq)
		}
	}

	if p.current.Type == TokenEOF {
		for _, t := range untilTokens {
			if t != TokenEOF {
				return script, ErrIncomplete
			}
		}
	}

	return script, nil
}

func (p *Parser) parseAssignment(isExport bool) (*Assignment, error) {
	name := p.current.Value
	if strings.HasPrefix(name, "#") || strings.HasPrefix(name, "$") {
		name = name[1:]
	}
	assign := &Assignment{Name: name, IsExport: isExport}
	p.advance() // consume name
	p.advance() // consume =

	if p.current.Type == TokenLParen {
		// Array assignment
		assign.IsArray = true
		p.advance() // consume (

		for p.current.Type != TokenEOF && p.current.Type != TokenRParen {
			if p.current.Type == TokenIncomplete {
				return assign, ErrIncomplete
			}
			if p.current.Type == TokenComma {
				p.advance()
				continue
			}
			
			if p.current.Type == TokenWord {
				assign.Values = append(assign.Values, p.parseArg(p.current))
				p.advance()
			} else {
				p.advance() // Unexpected token, skip for now
			}
		}

		if p.current.Type == TokenEOF {
			return assign, ErrIncomplete
		}
		p.advance() // consume )
	} else {
		// Scalar assignment
		if p.current.Type == TokenIncomplete {
			return assign, ErrIncomplete
		}
		if p.current.Type == TokenWord {
			assign.Value = p.parseArg(p.current)
			p.advance()
		}
	}
	return assign, nil
}

func (p *Parser) parseIf() (*IfControl, error) {
	p.advance() // consume 'if'
	
	ifCtrl := &IfControl{}
	var err error

	ifCtrl.Condition, err = p.ParseSequence()
	if err != nil {
		return nil, err
	}
	
	if p.current.Type == TokenSemicolon {
		p.advance()
	}

	ifCtrl.Body, err = p.parseScriptBlock(TokenElif, TokenElse, TokenFi)
	if err != nil {
		return nil, err
	}

	for p.current.Type == TokenElif {
		p.advance()
		elifBlock := &ElifBlock{}
		elifBlock.Condition, err = p.ParseSequence()
		if err != nil {
			return nil, err
		}
		if p.current.Type == TokenSemicolon {
			p.advance()
		}
		elifBlock.Body, err = p.parseScriptBlock(TokenElif, TokenElse, TokenFi)
		if err != nil {
			return nil, err
		}
		ifCtrl.Elifs = append(ifCtrl.Elifs, elifBlock)
	}

	if p.current.Type == TokenElse {
		p.advance()
		ifCtrl.ElseBody, err = p.parseScriptBlock(TokenFi)
		if err != nil {
			return nil, err
		}
	}

	if p.current.Type == TokenFi {
		p.advance() // consume 'fi'
	} else if p.current.Type == TokenEOF {
		return nil, ErrIncomplete
	}

	return ifCtrl, nil
}

func (p *Parser) parseWhile() (*WhileControl, error) {
	p.advance()

	whileCtrl := &WhileControl{}
	var err error

	whileCtrl.Condition, err = p.ParseSequence()
	if err != nil {
		return nil, err
	}

	if p.current.Type == TokenSemicolon {
		p.advance()
	}

	whileCtrl.Body, err = p.parseScriptBlock(TokenDone)
	if err != nil {
		return nil, err
	}

	if p.current.Type == TokenDone {
		p.advance()
	} else if p.current.Type == TokenEOF {
		return nil, ErrIncomplete
	}

	return whileCtrl, nil
}

func (p *Parser) parseFor() (*ForControl, error) {
	p.advance() // consume 'for'
	if p.current.Type == TokenLParen {
		p.advance()
	}

	init, err := p.parseAssignment(false)
	if err != nil {
		return nil, err
	}
	if p.current.Type == TokenSemicolon {
		p.advance()
	}

	cond := &MathCondition{}
	cond.Left = p.parseArg(p.current)
	p.advance()
	if p.current.Type == TokenRedirectIn {
		cond.Operator = "<"
	} else if p.current.Type == TokenRedirectOut {
		cond.Operator = ">"
	} else {
		cond.Operator = p.current.Value
	}
	p.advance()
	cond.Right = p.parseArg(p.current)
	p.advance()

	if p.current.Type == TokenSemicolon {
		p.advance()
	}

	inc := &MathAssignment{}
	inc.Name = p.current.Value
	if strings.HasPrefix(inc.Name, "#") || strings.HasPrefix(inc.Name, "$") {
		inc.Name = inc.Name[1:]
	}
	p.advance() // consume name
	p.advance() // consume =
	inc.Left = p.parseArg(p.current)
	p.advance()
	inc.Operator = p.current.Value
	p.advance()
	inc.Right = p.parseArg(p.current)
	p.advance()

	if p.current.Type == TokenRParen {
		p.advance()
	}
	if p.current.Type == TokenSemicolon {
		p.advance()
	}

	forCtrl := &ForControl{Init: init, Condition: cond, Increment: inc}
	forCtrl.Body, err = p.parseScriptBlock(TokenDone)
	if err != nil {
		return nil, err
	}

	if p.current.Type == TokenDone {
		p.advance()
	} else if p.current.Type == TokenEOF {
		return nil, ErrIncomplete
	}

	return forCtrl, nil
}

func (p *Parser) parseFn() (*FunctionDef, error) {
	p.advance() // consume 'fn'
	if p.current.Type != TokenWord {
		return nil, errors.New("expected function name")
	}
	fn := &FunctionDef{Name: p.current.Value}
	p.advance() // consume name

	if p.current.Type == TokenSemicolon {
		p.advance()
	}

	body, err := p.parseScriptBlock(TokenEnd)
	if err != nil {
		return nil, err
	}
	fn.Body = body

	if p.current.Type == TokenEnd {
		p.advance()
	} else if p.current.Type == TokenEOF {
		return nil, ErrIncomplete
	}

	return fn, nil
}

func (p *Parser) parseAlias() (*AliasDef, error) {
	p.advance() // consume 'alias'
	if p.current.Type != TokenWord {
		return nil, errors.New("expected alias name")
	}
	alias := &AliasDef{Name: p.current.Value}
	p.advance() // consume name

	if p.current.Type == TokenAssign {
		p.advance() // consume =
	}

	if p.current.Type == TokenIncomplete {
		return nil, ErrIncomplete
	}

	alias.Value = p.parseArg(p.current)
	p.advance()

	return alias, nil
}

func (p *Parser) ParseSequence() (*Sequence, error) {
	seq := &Sequence{}

	if p.current.Type == TokenEOF {
		return seq, nil
	}

	var currentOp Operator = OpNone

	for p.current.Type != TokenEOF && p.current.Type != TokenSemicolon && p.current.Type != TokenFi && p.current.Type != TokenDone && p.current.Type != TokenElif && p.current.Type != TokenElse {
		if p.current.Type == TokenIncomplete {
			return seq, ErrIncomplete
		}

		pipeline, err := p.parsePipeline()
		if err != nil {
			return seq, err
		}
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

	return seq, nil
}

func (p *Parser) parsePipeline() (*Pipeline, error) {
	pipeline := &Pipeline{}

	if p.current.Type == TokenEOF {
		return pipeline, nil
	}

	for {
		if p.current.Type == TokenIncomplete {
			return pipeline, ErrIncomplete
		}

		cmd, err := p.parseCommand()
		if err != nil {
			return pipeline, err
		}
		if cmd != nil {
			pipeline.Commands = append(pipeline.Commands, cmd)
		}

		if p.current.Type == TokenPipe {
			p.advance()
			continue
		}

		break
	}

	return pipeline, nil
}

func (p *Parser) parseArg(t Token) Arg {
	arg := Arg{Value: t.Value, IsGlobbable: t.IsGlobbable}
	if strings.HasPrefix(t.Value, "$") || strings.HasPrefix(t.Value, "#") {
		arg.IsVarRef = true
		arg.VarType = string(t.Value[0])
		name := t.Value[1:]
		
		// Check for array access e.g., $arr[0]
		if idx := strings.Index(name, "["); idx != -1 {
			if strings.HasSuffix(name, "]") {
				arg.IsArrayIdx = true
				arg.ArrayIndex = name[idx+1 : len(name)-1]
				arg.VarName = name[:idx]
			} else {
				arg.VarName = name
			}
		} else {
			arg.VarName = name
		}
	}
	return arg
}

func (p *Parser) parseCommand() (*Command, error) {
	if p.current.Type == TokenEOF || p.current.Type == TokenPipe || p.current.Type == TokenAnd || p.current.Type == TokenOr || p.current.Type == TokenRParen || p.current.Type == TokenSemicolon || p.current.Type == TokenFi || p.current.Type == TokenDone || p.current.Type == TokenElif || p.current.Type == TokenElse {
		return nil, nil
	}

	cmd := &Command{}

	if p.current.Type == TokenLParen {
		p.advance()
		cmd.IsSubshell = true

		var subshellTokens []string
		parenCount := 1

		for p.current.Type != TokenEOF {
			if p.current.Type == TokenIncomplete {
				return nil, ErrIncomplete
			}
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

		if parenCount > 0 && p.current.Type == TokenEOF {
			return nil, ErrIncomplete
		}

		cmd.SubshellString = strings.Join(subshellTokens, " ")
	}

	for p.current.Type != TokenEOF && p.current.Type != TokenPipe && p.current.Type != TokenAnd && p.current.Type != TokenOr && p.current.Type != TokenRParen && p.current.Type != TokenSemicolon && p.current.Type != TokenFi && p.current.Type != TokenDone && p.current.Type != TokenElif && p.current.Type != TokenElse {
		if p.current.Type == TokenIncomplete {
			return nil, ErrIncomplete
		}
		
		switch p.current.Type {
		case TokenWord:
			if !cmd.IsSubshell {
				cmd.Args = append(cmd.Args, p.parseArg(p.current))
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
		return cmd, nil
	}
	return nil, nil
}
