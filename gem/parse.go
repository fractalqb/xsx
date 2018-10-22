package gem

import (
	"strings"
)

type State struct {
	Results []Expr
	ctx     []*Sequence
}

func (pst *State) Begin(isMeta bool, brace byte) {
	s := &Sequence{}
	s.SetMeta(isMeta)
	s.SetBrace(FromRune(brace))
	if len(pst.ctx) > 0 {
		c := pst.ctx[len(pst.ctx)-1]
		c.Elems = append(c.Elems, s)
	}
	pst.ctx = append(pst.ctx, s)
}

func (pst *State) End(isMeta bool, brace byte) {
	lm1 := len(pst.ctx) - 1
	s := pst.ctx[lm1]
	pst.ctx = pst.ctx[:lm1]
	if len(pst.ctx) == 0 {
		pst.Results = append(pst.Results, s)
	}
}

func (pst *State) Atom(isMeta bool, atom []byte, quoted bool) {
	var sb strings.Builder
	sb.Write(atom)
	a := &Atom{Str: sb.String()}
	a.SetMeta(isMeta)
	a.SetQuoted(quoted)
	if len(pst.ctx) == 0 {
		pst.Results = append(pst.Results, a)
	} else {
		s := pst.ctx[len(pst.ctx)-1]
		s.Elems = append(s.Elems, a)
	}
}
