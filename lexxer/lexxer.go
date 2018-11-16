package lexxer

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
)

type TokType int

const (
	TokUnknown TokType = iota
	TokST
	TokGT
	TokIdent
	TokEquals
	TokLparen
	TokRparen
	TokLbracket
	TokRbracket
	TokQuote
	TokDquote
	TokSlash
	TokBang
	TokWhitespace
	TokMinus
	TokHash
	TokAt
	TokEOF
	TokStar
	TokBangTag
	TokCmtStart
	TokCmtEnd
	TokColon
	TokDot
)

type Token struct {
	t TokType
	v string
	l int
	c int
}

type Lexxer struct {
	r       *bufio.Reader
	curr    rune
	peek    rune
	eof     bool
	peekEOF bool
	l       int
	c       int
}

func (t TokType) String() string {
	switch t {
	case TokST:
		return "TokST"
	case TokColon:
		return "TokColon"
	case TokDot:
		return "TokDot"
	case TokGT:
		return "TokGT"
	case TokIdent:
		return "TokIdent"
	case TokEquals:
		return "TokEquals"
	case TokLparen:
		return "TokLparen"
	case TokQuote:
		return "TokQuote"
	case TokDquote:
		return "TokDquote"
	case TokRparen:
		return "TokRparen"
	case TokLbracket:
		return "TokLbracket"
	case TokRbracket:
		return "TokRbracket"
	case TokSlash:
		return "TokSlash"
	case TokBang:
		return "TokBang"
	case TokMinus:
		return "TokMinus"
	case TokHash:
		return "TokHash"
	case TokAt:
		return "TokAt"
	case TokWhitespace:
		return "TokWhitespace"
	case TokStar:
		return "TokStar"
	case TokBangTag:
		return "TokBangTag"
	case TokUnknown:
		return "TokUnknown"
	case TokCmtStart:
		return "TokCmtStart"
	case TokCmtEnd:
		return "TokCmtEnd"
	case TokEOF:
		return "TokEOF"
	default:
		return fmt.Sprintf("TokType(%d)", t)
	}
}

func (t *Token) Type() TokType {
	return t.t
}

func (t *Token) Value() string {
	return t.v
}

func (t Token) String() string {
	return fmt.Sprintf("[%d:%d](%s \"%s\")", t.l, t.c, t.Type(), t.Value())
}

func New(r io.Reader) *Lexxer {
	res := &Lexxer{bufio.NewReader(r), 0, 0, false, false, 0, 0}
	res.advance()
	res.advance()
	return res
}

func (l *Lexxer) Next() (Token, error) {
	if l.eof {
		return Token{TokEOF, "", l.l, l.c}, nil
	}

	if unicode.IsSpace(l.curr) {
		var s string
		for ; unicode.IsSpace(l.curr); l.advance() {
			s += string(l.curr)
		}
		return l.newToken(TokWhitespace, s), nil
	}

	switch l.curr {
	case '<':
		if l.peek == '!' {
			l.advance()
			l.advance()
			if l.curr == '-' && l.peek == '-' {
				l.advance()
				l.advance()
				return Token{TokCmtStart, "<!--", l.l, l.c}, nil
			}
			return Token{TokBangTag, "<!", l.l, l.c}, nil
		}
		return l.fromCurr(TokST)
	case '>':
		return l.fromCurr(TokGT)
	case '"':
		return l.fromCurr(TokDquote)
	case '\'':
		return l.fromCurr(TokQuote)
	case ':':
		return l.fromCurr(TokColon)
	case '.':
		return l.fromCurr(TokDot)
	case '=':
		return l.fromCurr(TokEquals)
	case '(':
		return l.fromCurr(TokLparen)
	case ')':
		return l.fromCurr(TokRparen)
	case '[':
		return l.fromCurr(TokLbracket)
	case ']':
		return l.fromCurr(TokRbracket)
	case '/':
		return l.fromCurr(TokSlash)
	case '!':
		return l.fromCurr(TokBang)
	case '-':
		l.advance()
		if l.curr == '-' && l.peek == '>' {
			l.advance()
			l.advance()
			return Token{TokCmtEnd, "-->", l.l, l.c}, nil
		}
		return Token{TokMinus, "-", l.l, l.c}, nil
	case '#':
		return l.fromCurr(TokHash)
	case '@':
		return l.fromCurr(TokAt)
	case '*':
		return l.fromCurr(TokStar)
	}

	if unicode.IsLetter(l.curr) {
		var s string
		for ; unicode.IsLetter(l.curr) || unicode.IsDigit(l.curr); l.advance() {
			s += string(l.curr)
		}
		return l.newToken(TokIdent, s), nil
	}

	t := l.curr
	l.advance()
	return Token{TokUnknown, string(t), l.l, l.c}, nil
}

func (l *Lexxer) advance() error {
	if l.peekEOF {
		l.eof = true
	}
	r, _, err := l.r.ReadRune()
	if err == io.EOF {
		l.peekEOF = true
	} else if err != nil {
		return err
	}
	if r == '\n' {
		l.l++
		l.c = 0
	} else {
		l.c++
	}
	l.curr = l.peek
	l.peek = r
	return nil
}

func (l *Lexxer) newToken(t TokType, v string) Token {
	return Token{t, v, l.l, l.c}
}

func (l *Lexxer) fromCurr(t TokType) (Token, error) {
	res := l.newToken(t, string(l.curr))
	if err := l.advance(); err != nil {
		return Token{}, err
	}
	return res, nil
}

func (l *Lexxer) All() ([]Token, error) {
	res := make([]Token, 0)
	for t, err := l.Next(); t.Type() != TokEOF; t, err = l.Next() {
		if err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	res = append(res, Token{TokEOF, "", l.l, l.c})
	return res, nil
}
