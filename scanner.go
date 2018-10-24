package xsx

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

const (
	// Meta is the rune used as escape for meta expressions.
	Meta = '\\'
	// MetaStr is the string variant of the Meta rune.
	MetaStr = string(Meta)
)

type ScanError struct {
	hint string
	pos  int64
	msg  string
	rsn  error
}

func (err *ScanError) Error() string {
	return fmt.Sprintf("%s@%d:%s", err.hint, err.pos, err.msg)
}

func (err *ScanError) Position() int64 {
	return err.pos
}

func (err *ScanError) Message() string {
	return err.msg
}

func (err *ScanError) Reason() error {
	return err.rsn
}

// BeginFunc is called by Scanner when an opening bracket is detected.
type BeginFunc func(isMeta bool, brace byte)

// BeginNop performs No OPeration on begin event.
func BeginNop(isMeta bool, brace byte) {}

// EndFunc is called by Scanner when a closing bracket is detected that matches
// the corresponding opening bracket. For non-matching brackets Scanner.Next
// returns a ScanError before EndFunc would have been called.
type EndFunc func(isMeta bool, brace byte)

// EndNop performs No OPeration on end event.
func EndNop(isMeta bool, brace byte) {}

// AtomFunc is called by Scanner when an XSX atom is detected.
type AtomFunc func(isMeta bool, atom []byte, quoted bool)

// AtomNop performs No OPeration on atom event.
func AtomNop(isMeta bool, atom []byte, quoted bool) {}

type atomHeadMode int

const (
	aheadPlain atomHeadMode = iota
	aheadQuote
	aheadEsc
)

type Scanner struct {
	Begin     BeginFunc
	End       EndFunc
	Atom      AtomFunc
	SrcHint   string
	WsBuf     *bytes.Buffer
	pos       int64
	meta      bool
	nest      []nesting
	atomHead  []byte
	aheadMode atomHeadMode
	qatomBuf  bytes.Buffer
}

type nesting struct {
	meta   bool
	cbrace byte
}

func NewScanner(begin BeginFunc, end EndFunc, atom AtomFunc) *Scanner {
	return &Scanner{
		Begin: begin,
		End:   end,
		Atom:  atom,
	}
}

func (s *Scanner) Complete() bool {
	return s.atomHead == nil && !s.meta && len(s.nest) == 0
}

func (s *Scanner) Finish() (err error) {
	if len(s.nest) > 0 {
		return &ScanError{
			hint: s.SrcHint,
			pos:  s.pos,
			msg:  "cannot finish scanning in nested expression",
		}
	}
	if s.atomHead != nil {
		if s.aheadMode != aheadPlain {
			return &ScanError{
				hint: s.SrcHint,
				pos:  s.pos,
				msg:  "unterminated quoted atom",
			}
		}
		defer func() {
			if p := recover(); p != nil {
				switch x := p.(type) {
				case *ScanError:
					err = x
				case error:
					err = &ScanError{hint: s.SrcHint, pos: s.pos, msg: x.Error(), rsn: x}
				default:
					err = &ScanError{
						hint: s.SrcHint,
						pos:  s.pos,
						msg:  fmt.Sprintf("%T:[%v]", p, p),
					}
				}
			}
		}()
		s.Atom(s.meta, s.atomHead, s.aheadMode == aheadQuote)
		s.meta = false
		s.atomHead = nil
	} else if s.meta {
		s.Atom(false, metaAtom, false)
	}
	return nil
}

func (s *Scanner) Depth() int { return len(s.nest) }

func (s *Scanner) Reset() {
	s.pos = 0
	s.meta = false
	if s.nest != nil {
		s.nest = s.nest[:0]
	}
	s.atomHead = nil
}

func (s *Scanner) push(meta bool, closing byte) {
	s.nest = append(s.nest, nesting{meta, closing})
}

