package gem

import (
	"fmt"

	"git.fractalqb.de/fractalqb/xsx"
)

func ReadCurrent(p *xsx.PullParser) (Expr, error) {
	switch p.LastToken() {
	case xsx.TokEOI:
		return nil, xsx.PullEOI
	case xsx.TokAtom:
		a := &Atom{Str: p.Atom}
		a.SetMeta(p.WasMeta())
		a.SetQuoted(p.WasQuot)
		return a, nil
	case xsx.TokBegin:
		return readSeq(p)
	default:
		return nil,
			fmt.Errorf("gem read current: unexpected token '%s'", p.LastToken())
	}
}

func ReadNext(p *xsx.PullParser) (Expr, error) {
	if tok, _ := p.Next(); tok != xsx.TokEOI {
		return ReadCurrent(p)
	} else {
		return nil, xsx.PullEOI
	}
}

func readSeq(p *xsx.PullParser) (Expr, error) {
	res := &Sequence{}
	res.SetMeta(p.WasMeta())
	switch p.LastBrace() {
	case '(':
		res.SetBrace(Paren)
	case '[':
		res.SetBrace(Square)
	case '{':
		res.SetBrace(Curly)
	default:
		panic(fmt.Sprintf("gem pull: illegal opening brace '%c'", p.LastBrace()))
	}
	for tok, err := p.Next(); tok != xsx.TokEnd; tok, err = p.Next() {
		if err != nil {
			return res, err
		}
		elm, err := ReadCurrent(p)
		if err != nil {
			return res, err
		}
		res.Elems = append(res.Elems, elm)
	}
	return res, nil
}
