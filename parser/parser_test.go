package parser

import (
	"strings"
	"testing"

	"github.com/dddent/attorder/lexxer"
)

func TestParser(t *testing.T) {
	l := lexxer.New(strings.NewReader("<a [href]=\"test\" (click)=\"clickEvent()\" class=\"test asd\"></a>"))
	p := New(l)
	ast, err := p.ParseAST()
	if err != nil {
		t.Fatalf("Error parsing ast: %s", err)
	}
	if len(ast.Nodes()) != 2 {
		t.Errorf("Expected 2 nodes but got %d", len(ast.Nodes()))
	}
	if tn, ok := ast.Nodes()[0].(*TagNode); !ok {
		t.Errorf("Expected tag node but got %v", ast.Nodes()[0])
	} else if tn.Ident() != "a" {
		t.Errorf("Expected identifier 'a' but got %s", tn.Ident())
	} else if len(tn.Attrs()) != 3 {
		t.Errorf("Expected 3 attributes, but got %d", len(tn.Attrs()))
	} else {
		expectAttribute(t, tn.Attrs()[0], "class", "\"test asd\"", attrPure)
		expectAttribute(t, tn.Attrs()[1], "click", "\"clickEvent()\"", attrParen)
		expectAttribute(t, tn.Attrs()[2], "href", "\"test\"", attrBracket)
	}

	if cn, ok := ast.Nodes()[1].(*CTagNode); !ok {
		t.Errorf("Expected CTag but got %v", ast.Nodes()[1])
	} else if cn.Ident() != "a" {
		t.Errorf("Expected CTag ident to be 'a' but got %s", cn.Ident())
	}
}

func TestComment(t *testing.T) {
	l := lexxer.New(strings.NewReader("<!-- <div> -->"))
	p := New(l)
	ast, err := p.ParseAST()
	if err != nil {
		t.Fatalf("Error parsing ast: %s", err)
	}
	if len(ast.Nodes()) != 1 {
		t.Errorf("Expected 1 node but got %d", len(ast.Nodes()))
	}
	if _, ok := ast.Nodes()[0].(*CommentTagNode); !ok {
		t.Errorf("Expected comment node but got %v", ast.Nodes()[0])
	}
}

func expectAttribute(t *testing.T, attr *Attribute, ident string, val string, tp attrType) {
	if attr.Ident() != ident {
		t.Errorf("Expected attribute ident to be %s but got %s", ident, attr.Ident())
	} else if attr.Val() != val {
		t.Errorf("Expected attribute val to be %s but got %s", val, attr.Val())
	} else if attr.Type() != tp {
		t.Errorf("Expected attribute type to be %s but got %s", tp, attr.Type())
	}
}
