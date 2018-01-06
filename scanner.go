package xsx

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
)

const (
	Meta    = '\\'
	MetaStr = string(Meta)
)

type scanStat int

const (
	tokNone scanStat = iota
	tokChars
	tokStr
)

type ScanError struct {
	pos uint64
	msg string
	rsn error
}

func (err *ScanError) Error() string {
	return fmt.Sprintf("%d:%s", err.pos, err.msg)
}

func (err *ScanError) Position() uint64 {
	return err.pos
}

func (err *ScanError) Message() string {
	return err.msg
}

func (err *ScanError) Reason() error {
	return err.rsn
}

// BeginFunc is called by Scanner when an opening bracked is detected.
type BeginFunc func(scnPos uint64, isMeta bool, brace rune) error

// EndFunc is called by Scanner when a closing bracked is detected that matches
// the correspondig opening bracked. For not matching bracktes Scanner.Next
// returns a ScanError before EndFunc would have been called.
type EndFunc func(scnPos uint64, brace rune) error

// AtomFunc is called by Scanner when an XSX atom is detected.
type AtomFunc func(scnPos uint64, isMeta bool, atom string, quoted bool) error

// Scanner implements a callback based scanner for XSX files
type Scanner struct {
	cbBegin BeginFunc
	cbEnd   EndFunc
	cbAtom  AtomFunc

	charCount uint64
	stat      scanStat
	strEsc    bool
	meta      bool
	nesting   []rune
	token     bytes.Buffer
}

func NewScanner(
	beginCallback BeginFunc,
	endCallback EndFunc,
	atomCallback AtomFunc) (parser *Scanner) {
	return &Scanner{
		cbBegin: beginCallback,
		cbEnd:   endCallback,
		cbAtom:  atomCallback}
}

func (s *Scanner) Push(c rune) (done bool, err error) {
	done = false
	s.charCount++
	switch s.stat {
	case tokNone:
		switch c {
		case '(':
			s.nestPush(')')
			if err = s.cbBegin(s.charCount, s.meta, '('); err != nil {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx psuh: begin '(' meta=%t failed: %s",
						s.meta,
						err.Error()),
					rsn: err}
			}
			s.meta = false
		case '[':
			s.nestPush(']')
			if err = s.cbBegin(s.charCount, s.meta, '['); err != nil {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx psuh: begin '[' meta=%t failed: %s",
						s.meta,
						err.Error()),
					rsn: err}
			}
			s.meta = false
		case '{':
			s.nestPush('}')
			if err = s.cbBegin(s.charCount, s.meta, '{'); err != nil {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx psuh: begin '{' meta=%t failed: %s",
						s.meta,
						err.Error()),
					rsn: err}
			}
			s.meta = false
		case ')', ']', '}':
			if s.meta {
				if err = s.cbAtom(s.charCount, false, MetaStr, false); err != nil {
					return true, &ScanError{
						pos: s.charCount,
						msg: fmt.Sprintf("xsx push: atom '%s' meta=%t quoted=%t failed: %s",
							MetaStr,
							false,
							false,
							err.Error()),
						rsn: err}
				}
				s.token.Reset()
				s.meta = false
			}
			if s.Depth() == 0 {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("closing bracket '%c' at top level",
						c)}
			}
			if e := s.nestPop(); e != c {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf(
						"(%d):unbalanced bracing: %c, expected %c",
						s.charCount,
						c,
						e)}
			}
			if err = s.cbEnd(s.charCount, c); err != nil {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx push: end '%c' failed: %s",
						c,
						err.Error()),
					rsn: err}
			}
		case '"':
			s.stat = tokStr
		case Meta:
			if s.meta {
				s.stat = tokChars
				s.token.WriteRune(Meta)
				s.meta = false
			} else {
				s.meta = true
			}
		default:
			if !unicode.IsSpace(c) {
				s.stat = tokChars
				s.token.WriteRune(c)
			} else if s.meta {
				if err = s.cbAtom(s.charCount, false, MetaStr, false); err != nil {
					return true, &ScanError{
						pos: s.charCount,
						msg: fmt.Sprintf("xsx push: atom '%s' meta=%t quoted=%t failed: %s",
							MetaStr,
							false,
							false,
							err.Error()),
						rsn: err}
				}
				s.token.Reset()
				s.meta = false
			}
		} // case: tokNone
	case tokChars:
		switch c {
		case '(':
			if err = s.cbAtom(s.charCount, s.meta, s.token.String(), false); err != nil {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx push: atom '%s' meta=%t quoted=%t failed: %s",
						s.token.String(),
						s.meta,
						false,
						err.Error()),
					rsn: err}
			}
			s.token.Reset()
			s.meta = false
			s.stat = tokNone
			s.nestPush(')')
			if err = s.cbBegin(s.charCount, s.meta, '('); err != nil {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx psuh: begin '(' meta=%t failed: %s",
						s.meta,
						err.Error()),
					rsn: err}
			}
		case '[':
			if err = s.cbAtom(s.charCount, s.meta, s.token.String(), false); err != nil {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx push: atom '%s' meta=%t quoted=%t failed: %s",
						s.token.String(),
						s.meta,
						false,
						err.Error()),
					rsn: err}
			}
			s.token.Reset()
			s.meta = false
			s.stat = tokNone
			s.nestPush(']')
			if s.cbBegin(s.charCount, s.meta, '['); err != nil {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx psuh: begin '[' meta=%t failed: %s",
						s.meta,
						err.Error()),
					rsn: err}
			}
		case '{':
			if err = s.cbAtom(s.charCount, s.meta, s.token.String(), false); err != nil {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx push: atom '%s' meta=%t quoted=%t failed: %s",
						s.token.String(),
						s.meta,
						false,
						err.Error()),
					rsn: err}
			}
			s.token.Reset()
			s.meta = false
			s.stat = tokNone
			s.nestPush('}')
			if err = s.cbBegin(s.charCount, s.meta, '{'); err != nil {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx psuh: begin '{' meta=%t failed: %s",
						s.meta,
						err.Error()),
					rsn: err}
			}
		case ')', ']', '}':
			if err = s.cbAtom(s.charCount, s.meta, s.token.String(), false); err != nil {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx push: atom '%s' meta=%t quoted=%t failed: %s",
						s.token.String(),
						s.meta,
						false,
						err.Error()),
					rsn: err}
			}
			s.token.Reset()
			s.meta = false
			s.stat = tokNone
			if s.Depth() == 0 {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx push: closing bracket '%c' at top level",
						c)}
			}
			if e := s.nestPop(); e != c {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf(
						"unbalanced bracing: %c, expected %c",
						c,
						e)}
			}
			if err = s.cbEnd(s.charCount, c); err != nil {
				return true, &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx push: end '%c' failed: %s",
						c,
						err.Error()),
					rsn: err}
			}
			done = s.Depth() == 0
		case '"':
			s.cbAtom(s.charCount, s.meta, s.token.String(), false)
			s.token.Reset()
			s.meta = false
			s.stat = tokStr
		case Meta:
			s.cbAtom(s.charCount, s.meta, s.token.String(), false)
			s.token.Reset()
			s.meta = true
			s.stat = tokNone
		default:
			if unicode.IsSpace(c) {
				s.cbAtom(s.charCount, s.meta, s.token.String(), false)
				s.token.Reset()
				s.meta = false
				s.stat = tokNone
				done = s.Depth() == 0
			} else {
				s.token.WriteRune(c)
			}
		} // case: tokChars
	case tokStr:
		if s.strEsc {
			s.token.WriteRune(c)
			s.strEsc = false
		} else {
			switch c {
			case '"':
				s.cbAtom(s.charCount, s.meta, s.token.String(), true)
				s.token.Reset()
				s.meta = false
				s.stat = tokNone
				done = s.Depth() == 0
			case '\\':
				// assert: !s.strEsc
				s.strEsc = true
			default:
				s.token.WriteRune(c)
			}
		}
	}
	return done, nil
}

