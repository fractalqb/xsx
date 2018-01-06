package xsx

import (
	"errors"
	"fmt"
	"io"
)

type PrettyPrinter struct {
	Writer io.Writer
	Indent string
	nest   []rune
	sep    bool
}

func Pretty(wr io.Writer) *PrettyPrinter {
	res := &PrettyPrinter{Writer: wr, Indent: "  "}
	return res
}

func (p *PrettyPrinter) indent(nl bool) (err error) {
	if nl {
		if _, err = fmt.Fprintln(p.Writer); err != nil {
			return err
		}
	}
	for i := 0; i < len(p.nest); i++ {
		_, err = p.Writer.Write([]byte(p.Indent))
	}
	return err
}

func (p *PrettyPrinter) Begin(bracket rune, meta bool) (err error) {
	if p.sep {
		if err = p.indent(true); err != nil {
			return err
		}
	}
	p.sep = false
	if meta {
		if _, err := p.Writer.Write([]byte(MetaStr)); err != nil {
			return err
		}
	}
	switch bracket {
	case '(':
		if _, err := p.Writer.Write([]byte("(")); err != nil {
			return err
		}
		p.nest = append(p.nest, ')')
	case '[':
		if _, err := p.Writer.Write([]byte("[")); err != nil {
			return err
		}
		p.nest = append(p.nest, ']')
	case '{':
		if _, err := p.Writer.Write([]byte("{")); err != nil {
			return err
		}
		p.nest = append(p.nest, '}')
	default:
		return fmt.Errorf("illegal opening bracket '%c'", bracket)
	}
	return nil
}

func (p *PrettyPrinter) End() (err error) {
	if len(p.nest) == 0 {
		return errors.New("nothing to end")
	}
	b := p.nest[len(p.nest)-1]
	p.nest = p.nest[:len(p.nest)-1]
	_, err = p.Writer.Write([]byte(string(b)))
	p.sep = true
	return err
}

func (p *PrettyPrinter) Atom(atom string, meta bool, quote QuoteMode) (err error) {
	if p.sep {
		if err = p.indent(true); err != nil {
			return err
		}
	}
	p.sep = true
	if meta {
		if _, err = p.Writer.Write([]byte(MetaStr)); err != nil {
			return err
		}
	}
	switch quote {
	case Qforce:
		err = QuoteTo(atom, p.Writer)
	case QSUPPRESS:
		_, err = p.Writer.Write([]byte(atom))
	default:
		_, err = CondQuoteTo(atom, p.Writer)
	}
	return err
}

func (p *PrettyPrinter) Newline(count int, indent int) error {
	return nil
}
