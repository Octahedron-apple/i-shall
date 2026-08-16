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
	TokenCommandSub     // $(...)
	TokenIncomplete     // Returned when EOF is hit inside quotes or $()
)

type Token struct {
	Type           TokenType
	Value          string
	IsGlobbable    bool
	IsDoubleQuoted bool
	IsSingleQuoted bool
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
		// Check for $( — handled in readWord as a command sub when part of a word
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
	case '$':
		// Could be $(...) command substitution
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '(' {
			return l.readCommandSub()
		}
	}

	return l.readWord()
}

func (l *Lexer) readCommandSub() Token {
	l.pos += 2 // consume $(
	var body strings.Builder
	parenCount := 1
	for l.pos < len(l.input) {
		c := l.input[l.pos]
		if c == '\\' {
			body.WriteByte(c)
			l.pos++
			if l.pos < len(l.input) {
				body.WriteByte(l.input[l.pos])
				l.pos++
			}
			continue
		}
		if c == '(' {
			parenCount++
		} else if c == ')' {
			parenCount--
			if parenCount == 0 {
				l.pos++
				return Token{Type: TokenCommandSub, Value: body.String()}
			}
		}
		body.WriteByte(c)
		l.pos++
	}
	return Token{Type: TokenIncomplete, Value: "command-sub"}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && (l.input[l.pos] == ' ' || l.input[l.pos] == '\t' || l.input[l.pos] == '\r') {
		l.pos++
	}
}

func (l *Lexer) readWord() Token {
	var word strings.Builder
	var inSingleQuote bool
	var inDoubleQuote bool
	var isGlobbable bool
	var isDoubleQuotedWord bool
	var isSingleQuotedWord bool
	var inBrace int // depth counter for { }

	for {
		if l.pos >= len(l.input) {
			if inSingleQuote || inDoubleQuote {
				return Token{Type: TokenIncomplete, Value: "quote"}
			}
			break
		}

		char := l.input[l.pos]

		// Escape sequences
		if char == '\\' {
			l.pos++
			if l.pos < len(l.input) {
				word.WriteByte(l.input[l.pos])
				l.pos++
			}
			continue
		}

		// $(...) inside a word — break so NextToken picks it up separately
		if char == '$' && !inSingleQuote && l.pos+1 < len(l.input) && l.input[l.pos+1] == '(' {
			if word.Len() > 0 {
				break // return current word first; next call will return the command sub
			}
			return l.readCommandSub()
		}

		if char == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			isSingleQuotedWord = true
			l.pos++
			continue
		}

		if char == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			isDoubleQuotedWord = true
			l.pos++
			continue
		}

		// Track brace depth so commas inside {} don't split the token
		if !inSingleQuote && !inDoubleQuote {
			if char == '{' {
				inBrace++
			} else if char == '}' {
				if inBrace > 0 {
					inBrace--
				}
			}
		}

		// Word-boundary characters (only outside quotes/braces)
		if !inSingleQuote && !inDoubleQuote && inBrace == 0 {
			switch char {
			case ' ', '\t', '\n', '\r', '|', '<', '>', '&', '(', ')', ';', '=', ',':
				if word.Len() > 0 {
					goto done
				}
				// Skip leading whitespace (shouldn't normally happen here)
				if char == ' ' || char == '\t' || char == '\r' {
					l.pos++
					continue
				}
				goto done
			case '*', '?':
				isGlobbable = true
			}
		}

		word.WriteByte(char)
		l.pos++
	}

done:
	val := word.String()
	if val == "" {
		return Token{Type: TokenEOF}
	}

	// Return quoted tokens early — skip keyword matching
	if isDoubleQuotedWord && !isSingleQuotedWord {
		return Token{Type: TokenWord, Value: val, IsDoubleQuoted: true}
	}
	if isSingleQuotedWord && !isDoubleQuotedWord {
		return Token{Type: TokenWord, Value: val, IsSingleQuoted: true}
	}

	// Keyword matching (only for plain unquoted words)
	if !isGlobbable {
		switch val {
		case "if":
			return Token{Type: TokenIf, Value: val}
		case "elif":
			return Token{Type: TokenElif, Value: val}
		case "else":
			return Token{Type: TokenElse, Value: val}
		case "fi":
			return Token{Type: TokenFi, Value: val}
		case "while":
			return Token{Type: TokenWhile, Value: val}
		case "done":
			return Token{Type: TokenDone, Value: val}
		case "for":
			return Token{Type: TokenFor, Value: val}
		case "export":
			return Token{Type: TokenExport, Value: val}
		case "fn":
			return Token{Type: TokenFn, Value: val}
		case "end":
			return Token{Type: TokenEnd, Value: val}
		case "alias":
			return Token{Type: TokenAlias, Value: val}
		}
	}

	return Token{Type: TokenWord, Value: val, IsGlobbable: isGlobbable}
}
