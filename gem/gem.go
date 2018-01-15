// Package gem provides a GEneral Model for XSX data
package gem

import (
	"fmt"
)

type Expr interface {
	Meta() bool
	SetMeta(flag bool)
}

type expBase int

const (
	maskMeta     = 1
	maskAtomQuot = 2
)

func (e expBase) Meta() bool {
	return e&maskMeta != 0
}

func (e *expBase) SetMeta(flag bool) {
	if flag {
		*e |= maskMeta
	} else {
		*e &= ^maskMeta
	}
}

type Atom struct {
	expBase
	Str string
}

func (a *Atom) Quoted() bool {
	return a.expBase&maskAtomQuot != 0
}

func (a *Atom) SetQuoted(flag bool) {
	if flag {
		a.expBase |= maskAtomQuot
	} else {
		a.expBase &= ^maskAtomQuot
	}
}

type Sequence struct {
	expBase
	Elems []Expr
}

//go:generate -type=Brace
type Brace int

const (
	Undef  Brace = 0
	Paren  Brace = 1 << 1
	Square Brace = 2 << 1
	Curly  Brace = 3 << 1
)

const braceMask = 3 << 1

var barceOpen = []rune{0, '(', '[', '{'}
var barceClose = []rune{0, ')', ']', '}'}

func FromRune(c rune) Brace {
	switch c {
	case '(', ')':
		return Paren
	case '[', ']':
		return Square
	case '{', '}':
		return Curly
	default:
		return Undef
	}
}

func (b Brace) Opening() rune {
	if b < 1 || b > 3 {
		panic(fmt.Sprintf("not a valid xsx brace: '%d'", int(b)))
	}
	return barceOpen[b]
}

func (b Brace) Closing() rune {
	if b < 1 || b > 3 {
		panic(fmt.Sprintf("not a valid xsx brace: '%d'", int(b)))
	}
	return barceClose[b]
}

func (s *Sequence) Brace() Brace {
	return Brace(s.expBase & braceMask)
}

func (s *Sequence) SetBrace(b Brace) {
	clear := s.expBase & ^braceMask
	s.expBase = clear | (expBase(b) & braceMask)
}