func (s *Scanner) pop(found byte) (meta bool) {
	if end := len(s.nest); end > 0 {
		end--
		n := s.nest[end]
		s.nest = s.nest[:end]
		if n.cbrace != found {
			panic(fmt.Errorf("unbalanced bracing: '%c', expected '%c'",
				found, n.cbrace))
		}
		return n.meta
	} else {
		panic(fmt.Errorf("%s:%d:poping '%c' from unnested",
			s.SrcHint, s.pos, found))
	}
}

// TODO make it fast & use it everywhere (implemented 1st for ppretty)
func closing(open byte) byte {
	switch open {
	case '(':
		return ')'
	case '[':
		return ']'
	case '{':
		return '}'
	default:
		return 0
	}
}

type cClass int

const (
	ccSpace cClass = (1 << iota)
	ccBegin
	ccEnd
	ccTok
)

var cclasses = make([]cClass, 256)

func init() {
	for _, c := range []byte{'\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0} {
		cclasses[c] |= ccSpace
	}
	for _, c := range []byte{'(', '[', '{'} {
		cclasses[c] |= ccBegin
	}
	for _, c := range []byte{')', ']', '}'} {
		cclasses[c] |= ccEnd
	}
	for _, c := range []byte{'"', Meta} {
		cclasses[c] |= ccTok
	}
}

func isAny(c byte, class cClass) bool {
	return (cclasses[c] & class) != 0
}

func (s *Scanner) skipspace(txt []byte) (res int) {
	if s.WsBuf == nil {
		for res < len(txt) {
			if isAny(txt[res], ccSpace) {
				res++
			} else {
				return res
			}
		}
	} else {
		s.WsBuf.Reset()
		for res < len(txt) {
			if isAny(txt[res], ccSpace) {
				s.WsBuf.WriteByte(txt[res])
				res++
			} else {
				return res
			}
		}
	}
	return res
}

func skipUAtom(txt []byte) (atom int) {
	for atom < len(txt) {
		if isAny(txt[atom], ccSpace|ccBegin|ccEnd|ccTok) {
			return atom
		}
		atom++
	}
	return -1
}

func skipQAtom(txt []byte, sb *bytes.Buffer) (atom int, ahead atomHeadMode) {
	sb.Reset()
	for atom < len(txt) {
		switch c := txt[atom]; c {
		case '"':
			return atom, aheadQuote
		case '\\':
			sb.Reset()
			sb.Write(txt[:atom]) // TODO error
			esc := true
			for atom++; atom < len(txt); atom++ {
				if esc {
					sb.WriteByte(txt[atom])
					esc = false
				} else {
					switch c := txt[atom]; c {
					case '"':
						return atom, aheadQuote
					case '\\':
						esc = true
					default:
						sb.WriteByte(c)
					}
				}
			}
			if esc {
				return -1, aheadEsc
			} else {
				return -1, aheadQuote
			}
		}
		atom++
	}
	return -1, aheadQuote
}

func (s *Scanner) callBegin(o, c byte) {
	s.Begin(s.meta, o)
	s.push(s.meta, c)
	s.meta = false
}

func (s *Scanner) callEnd(c byte) {
	if s.meta {
		s.Atom(false, metaAtom, false)
		s.meta = false
	}
	m := s.pop(c)
	s.End(m, c)
}

var metaAtom = []byte{Meta}

