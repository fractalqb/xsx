package xsx

import (
	"io"
)

type Printer interface {
	Begin(bracket rune, meta bool) error
	End() error
	Atom(atom string, meta bool, quote QuoteMode) error
	Newline(count int, indent int) error
}

func closingRune(open rune) rune {
	return rune(closing(byte(open)))
}

func printAtom(wr io.Writer, atom string, meta bool, quote QuoteMode) (err error) {
	if meta {
		_, err = wr.Write(metaAtom)
		if err != nil {
			return err
		}
	}
	switch quote {
	case Qcond:
		_, err = CondQuoteTo(atom, wr)
	case Qforce:
		err = QuoteTo(atom, wr)
	case QSUPPRESS:
		_, err = wr.Write([]byte(atom))
	default:
		_, err = CondQuoteTo(atom, wr)
	}
	return err
}
