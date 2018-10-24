package xsx

import (
	"errors"
	"fmt"
	"io"
)

// PrettyPrinter currently is a quick'n'dirty impl that is about to change
type PrettyPrinter struct {
	Writer io.Writer
	ilvl   int
	istr   []byte
	ends   []rune
}

func Pretty(wr io.Writer, indent string) *PrettyPrinter {
	res := &PrettyPrinter{
		Writer: wr,
		istr:   []byte(indent),
	}
	return res
}

func (pp *PrettyPrinter) indent() (err error) {
	for i := pp.ilvl; i > 0; i-- {
		_, err = pp.Writer.Write(pp.istr)
		if err != nil {
			return err
		}
	}
	return err
}

func (pp *PrettyPrinter) Begin(bracket rune, meta bool) (err error) {
	err = pp.indent()
	if err != nil {
		return err
	}
	if meta {
		_, err = pp.Writer.Write(metaAtom)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(pp.Writer, "%c\n", bracket)
	pp.ilvl++
	pp.ends = append(pp.ends, closingRune(bracket))
	return err
}

func (pp *PrettyPrinter) End() (err error) {
	if len(pp.ends) == 0 {
		return errors.New("nothing to end")
	}
	end := pp.ends[len(pp.ends)-1]
	pp.ends = pp.ends[:len(pp.ends)-1]
	pp.ilvl--
	err = pp.indent()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(pp.Writer, "%c\n", end)
	return err
}

func (pp *PrettyPrinter) Atom(atom string, meta bool, quote QuoteMode) (err error) {
	err = pp.indent()
	if err != nil {
		return err
	}
	err = printAtom(pp.Writer, atom, meta, quote)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(pp.Writer)
	return err
}

func (p *PrettyPrinter) Newline(count int, indent int) error { return nil }
