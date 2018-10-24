package xsx

import (
	"fmt"
	"io"
)

type IndentingPrinter struct {
	Writer  io.Writer
	Indent  string
	nest    []rune
	indlvl  int
	needind bool
	needsep bool
}

func Indenting(wr io.Writer, indent string) *IndentingPrinter {
	res := &IndentingPrinter{
		Writer: wr,
		Indent: indent}
	return res
}

func (p *IndentingPrinter) doIndent() (err error) {
	if p.needind {
		for i := 0; i < p.indlvl; i++ {
			if _, err = p.Writer.Write([]byte(p.Indent)); err != nil {
				return err
			}
		}
		p.needind = false
	}
	return nil
}

func (p *IndentingPrinter) Begin(bracket rune, meta bool) (err error) {
	if err = p.doIndent(); err != nil {
		return err
	}
	if p.needsep {
		if _, err = p.Writer.Write([]byte(" ")); err != nil {
			return err
		}
		p.needsep = false
	}
	if meta {
		if _, err := p.Writer.Write([]byte(MetaStr)); err != nil {
			return err
		}
	}
	if isAny(byte(bracket), ccBegin) {
		if _, err := p.Writer.Write([]byte{byte(bracket)}); err != nil {
			return err
		}
		p.nest = append(p.nest, closingRune(bracket))
	} else {
		return fmt.Errorf("illegal opening bracket '%c'", bracket)
	}
	return nil
}

func (p *IndentingPrinter) End() (err error) {
	if err = p.doIndent(); err != nil {
		return err
	}
	b := p.nest[len(p.nest)-1]
	p.nest = p.nest[:len(p.nest)-1]
	_, err = p.Writer.Write([]byte(string(b)))
	p.needsep = true
	return err
}

func (p *IndentingPrinter) Atom(atom string, meta bool, quote QuoteMode) (err error) {
	if err = p.doIndent(); err != nil {
		return err
	}
	if p.needsep {
		if _, err = p.Writer.Write([]byte(" ")); err != nil {
			return err
		}
	} else {
		p.needsep = true
	}
	return printAtom(p.Writer, atom, meta, quote)
}

func (p *IndentingPrinter) Newline(count int, indent int) (err error) {
	for count > 0 {
		if _, err = p.Writer.Write([]byte("\n")); err != nil {
			return err
		}
		count--
	}
	p.indlvl += indent
	p.needind = true
	p.needsep = false
	return err
}
