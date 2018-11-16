package parser

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/dddent/attorder/lexxer"
)

type attrType int

const (
	attrPure attrType = iota
	attrBracket
	attrParen
	attrHash
	attrAt
	attrStar
	attrTwoway
)

func (t attrType) String() string {
	switch t {
	case attrPure:
		return "attrPure"
	case attrBracket:
		return "attrBracket"
	case attrParen:
		return "attrParen"
	case attrTwoway:
		return "attrTwoway"
	default:
		return "UNKNOWN_ATTR_TYPE"
	}
}

type TextNode struct {
	content string
}

func (n *TextNode) HTMLString() string {
	return n.content
}

type Attribute struct {
	ident  string
	val    string
	t      attrType
	hasVal bool
}

func (a *Attribute) Ident() string {
	return a.ident
}

func (a *Attribute) Val() string {
	return a.val
}

func (a *Attribute) Type() attrType {
	return a.t
}

func (a *Attribute) HTMLString() string {
	var ident string
	switch a.Type() {
	case attrPure:
		ident = a.ident
	case attrBracket:
		ident = "[" + a.ident + "]"
	case attrParen:
		ident = "(" + a.ident + ")"
	case attrHash:
		ident = "#" + a.ident
	case attrAt:
		ident = "@" + a.ident
	case attrStar:
		ident = "*" + a.ident
	case attrTwoway:
		ident = "[(" + a.ident + ")]"
	}
	if a.hasVal {
		return fmt.Sprintf("%s=%s", ident, a.val)
	} else {
		return ident
	}
}

type Node interface {
	HTMLString() string
}

type BangTagNode struct {
	content string
}

func (n *BangTagNode) HTMLString() string {
	return fmt.Sprintf("<!%s>", n.content)
}

type CommentTagNode struct {
	content string
}

func (n *CommentTagNode) HTMLString() string {
	return fmt.Sprintf("<!--%s-->", n.content)
}

type TagNode struct {
	ident string
	attrs []*Attribute
	ws    []string
}

func (p *Parser) sortAttrs(attrs []*Attribute) error {
	idents := make([]string, len(attrs))
	ordered := make([]string, 0, len(attrs))
	m := make(map[string]*Attribute)
	for i, a := range attrs {
		idents[i] = a.ident
		m[a.ident] = a
	}
	for i, id := range idents {
		for _, a := range p.settings.AttrOrder {
			if match, err := regexp.MatchString("^"+a+"$", id); err != nil {
				return err
			} else if match {
				ordered = append(ordered, id)
				if len(idents) > i+1 {
					copy(idents[i:], idents[i+1:])
				}
				idents = idents[:len(idents)-1]
				break
			}
		}
	}
	sort.Strings(idents)
	ordered = append(ordered, idents...)
	for i, id := range ordered {
		attrs[i] = m[id]
	}
	return nil
}

func (n *TagNode) Ident() string {
	return n.ident
}

func (n *TagNode) Attrs() []*Attribute {
	return n.attrs
}

type CTagNode struct {
	ident string
}

func (n *CTagNode) Ident() string {
	return n.ident
}

type AST struct {
	nodes []Node
}

func (a *AST) HTMLString() string {
	var res string
	for _, n := range a.nodes {
		res += n.HTMLString()
	}
	return res
}

func (a *AST) Nodes() []Node {
	return a.nodes
}

type Settings struct {
	AttrOrder []string
	attrMap   map[string]struct{}
}

func (s *Settings) init() {
	s.attrMap = make(map[string]struct{})
	for _, a := range s.AttrOrder {
		s.attrMap[a] = struct{}{}
	}
}

type Parser struct {
	l        *lexxer.Lexxer
	curr     lexxer.Token
	peek     lexxer.Token
	settings Settings
}

func New(l *lexxer.Lexxer, settings Settings) (*Parser, error) {
	res := &Parser{l: l, settings: settings}
	res.settings.init()
	if err := res.step(false); err != nil {
		return nil, err
	}
	if err := res.step(false); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Parser) ParseAST() (*AST, error) {
	res := &AST{make([]Node, 0)}
	for !p.isCurr(lexxer.TokEOF) {
		n, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		res.nodes = append(res.nodes, n)
	}
	return res, nil
}

