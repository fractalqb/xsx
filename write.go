package xsx

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
)

// EscapeTo escapes characters in str according to the escape rules of XSX
// quoted atoms. I.e. '"' is replaces with '\"' and '\' is replaced with '\\'.
// The result is written to the dst output buffer.
func EscapeTo(str string, dst io.Writer) (numEsc int, err error) {
	var res int = 0
	for _, c := range str {
		switch c {
		case '"':
			if _, err = dst.Write([]byte("\\\"")); err != nil {
				return 0, err
			}
			res++
		case '\\':
			if _, err = dst.Write([]byte("\\\\")); err != nil {
				return 0, err
			}
			res++
		default:
			if _, err = dst.Write([]byte(string(c))); err != nil {
				return 0, err
			}
		}
	}
	return res, nil
}

func NeedQuote(str string) bool {
	if len(str) == 0 {
		return true
	}
	for _, c := range str {
		switch c {
		case '"', '\\', '(', '[', '{', ' ', '\t':
			return true

		default:
			if unicode.IsSpace(c) {
				return true
			}
		}
	}
	return false
}

func QuoteTo(str string, wr io.Writer) (err error) {
	if _, err = wr.Write([]byte("\"")); err != nil {
		return err
	}
	if _, err = EscapeTo(str, wr); err != nil {
		return err
	}
	_, err = wr.Write([]byte("\""))
	return err
}

func Quoted(str string) string {
	buf := bytes.NewBuffer(nil)
	QuoteTo(str, buf)
	return buf.String()
}

func CondQuoteTo(str string, wr io.Writer) (quoted bool, err error) {
	if NeedQuote(str) {
		err = QuoteTo(str, wr)
		return true, err
	} else {
		_, err = wr.Write([]byte(str))
		return false, err
	}
}

func CondQuoted(str string) (string, bool) {
	if NeedQuote(str) {
		return Quoted(str), true
	} else {
		return str, false
	}
}

type QuoteMode int

const (
	Qcond QuoteMode = iota
	Qforce
	// QSUPPRESS supresses quoting of an atom. This might break XSX syntax!
	QSUPPRESS
)

type Printer interface {
	Begin(bracket rune, meta bool) error
	End() error
	Atom(atom string, meta bool, quote QuoteMode) error
	Newline(count int, indent int) error
}

type B rune
type Bm rune
type printEnd int

// End is passed to the Print function to end the current structure. Print
// will choose the correct bracket to keep them balanced.
const End printEnd = printEnd(0)

type Nl struct {
	Count  int
	Indent int
}

func Print(p Printer, token ...interface{}) (err error) {
	for _, t := range token {
		switch tok := t.(type) {
		case Nl:
			if err = p.Newline(tok.Count, tok.Indent); err != nil {
				return err
			}
		case B:
			if err = p.Begin(rune(tok), false); err != nil {
				return err
			}
		case Bm:
			if err = p.Begin(rune(tok), true); err != nil {
				return err
			}
		case printEnd:
			if err = p.End(); err != nil {
				return err
			}
		case string:
			if err = p.Atom(tok, false, Qcond); err != nil {
				return err
			}
		case int, uint, bool, float32, float64, int8, uint8,
			int16, uint16, int32, uint32, int64, uint64, uintptr:
			str := fmt.Sprint(t)
			if err = p.Atom(str, false, QSUPPRESS); err != nil {
				return err
			}
		default:
			str := fmt.Sprint(t)
			if err = p.Atom(str, false, Qcond); err != nil {
				return err
			}
		}
	}
	return nil
}