func (s *Scanner) Scan(txt []byte) (err error) {
	rp, end := int64(0), int64(len(txt))
	defer func() {
		s.pos += rp
		if p := recover(); p != nil {
			switch x := p.(type) {
			case *ScanError:
				err = x
			case error:
				err = &ScanError{hint: s.SrcHint, pos: s.pos, msg: x.Error(), rsn: x}
			default:
				err = &ScanError{
					hint: s.SrcHint,
					pos:  s.pos,
					msg:  fmt.Sprintf("%T:[%v]", p, p),
				}
			}
		}
	}()
	if s.atomHead != nil {
		if end == 0 {
			return nil
		}
		if s.aheadMode == aheadPlain {
			aLen := skipUAtom(txt)
			if aLen < 0 {
				s.atomHead = append(s.atomHead, txt...)
				rp = end
				return nil
			}
			s.atomHead = append(s.atomHead, txt[:aLen]...)
			s.Atom(s.meta, s.atomHead, false)
			s.meta = false
			s.atomHead = nil
			rp = int64(aLen)
		} else {
			if s.aheadMode == aheadEsc {
				s.atomHead = append(s.atomHead, txt[rp])
				rp++
			}
			aLen, aEsc := skipQAtom(txt[rp:], &s.qatomBuf)
			if aLen < 0 {
				if s.qatomBuf.Len() == 0 {
					s.atomHead = append(s.atomHead, txt...)
				} else {
					s.atomHead = append(s.atomHead, s.qatomBuf.Bytes()...)
				}
				s.aheadMode = aEsc
				return nil
			}
			if s.qatomBuf.Len() == 0 {
				s.atomHead = append(s.atomHead, txt[rp:rp+int64(aLen)]...)
			} else {
				s.atomHead = append(s.atomHead, s.qatomBuf.Bytes()...)
			}
			s.Atom(s.meta, s.atomHead, true)
			s.meta = false
			s.atomHead = nil
			rp += int64(aLen + 1)
		}
	}
	// assert s.atomHead == nil
	for rp < end {
		if wse := s.skipspace(txt[rp:]); wse > 0 {
			if s.meta {
				s.Atom(false, metaAtom, false)
				s.meta = false
			}
			rp += int64(wse)
			if rp >= end {
				return nil
			}
		}
		switch txt[rp] {
		case '(':
			s.callBegin('(', ')')
			rp++
		case '[':
			s.callBegin('[', ']')
			rp++
		case '{':
			s.callBegin('{', '}')
			rp++
		case ')':
			s.callEnd(')')
			rp++
		case ']':
			s.callEnd(']')
			rp++
		case '}':
			s.callEnd('}')
			rp++
		case '"':
			rp++
			aLen, aEsc := skipQAtom(txt[rp:], &s.qatomBuf)
			if aLen < 0 {
				if len := s.qatomBuf.Len(); len == 0 {
					s.atomHead = make([]byte, end-rp)
					copy(s.atomHead, txt[rp:])
				} else {
					s.atomHead = make([]byte, len)
					copy(s.atomHead, s.qatomBuf.Bytes())
				}
				s.aheadMode = aEsc
				rp = end
			} else {
				if s.qatomBuf.Len() == 0 {
					ae := rp + int64(aLen)
					s.Atom(s.meta, txt[rp:ae], true)
				} else {
					s.Atom(s.meta, s.qatomBuf.Bytes(), true)
				}
				s.meta = false
				rp += int64(aLen + 1)
			}
		case Meta:
			if s.meta {
				s.Atom(true, metaAtom, false)
				s.meta = false
			} else {
				s.meta = true
			}
			rp++
		default:
			aLen := skipUAtom(txt[rp:])
			if aLen < 0 {
				s.atomHead = make([]byte, end-rp)
				copy(s.atomHead, txt[rp:])
				s.aheadMode = aheadPlain
				rp = end
			} else {
				ae := rp + int64(aLen)
				s.Atom(s.meta, txt[rp:ae], false)
				s.meta = false
				rp = ae
			}
		}
	}
	return err
}

func (s *Scanner) ScanString(str string) (err error) {
	err = s.Scan([]byte(str))
	if err != nil {
		return err
	}
	err = s.Finish()
	return err
}

var buf4k = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

func (s *Scanner) Read(rd io.Reader) (err error) {
	buf := buf4k.Get().([]byte)
	defer func() { buf4k.Put(buf) }()
	for sz, err := rd.Read(buf); err == nil; sz, err = rd.Read(buf) {
		serr := s.Scan(buf[:sz])
		if serr != nil {
			return serr
		}
	}
	if err == io.EOF {
		err = s.Finish()
	}
	return err
}
