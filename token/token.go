package token

// TokenType defines what type a Token is
type TokenType string

// Token defines an AST toklen
type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT = "IDENT"
	INT   = "INT"

	ASSIGN  = "="
	PLUS    = "+"
	MINUS   = "-"
	ASTERIX = "*"
	FSLASH  = "/"

	BANG = "!"

	EQ    = "=="
	NOTEQ = "!="

	GT  = "GT"
	LT  = "LT"
	GTE = "GTE"
	LTE = "LTE"

	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	STRING   = "STRING"
	USE      = "USE"
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"return": RETURN,
	"if":     IF,
	"else":   ELSE,
	"use":    USE,
}

// LookupIdent checks if an identifier is a keyword or a user identifier
func LookupIdent(ident string) TokenType {
	if tok, found := keywords[ident]; found {
		return tok
	}
	return IDENT
}
