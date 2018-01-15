package gem

type State struct {
	Results []Expr
	ctx     []*Sequence
}

func (pst *State) Begin(scanPos uint64, isMeta bool, brace rune) error {
	s := &Sequence{}
	s.SetMeta(isMeta)
	s.SetBrace(FromRune(brace))
	if len(pst.ctx) > 0 {
		c := pst.ctx[len(pst.ctx)-1]
		c.Elems = append(c.Elems, s)
	}
	pst.ctx = append(pst.ctx, s)
	return nil
}

func (pst *State) End(scanPos uint64, brace rune) error {
	lm1 := len(pst.ctx) - 1
	s := pst.ctx[lm1]
	pst.ctx = pst.ctx[:lm1]
	if len(pst.ctx) == 0 {
		pst.Results = append(pst.Results, s)
	}
	return nil
}

func (pst *State) Atom(scanPos uint64, isMeta bool, atom string, quoted bool) error {
	a := &Atom{Str: atom}
	a.SetMeta(isMeta)
	a.SetQuoted(quoted)
	if len(pst.ctx) == 0 {
		pst.Results = append(pst.Results, a)
	} else {
		s := pst.ctx[len(pst.ctx)-1]
		s.Elems = append(s.Elems, a)
	}
	return nil
}
