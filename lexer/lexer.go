package lexer

import "github.com/gilmae/monkey/token"

// Lexer represents the internal state of a lexer parsing through input
type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
}

// New initialises a Lexer
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) Line() int {
	return l.line
}

// NextToken reads the next token from the input
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()
	switch l.ch {
	case '#':
		// Comment, skip
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
		return (l.NextToken())
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			lit := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: lit}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}

	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '*':
		tok = newToken(token.ASTERIX, l.ch)
	case '/':
		tok = newToken(token.FSLASH, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			lit := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOTEQ, Literal: lit}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			lit := string(ch) + string(l.ch)
			tok = token.Token{Type: token.LTE, Literal: lit}
		} else {
			tok = newToken(token.LT, l.ch)
		}

	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			lit := string(ch) + string(l.ch)
			tok = token.Token{Type: token.GTE, Literal: lit}
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		}
		if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		}
		tok = newToken(token.ILLEGAL, l.ch)

	}

	l.readChar()
	return tok
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	if l.ch == '\n' {
		l.line += 1
	}
	l.readPosition++
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]

}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}

	str := l.input[position:l.position]
	if len(str) == 0 {
		return str
	}
	ret := ""
	position = 0
	for {
		ch := str[position]
		if ch == '\\' {
			switch str[position+1] {
			case 't':
				ret = ret + string('\t')
				position += 1
			case 'n':
				ret = ret + string('\n')
				position += 1
			case '\\':
				ret = ret + string('\\')
				position += 1
			default:
				ret = ret + string('\\')
			}
		} else {
			ret = ret + string(ch)
		}
		position = position + 1
		if position >= len(str) {
			break
		}

	}

	return ret
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}
