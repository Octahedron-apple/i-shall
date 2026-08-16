package shell

import (
	"strings"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenWord
	TokenPipe           // |
	TokenRedirectOut    // >
	TokenRedirectAppend // >>
	TokenRedirectIn     // <
	TokenAnd            // &&
	TokenOr             // ||
	TokenLParen         // (
	TokenRParen         // )
	TokenSemicolon      // ;
	TokenIf
	TokenElif
	TokenElse
	TokenFi
	TokenWhile
	TokenDone
	TokenAssign         // =
	TokenComma          // ,
	TokenFor            // for
	TokenExport         // export
	TokenFn             // fn
	TokenEnd            // end
	TokenAlias          // alias
	TokenIncomplete     // Returned when EOF is hit inside quotes
)

type Token struct {
	Type        TokenType
	Value       string
	IsGlobbable bool
}

type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, pos: 0}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, Value: ""}
	}

	char := l.input[l.pos]

	switch char {
	case '/':
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '/' {
			for l.pos < len(l.input) && l.input[l.pos] != '\n' {
				l.pos++
			}
			return l.NextToken()
		}
		l.pos++
		return Token{Type: TokenWord, Value: "/"}
	case '&':
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '&' {
			l.pos += 2
			return Token{Type: TokenAnd, Value: "&&"}
		}
	case '|':
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '|' {
			l.pos += 2
			return Token{Type: TokenOr, Value: "||"}
		}
		l.pos++
		return Token{Type: TokenPipe, Value: "|"}
	case '<':
		l.pos++
		return Token{Type: TokenRedirectIn, Value: "<"}
	case '>':
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '>' {
			l.pos += 2
			return Token{Type: TokenRedirectAppend, Value: ">>"}
		}
		l.pos++
		return Token{Type: TokenRedirectOut, Value: ">"}
	case '(':
		l.pos++
		return Token{Type: TokenLParen, Value: "("}
	case ')':
		l.pos++
		return Token{Type: TokenRParen, Value: ")"}
	case ';', '\n':
		l.pos++
		return Token{Type: TokenSemicolon, Value: ";"}
	case '=':
		l.pos++
		return Token{Type: TokenAssign, Value: "="}
	case ',':
		l.pos++
		return Token{Type: TokenComma, Value: ","}
	}

	return l.readWord()
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && (l.input[l.pos] == ' ' || l.input[l.pos] == '\t' || l.input[l.pos] == '\r') {
		l.pos++
	}
}

func (l *Lexer) readWord() Token {
	var builder strings.Builder
	var inSingleQuote bool
	var inDoubleQuote bool
	var isGlobbable bool

	for {
		if l.pos >= len(l.input) {
			if inSingleQuote || inDoubleQuote {
				return Token{Type: TokenIncomplete, Value: "quote"}
			}
			break
		}

		char := l.input[l.pos]

		if char == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			l.pos++
			continue
		}

		if char == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			l.pos++
			continue
		}

		if !inSingleQuote && !inDoubleQuote {
			if char == ' ' || char == '\t' || char == '\n' || char == '\r' || char == '|' || char == '<' || char == '>' || char == '&' || char == '(' || char == ')' || char == ';' || char == '=' || char == ',' {
				break
			}
			if char == '*' || char == '?' {
				isGlobbable = true
			}
		}

		builder.WriteByte(char)
		l.pos++
	}

	val := builder.String()
	if !isGlobbable && !inSingleQuote && !inDoubleQuote {
		switch val {
		case "if":
			return Token{Type: TokenIf, Value: val, IsGlobbable: false}
		case "elif":
			return Token{Type: TokenElif, Value: val, IsGlobbable: false}
		case "else":
			return Token{Type: TokenElse, Value: val, IsGlobbable: false}
		case "fi":
			return Token{Type: TokenFi, Value: val, IsGlobbable: false}
		case "while":
			return Token{Type: TokenWhile, Value: val, IsGlobbable: false}
		case "done":
			return Token{Type: TokenDone, Value: val, IsGlobbable: false}
		case "for":
			return Token{Type: TokenFor, Value: val, IsGlobbable: false}
		case "export":
			return Token{Type: TokenExport, Value: val, IsGlobbable: false}
		case "fn":
			return Token{Type: TokenFn, Value: val, IsGlobbable: false}
		case "end":
			return Token{Type: TokenEnd, Value: val, IsGlobbable: false}
		case "alias":
			return Token{Type: TokenAlias, Value: val, IsGlobbable: false}
		}
	}

	return Token{Type: TokenWord, Value: val, IsGlobbable: isGlobbable}
}
