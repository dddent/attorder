package lexxer

import (
	"log"
	"strings"
	"testing"
)

func TestLexxer(t *testing.T) {
	l := New(strings.NewReader("<a [href]=\"test\" (click)=\"clickEvent()\" class=\"test asd\"></a>"))
	ts, err := l.All()
	if err != nil {
		t.Fatalf("Error getting tokens: %s", err)
	}
	exp := []TokType{
		TokST,
		TokIdent,
		TokWhitespace,
		TokLbracket,
		TokIdent,
		TokRbracket,
		TokEquals,
		TokString,
		TokWhitespace,
		TokLparen,
		TokIdent,
		TokRparen,
		TokEquals,
		TokString,
		TokWhitespace,
		TokIdent,
		TokEquals,
		TokString,
		TokGT,
		TokST,
		TokSlash,
		TokIdent,
		TokGT,
		TokEOF,
	}
	log.Println(ts)
	checkTokens(exp, ts, t)
}

func TestComment(t *testing.T) {
	l := New(strings.NewReader("<div></div>\n<!-- <div> -->"))
	ts, err := l.All()
	if err != nil {
		t.Fatalf("Error getting tokens: %s", err)
	}
	exp := []TokType{
		TokST,
		TokIdent,
		TokGT,
		TokST,
		TokSlash,
		TokIdent,
		TokGT,
		TokWhitespace,
		TokCmtStart,
		TokWhitespace,
		TokST,
		TokIdent,
		TokGT,
		TokWhitespace,
		TokCmtEnd,
		TokEOF,
	}
	checkTokens(exp, ts, t)
}

func checkTokens(exp []TokType, got []Token, t *testing.T) {
	if len(got) != len(exp) {
		t.Fatalf("Expected %d tokens but got %d", len(exp), len(got))
	}

	for i := range got {
		if got[i].Type() != exp[i] {
			t.Fatalf("Expected %s but got %s in pos %d", exp[i], got[i], i)
		}
	}
}
