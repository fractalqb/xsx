package xsx

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Token int

const (
	noToken        = 0
	TokBegin Token = (1 << iota)
	TokEnd
	TokAtom
	TokEOI
)

func (t Token) String() string {
	switch t {
	case noToken:
		return "<no token>"
	case TokEOI:
		return "EOI"
	case TokBegin:
		return "begin"
	case TokEnd:
		return "end"
	case TokAtom:
		return "atom"
	default:
		return fmt.Sprintf("<illgeal token: %d>", t)
	}
}

type tokInfo struct {
	tok     Token
	meta    bool
	bracket rune
}

type PullParser struct {
	rd           *bufio.Reader
	tokWr, tokRd int
	toks         [2]tokInfo
	Atom         string // atom can only appear in tokInfo[0]
	WasQuot      bool
	scn          *Scanner
}

func NewPullParser(rd *bufio.Reader) *PullParser {
	res := &PullParser{rd: rd}
	scn := NewScanner(
		func(scnPos uint64, isMeta bool, bracket rune) error {
			ti := &res.toks[res.tokWr]
			ti.tok = TokBegin
			ti.bracket = bracket
			ti.meta = isMeta
			res.tokWr++
			return nil
		},
		func(scnPos uint64, bracket rune) error {
			ti := &res.toks[res.tokWr]
			ti.tok = TokEnd
			ti.bracket = bracket
			// ti.meta undefined
			res.tokWr++
			if res.scn.Depth() < 1 {
				res.scn.Reset()
			}
			return nil
		},
		func(scnPos uint64, isMeta bool, atom string, quoted bool) error {
			ti := &res.toks[res.tokWr]
			ti.tok = TokAtom
			ti.meta = isMeta
			// ti.bracket undefined
			res.Atom = atom
			res.WasQuot = quoted
			res.tokWr++
			return nil
		})
	res.scn = scn
	return res
}

func (p *PullParser) Next() (res Token, err error) {
	if p.tokRd < p.tokWr {
		res = p.toks[p.tokRd].tok
		p.tokRd++
		return res, nil
	}
	p.tokWr = 0
	p.tokRd = 0
	for p.tokRd >= p.tokWr {
		r, _, err := p.rd.ReadRune()
		if err == io.EOF {
			err = p.scn.Finish()
			if p.tokWr >= len(p.toks) {
				panic("assert failed")
			}
			p.toks[p.tokWr].tok = TokEOI
			p.tokWr++
			if p.tokRd >= p.tokWr {
				panic("assert failed")
			}
			res = p.toks[p.tokRd].tok
			p.tokRd++
			return res, err
		} else if err != nil {
			return TokEOI, err
		} else if _, err = p.scn.Push(r); err != nil {
			return TokEOI, err
		}
	}
	res = p.toks[p.tokRd].tok
	p.tokRd++
	return res, nil
}

func (p *PullParser) LastToken() Token {
	if p.tokRd <= 0 {
		return 0
	}
	return p.toks[p.tokRd-1].tok
}

func (p *PullParser) LastBrace() rune {
	if p.tokRd <= 0 {
		return 0
	}
	// TODO may check if last token was a begin or end
	return p.toks[p.tokRd-1].bracket
}

func (p *PullParser) WasMeta() bool {
	if p.tokRd <= 0 {
		return false
	}
	// TODO may check if last token was not end
	return p.toks[p.tokRd-1].meta
}

// SkipMeta skips all tokens that are part of the current meta XSX, if so.
// If the last token was not meta, nothing is skipped. Otherwise SkipMeta stops
// at the last token of the current meta.
func (p *PullParser) SkipMeta() (err error) {
	if !p.WasMeta() {
		return nil
	}
	switch p.LastToken() {
	case TokAtom:
		return nil
	case TokBegin:
		depth := 1
		for depth > 0 {
			tok, err := p.Next()
			if err != nil {
				return err
			}
			switch tok {
			case TokBegin:
				depth++
			case TokEnd:
				depth--
			case TokEOI:
				return PullEOI
			}
		}
		return nil
	default:
		panic("unreachable code")
	}
}

func (p *PullParser) WasAny(joinToken Token) Token {
	return joinToken & p.LastToken()
}

type Unexpected int

const (
	PullEOI Unexpected = iota
	PullMeta
	PullNoMeta
)

func (err Unexpected) Error() string {
	switch err {
	case PullEOI:
		return "pulled end of input"
	case PullMeta:
		return "pulled meta token"
	}
	panic("xsx pulled unkonwn unexpected error")
}

type ExpectMeta int

const (
	AllowMeta ExpectMeta = iota
	RequireMeta
	NoMeta
)

func Pulled(what Unexpected, err error) bool {
	if unx, ok := err.(Unexpected); ok {
		return unx == what
	} else {
		return false
	}
}

func (p *PullParser) NextBegin(whichBrackets string, meta ExpectMeta) error {
	_, err := p.Next()
	if err != nil {
		return err
	}
	return p.ExpectBegin(whichBrackets, meta)
}

func checkMeta(p *PullParser, meta ExpectMeta) error {
	switch meta {
	case AllowMeta:
		return nil
	case RequireMeta:
		if !p.WasMeta() {
			return PullNoMeta
		}
	case NoMeta:
		if p.WasMeta() {
			return PullMeta
		}
	default:
		panic("Illegal ExpectMeta: " + strconv.Itoa(int(meta)))
	}
	return nil
}

func (p *PullParser) ExpectBegin(whichBrackets string, meta ExpectMeta) error {
	tok := p.LastToken()
	switch {
	case tok == TokEOI:
		return PullEOI
	case tok != TokBegin:
		return fmt.Errorf("expected begin token, got %s", tok)
	case len(whichBrackets) > 0 &&
		strings.IndexRune(whichBrackets, p.LastBrace()) < 0:
		return fmt.Errorf("expected on of '%s', got %c", whichBrackets, p.LastBrace())
	}
	return checkMeta(p, meta)
}

func (p *PullParser) NextEnd(whichBrackets string) error {
	_, err := p.Next()
	if err != nil {
		return err
	}
	return p.ExpectEnd(whichBrackets)
}

func (p *PullParser) ExpectEnd(whichBrackets string) error {
	tok := p.LastToken()
	switch {
	case tok == TokEOI:
		return PullEOI
	case p.WasMeta():
		return PullMeta
	case tok != TokEnd:
		return fmt.Errorf("expected end token, got %s", tok)
	case len(whichBrackets) > 0 &&
		strings.IndexRune(whichBrackets, p.LastBrace()) < 0:
		return fmt.Errorf("expected on of '%s', got %c", whichBrackets, p.LastBrace())
	default:
		return nil
	}
}

func (p *PullParser) NextAtom(meta ExpectMeta) (string, error) {
	_, err := p.Next()
	if err != nil {
		return "", err
	}
	return p.ExpectAtom(meta)
}

func (p *PullParser) ExpectAtom(meta ExpectMeta) (string, error) {
	tok := p.LastToken()
	switch {
	case tok == TokEOI:
		return "", PullEOI
	case tok != TokAtom:
		return "", fmt.Errorf("expected atom token, got %s", tok)
	}
	return p.Atom, checkMeta(p, meta)
}
