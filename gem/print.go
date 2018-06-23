package gem

import (
	"git.fractalqb.de/fractalqb/xsx"
)

func Print(pr xsx.Printer, xpr Expr) (err error) {
	switch expr := xpr.(type) {
	case *Atom:
		if expr.Quoted() {
			err = pr.Atom(expr.Str, expr.Meta(), xsx.Qforce)
		} else {
			err = pr.Atom(expr.Str, expr.Meta(), xsx.Qcond)
		}
	case *Sequence:
		switch expr.Brace() {
		case Paren:
			pr.Begin('(', expr.Meta())
		case Square:
			pr.Begin('[', expr.Meta())
		case Curly:
			pr.Begin('{', expr.Meta())
		default:
			pr.Begin('(', expr.Meta()) // be forgiving
		}
		for _, sub := range expr.Elems {
			Print(pr, sub)
		}
		pr.End()
	}
	return err
}