func (p *Parser) step(skipWs bool) error {
	peek, err := p.l.Next()
	if err != nil {
		return err
	}
	p.curr = p.peek
	p.peek = peek
	if skipWs {
		p.skipWS()
	}
	return nil
}

func (p *Parser) isCurr(t lexxer.TokType) bool {
	return p.curr.Type() == t
}

func (p *Parser) isPeek(t lexxer.TokType) bool {
	return p.peek.Type() == t
}

func (p *Parser) expect(desc string, t lexxer.TokType) (lexxer.Token, error) {
	if !p.isCurr(t) {
		return lexxer.Token{}, fmt.Errorf("Expected %s (%s) but got %s", desc, t, p.curr)
	}
	c := p.curr
	if err := p.step(false); err != nil {
		return lexxer.Token{}, err
	}
	return c, nil
}

func (p *Parser) parseTextNode() (Node, error) {
	res := &TextNode{}
	for !p.isCurr(lexxer.TokCmtStart) && !p.isCurr(lexxer.TokST) && !p.isCurr(lexxer.TokEOF) {
		res.content += p.curr.Value()
		if err := p.step(false); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (p *Parser) parseNode() (Node, error) {
	switch p.curr.Type() {
	case lexxer.TokST:
		return p.parseTagNode()
	case lexxer.TokBangTag:
		return p.parseBangTagNode()
	case lexxer.TokCmtStart:
		return p.parseCommentNode()
	default:
		return p.parseTextNode()
	}
	return nil, fmt.Errorf("Expected node but got %s", p.curr)
}

func (n *TagNode) HTMLString() string {
	res := fmt.Sprintf("<%s", n.ident)
	for i, a := range n.attrs {
		res += n.ws[i] + a.HTMLString()
	}
	res += ">"
	return res
}

func (p *Parser) parseCommentNode() (Node, error) {
	if _, err := p.expect("Comment start ('<!--')", lexxer.TokCmtStart); err != nil {
		return nil, err
	}
	res := &CommentTagNode{}
	for !p.isCurr(lexxer.TokEOF) && !p.isCurr(lexxer.TokCmtEnd) {
		res.content += p.curr.Value()
		err := p.step(false)
		if err != nil {
			return nil, err
		}
	}
	if _, err := p.expect("Comment end", lexxer.TokCmtEnd); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Parser) parseTagName() (string, error) {
	var name string
	for p.isCurr(lexxer.TokIdent) || p.isCurr(lexxer.TokMinus) {
		name += p.curr.Value()
		if err := p.step(false); err != nil {
			return "", err
		}
	}
	return name, nil
}

func (p *Parser) parseTagNode() (Node, error) {
	if !p.isCurr(lexxer.TokST) {
		return nil, expectedError(lexxer.TokST, p.curr)
	}
	if p.isPeek(lexxer.TokSlash) {
		return p.parseCTagNode()
	}
	if err := p.step(true); err != nil {
		return nil, err
	}
	name, err := p.parseTagName()
	if err != nil {
		return nil, err
	}
	res := &TagNode{ident: name, attrs: make([]*Attribute, 0), ws: make([]string, 0)}
	for !p.isCurr(lexxer.TokGT) && !p.isCurr(lexxer.TokEOF) {
		ws, err := p.skipWS()
		if err != nil {
			return nil, err
		}
		res.ws = append(res.ws, ws)
		if p.isCurr(lexxer.TokSlash) {
			if err := p.step(false); err != nil {
				return nil, err
			}
			break
		}
		if p.isCurr(lexxer.TokGT) {
			break
		}
		a, err := p.parseAttribute()
		if err != nil {
			return nil, err
		}
		res.attrs = append(res.attrs, a)
	}
	if _, err := p.expect("Closing '>'", lexxer.TokGT); err != nil {
		return nil, err
	}
	if err := p.sortAttrs(res.attrs); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Parser) parseBangTagNode() (Node, error) {
	if _, err := p.expect("Bang Tag ('<!')", lexxer.TokBangTag); err != nil {
		return nil, err
	}
	var content string
	for !p.isCurr(lexxer.TokGT) && !p.isCurr(lexxer.TokEOF) {
		content += p.curr.Value()
		if err := p.step(false); err != nil {
			return nil, err
		}
	}

	if _, err := p.expect("Bang Tag Closing '>'", lexxer.TokGT); err != nil {
		return nil, err
	}
	return &BangTagNode{content}, nil
}

func (p *Parser) parseAttribute() (*Attribute, error) {
	var t attrType
	switch p.curr.Type() {
	case lexxer.TokLparen:
		t = attrParen
	case lexxer.TokLbracket:
		if p.isPeek(lexxer.TokLparen) {
			t = attrTwoway
			if err := p.step(false); err != nil {
				return nil, err
			}
		} else {
			t = attrBracket
		}
	case lexxer.TokHash:
		t = attrHash
	case lexxer.TokIdent:
		t = attrPure
	case lexxer.TokAt:
		t = attrAt
	case lexxer.TokStar:
		t = attrStar
	default:
		return nil, expectedError("attribute identifier", p.curr)
	}
	if t != attrPure {
		if err := p.step(false); err != nil {
			return nil, err
		}
	}
	var name string
	for p.isCurr(lexxer.TokIdent) || p.isCurr(lexxer.TokMinus) || p.isCurr(lexxer.TokColon) || p.isCurr(lexxer.TokDot) || p.isCurr(lexxer.TokAt) {
		name += p.curr.Value()
		if err := p.step(false); err != nil {
			return nil, err
		}
	}
	if t == attrParen || t == attrTwoway {
		if _, err := p.expect("Attribute ')'", lexxer.TokRparen); err != nil {
			return nil, err
		}
	}
	if t == attrBracket || t == attrTwoway {
		if _, err := p.expect("Attribute ']'", lexxer.TokRbracket); err != nil {
			return nil, err
		}
	}
	if !p.isCurr(lexxer.TokEquals) {
		return &Attribute{name, "", t, false}, nil
	}
	if _, err := p.expect("Attribute '='", lexxer.TokEquals); err != nil {
		return nil, err
	}
	var s string
	if p.isCurr(lexxer.TokIdent) {
		s = p.curr.Value()
	} else {
		str, err := p.parseString()
		if err != nil {
			return nil, err
		}
		s = str
	}
	return &Attribute{name, s, t, true}, nil
}

func (p *Parser) parseString() (string, error) {
	if !p.isCurr(lexxer.TokQuote) && !p.isCurr(lexxer.TokDquote) {
		_, err := p.expect("String start ('\"' or ''')", lexxer.TokQuote)
		return "", err
	}
	delimType := p.curr.Type()
	res := p.curr.Value()
	if err := p.step(false); err != nil {
		return "", err
	}
	for !p.isCurr(delimType) && !p.isCurr(lexxer.TokEOF) {
		res += p.curr.Value()
		if err := p.step(false); err != nil {
			return "", err
		}
	}
	t, err := p.expect("String end", delimType)
	if err != nil {
		return "", err
	}
	res += t.Value()
	return res, nil
}

func (n *CTagNode) HTMLString() string {
	return fmt.Sprintf("</%s>", n.ident)
}

func (p *Parser) parseCTagNode() (Node, error) {
	if _, err := p.expect("Tag opening '<'", lexxer.TokST); err != nil {
		return nil, err
	}
	if _, err := p.expect("Closing Tag '/'", lexxer.TokSlash); err != nil {
		return nil, err
	}
	name, err := p.parseTagName()
	if err != nil {
		return nil, err
	}
	res := &CTagNode{name}
	if _, err := p.expect("Closing Tag '>'", lexxer.TokGT); err != nil {
		return nil, err
	}
	return res, nil
}

func expectedError(exp interface{}, got interface{}) error {
	return fmt.Errorf("Expected %s but got %s", exp, got)
}

func (p *Parser) skipWS() (string, error) {
	var res string
	if p.isCurr(lexxer.TokWhitespace) {
		res += p.curr.Value()
		if err := p.step(true); err != nil {
			return "", err
		}
	}
	return res, nil
}
