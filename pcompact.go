package xsx

import (
	"errors"
	"fmt"
	"io"
)

type CompactPrinter struct {
	Writer io.Writer
	nest   []rune
	sep    bool
}

func Compact(wr io.Writer) *CompactPrinter {
	res := &CompactPrinter{Writer: wr}
	return res
}

func (p *CompactPrinter) Begin(bracket rune, meta bool) error {
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

func (p *CompactPrinter) End() (err error) {
	p.sep = false
	if len(p.nest) == 0 {
		return errors.New("nothing to end")
	}
	b := p.nest[len(p.nest)-1]
	p.nest = p.nest[:len(p.nest)-1]
	_, err = p.Writer.Write([]byte(string(b)))
	return err
}

func (p *CompactPrinter) Atom(atom string, meta bool, quote QuoteMode) (err error) {
	if p.sep {
		if _, err = p.Writer.Write([]byte(" ")); err != nil {
			return err
		}
	}
	p.sep = true
	return printAtom(p.Writer, atom, meta, quote)
}

func (p *CompactPrinter) Newline(count int, indent int) error {
	return nil
}