func (s *Scanner) Finish() (err error) {
	switch s.stat {
	case tokNone:
		if s.meta {
			if err = s.cbAtom(s.charCount, false, MetaStr, false); err != nil {
				return &ScanError{
					pos: s.charCount,
					msg: fmt.Sprintf("xsx push: atom '%s' meta=%t quoted=%t failed: %s",
						s.token.String(),
						false,
						false,
						err.Error()),
					rsn: err}
			}
		}
	case tokChars:
		if err = s.cbAtom(s.charCount, s.meta, s.token.String(), false); err != nil {
			return &ScanError{
				pos: s.charCount,
				msg: fmt.Sprintf("xsx push: atom '%s' meta=%t quoted=%t failed: %s",
					s.token.String(),
					s.meta,
					false,
					err.Error()),
				rsn: err}
		}
	case tokStr:
		err = &ScanError{
			pos: s.charCount,
			msg: fmt.Sprintf("(%d):finish inside quoted atom", s.charCount)}
		return err
	}
	if s.Depth() > 0 {
		err = &ScanError{
			pos: s.charCount,
			msg: fmt.Sprintf("(%d):finish inside structure", s.charCount)}
	}
	if err == nil {
		s.Reset()
	}
	return err
}

func (s *Scanner) Reset() {
	s.nesting = s.nesting[0:0]
	s.token.Reset()
	s.stat = tokNone
	s.strEsc = false
	s.charCount = 0
	s.meta = false
}

func (s *Scanner) PushString(txt string, final bool) error {
	for _, c := range txt {
		if _, err := s.Push(c); err != nil {
			return err
		}
	}
	if final {
		if err := s.Finish(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scanner) Read(rd io.RuneReader, final bool) (err error) {
	for c, _, err := rd.ReadRune(); err != nil; c, _, err = rd.ReadRune() {
		if _, err := s.Push(c); err != nil {
			return err
		}
	}
	if final {
		if err := s.Finish(); err != nil {
			return err
		}
	}
	return err
}

func (s *Scanner) nestPush(c rune) {
	s.nesting = append(s.nesting, c)
}

func (s *Scanner) nestPop() rune {
	// assert len(s.nesting) > 0
	res := s.nesting[len(s.nesting)-1]
	s.nesting = s.nesting[:len(s.nesting)-1]
	return res
}

func (s *Scanner) Depth() int {
	return len(s.nesting)
}
