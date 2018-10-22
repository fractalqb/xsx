package xsx

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/stvp/assert"
)

// Token Details (tokDtls)
// TokEOI: <empty>
// TokBegin: bracket rune, meta bool
// TokEnd: bracket rune
// TokAtom: atom string, meta bool, quoted bool
func assertThisTok(t *testing.T, expect Token, pp *PullParser, tokDtls ...interface{}) {
	got := pp.LastToken()
	assert.Equal(t, expect, got, expect, " ≠ ", got)
	switch expect {
	case TokEOI:
	case TokBegin:
		assert.Equal(t, byte(tokDtls[0].(rune)), pp.LastBrace(), tokDtls[2:]...)
		assert.Equal(t, tokDtls[1].(bool), pp.WasMeta(), tokDtls[2:]...)
	case TokEnd:
		assert.Equal(t, byte(tokDtls[0].(rune)), pp.LastBrace(), tokDtls[1:]...)
	case TokAtom:
		assert.Equal(t, tokDtls[0].(string), pp.Atom, tokDtls[3:]...)
		assert.Equal(t, tokDtls[1].(bool), pp.WasMeta(), tokDtls[3:]...)
		assert.Equal(t, tokDtls[2].(bool), pp.WasQuot, tokDtls[3:]...)
	default:
		t.Fatalf("illegal expected token: %v", expect)
	}
}

func assertNextTok(t *testing.T, expect Token, pp *PullParser, tokDtls ...interface{}) {
	_, err := pp.Next()
	assert.Nil(t, err)
	assertThisTok(t, expect, pp, tokDtls...)
}

func testOBrackets(t *testing.T, bracket rune, tok Token) {
	in := bytes.NewBuffer([]byte(string(bracket)))
	pp := NewPullParser(bufio.NewReader(in))
	assertNextTok(t, tok, pp, bracket, false)
}

func TestPullParser_OpenBrackets(t *testing.T) {
	testOBrackets(t, '(', TokBegin)
	testOBrackets(t, '[', TokBegin)
	testOBrackets(t, '{', TokBegin)
}

func testOCBrackets(t *testing.T, pair string, ob Token, cb Token) {
	in := bytes.NewBuffer([]byte(pair))
	pp := NewPullParser(bufio.NewReader(in))
	runes := []rune(pair)
	assertNextTok(t, ob, pp, runes[0], false)
	assertNextTok(t, cb, pp, runes[1], false)
}

func TestPullParser_Brackets(t *testing.T) {
	testOCBrackets(t, "()", TokBegin, TokEnd)
	testOCBrackets(t, "[]", TokBegin, TokEnd)
	testOCBrackets(t, "{}", TokBegin, TokEnd)
}

func TestPullParser_atomBeforeBegin(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo()"))
	pp := NewPullParser(bufio.NewReader(in))
	assertNextTok(t, TokAtom, pp, "foo", false, false)
	assertNextTok(t, TokBegin, pp, '(', false)
	assertNextTok(t, TokEnd, pp, ')', false)
	assertNextTok(t, TokEOI, pp)
}

func TestPullParser_atomBeforeEnd(t *testing.T) {
	in := bytes.NewBuffer([]byte("(foo)"))
	pp := NewPullParser(bufio.NewReader(in))
	assertNextTok(t, TokBegin, pp, '(', false)
	assertNextTok(t, TokAtom, pp, "foo", false, false)
	assertNextTok(t, TokEnd, pp, ')', false)
	assertNextTok(t, TokEOI, pp)
}

func TestPullParser_general(t *testing.T) {
	in := bytes.NewBuffer([]byte("(\\foo \\[\"bar\"])"))
	pp := NewPullParser(bufio.NewReader(in))
	assertNextTok(t, TokBegin, pp, '(', false)
	assertNextTok(t, TokAtom, pp, "foo", true, false)
	assertNextTok(t, TokBegin, pp, '[', true)
	assertNextTok(t, TokAtom, pp, "bar", false, true)
	assertNextTok(t, TokEnd, pp, ']')
	assertNextTok(t, TokEnd, pp, ')')
	assertNextTok(t, TokEOI, pp)
}

func TestPullParser_SkipMeta(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo ( bar \\quux \\[das wird überspringen]) baz"))
	pp := NewPullParser(bufio.NewReader(in))
	assertNextTok(t, TokAtom, pp, "foo", false, false)
	assertNextTok(t, TokBegin, pp, '(', false)
	assertNextTok(t, TokAtom, pp, "bar", false, false)
	assertNextTok(t, TokAtom, pp, "quux", true, false)
	pp.SkipMeta()
	assertNextTok(t, TokBegin, pp, '[', true)
	pp.SkipMeta()
	assertNextTok(t, TokEnd, pp, ')')
	assertNextTok(t, TokAtom, pp, "baz", false, false)
	assertNextTok(t, TokEOI, pp)
}

func ExamplePullParser_WasAny() {
	in := bytes.NewBuffer([]byte("{}"))
	pp := NewPullParser(bufio.NewReader(in))
	pp.Next()
	fmt.Printf("Last Token: %s\n", pp.WasAny(TokBegin|TokEnd))
	fmt.Printf("Last Token: %s\n", pp.WasAny(TokAtom))
	// Output:
	// Last Token: begin
	// Last Token: <no token>
}
