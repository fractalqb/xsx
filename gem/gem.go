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

//go:generate stringer -type Brace
type Brace int

const (
	Undef  Brace = 0
	Paren  Brace = 1 << 0
	Square Brace = 1 << 1
	Curly  Brace = 1 << 2
)

const braceShift = 2
const braceMask = 7 << braceShift

var barceOpen = []rune{0, '(', '[', 0, '{'}
var barceClose = []rune{0, ')', ']', 0, '}'}

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

func (b Brace) Opening() (res rune) {
	if b < 1 || b > 4 {
		panic(fmt.Sprintf("not a valid xsx brace: '%d'", int(b)))
	}
	res = barceOpen[b]
	if res == 0 {
		panic(fmt.Sprintf("not a valid xsx brace: '%d'", int(b)))
	}
	return res
}

func (b Brace) Closing() (res rune) {
	if b < 1 || b > 4 {
		panic(fmt.Sprintf("not a valid xsx brace: '%d'", int(b)))
	}
	res = barceClose[b]
	if res == 0 {
		panic(fmt.Sprintf("not a valid xsx brace: '%d'", int(b)))
	}
	return res
}

func (s *Sequence) Brace() Brace {
	return Brace((s.expBase & braceMask) >> braceShift)
}

func (s *Sequence) SetBrace(b Brace) {
	clear := s.expBase & ^braceMask
	s.expBase = clear | ((expBase(b) << braceShift) & braceMask)
}
