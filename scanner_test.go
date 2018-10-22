package xsx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

func exampleBegin(wr io.Writer) func(bool, byte) {
	return func(meta bool, brace byte) {
		_, err := fmt.Fprintf(wr, "begin: %t %c\n", meta, brace)
		if err != nil {
			panic(err)
		}
	}
}

func exampleEnd(wr io.Writer) func(bool, byte) {
	return func(meta bool, brace byte) {
		_, err := fmt.Fprintf(wr, "end: %t %c\n", meta, brace)
		if err != nil {
			panic(err)
		}
	}
}

func exampleAtom(wr io.Writer) func(bool, []byte, bool) {
	return func(meta bool, atom []byte, quoted bool) {
		_, err := fmt.Fprintf(wr, "atom: %t [%s] %t\n", meta, string(atom), quoted)
		if err != nil {
			panic(err)
		}
	}
}

func NewTestScanner(example bool) *Scanner {
	if example {
		return NewScanner(
			exampleBegin(os.Stdout),
			exampleEnd(os.Stdout),
			exampleAtom(os.Stdout),
		)
	} else {
		return NewScanner(
			func(meta bool, cb byte) {},
			func(meta bool, cb byte) {},
			func(meta bool, atom []byte, quoted bool) {},
		)
	}
}

func mustExample(err error) {
	if err != nil {
		fmt.Println("ERROR:", err)
	}
}

func ExampleAtom() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("foo"))
	// Output:
	// atom: false [foo] false
}

func ExampleDoubleFinish() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("foo"))
	s.Finish()
	// Output:
	// atom: false [foo] false
}

func ExampleWsAtom() {
	s := NewTestScanner(true)
	mustExample(s.ScanString(" foo"))
	// Output:
	// atom: false [foo] false
}

func ExampleAtomWs() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("foo "))
	// Output:
	// atom: false [foo] false
}

func ExampleQAtom() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("\"foo\""))
	// Output:
	// atom: false [foo] true
}

func ExampleWsQAtom() {
	p := NewTestScanner(true)
	mustExample(p.ScanString(" \"foo\""))
	// Output:
	// atom: false [foo] true
}

func ExampleQAtomWithEsc() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("\"ab\\\\cd\""))
	// Output:
	// atom: false [ab\cd] true
}

func ExampleNil1() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("()"))
	mustExample(s.ScanString(" ()"))
	mustExample(s.ScanString("( )"))
	mustExample(s.ScanString("() "))
	mustExample(s.ScanString(" () "))
	mustExample(s.ScanString(" ( ) "))
	// Output:
	// begin: false (
	// end: false )
	// begin: false (
	// end: false )
	// begin: false (
	// end: false )
	// begin: false (
	// end: false )
	// begin: false (
	// end: false )
	// begin: false (
	// end: false )
}

func ExampleNil2() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("[]"))
	mustExample(s.ScanString(" []"))
	mustExample(s.ScanString("[ ]"))
	mustExample(s.ScanString("[] "))
	mustExample(s.ScanString(" [] "))
	mustExample(s.ScanString(" [ ] "))
	// Output:
	// begin: false [
	// end: false ]
	// begin: false [
	// end: false ]
	// begin: false [
	// end: false ]
	// begin: false [
	// end: false ]
	// begin: false [
	// end: false ]
	// begin: false [
	// end: false ]
}

func ExampleNil3() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("{}"))
	mustExample(s.ScanString(" {}"))
	mustExample(s.ScanString("{ }"))
	mustExample(s.ScanString("{} "))
	mustExample(s.ScanString(" {} "))
	mustExample(s.ScanString(" { } "))
	// Output:
	// begin: false {
	// end: false }
	// begin: false {
	// end: false }
	// begin: false {
	// end: false }
	// begin: false {
	// end: false }
	// begin: false {
	// end: false }
	// begin: false {
	// end: false }
}

func TestUnbalanced(t *testing.T) {
	s := NewTestScanner(false)
	err := s.ScanString("(}")
	if err == nil {
		t.Error("expected unbalanced bracing error, got no error")
	} else if scnErr, ok := err.(*ScanError); !ok {
		t.Errorf("expected unbalanced bracing error, got: %s", err)
	} else {
		if scnErr.Position() != 1 {
			t.Errorf("unbalanced bracing in wrong position: %d", scnErr.Position())
		}
		if scnErr.Message() != "unbalanced bracing: '}', expected ')'" {
			t.Error(scnErr.Message())
		}
		if scnErr.Error() != "@1:unbalanced bracing: '}', expected ')'" {
			t.Error(scnErr.Error())
		}
	}
}

func TestTokenTouchesUnbalanced(t *testing.T) {
	s := NewTestScanner(false)
	err := s.ScanString("(foo}")
	if err == nil {
		t.Error("expected unbalanced bracing error, got no error")
	} else if scnErr, ok := err.(*ScanError); !ok {
		t.Errorf("expected unbalanced bracing error, got: %s", err)
	} else if scnErr.Position() != 4 {
		t.Errorf("unbalanced bracing in wrong position: %d", scnErr.Position())
	}
}

func TestPrematureEndOfString(t *testing.T) {
	s := NewTestScanner(false)
	err := s.ScanString("\"foo")
	if err == nil {
		t.Error("premature end in quoted atom error, got one")
	} else if scnErr, ok := err.(*ScanError); !ok {
		t.Errorf("expected scan error, got: %s", err)
	} else if scnErr.pos != 4 {
		t.Errorf("scan error in wrong position: %d", scnErr.Position())
	}
}

func beginFail(isMeta bool, brace byte) {
	panic(fmt.Errorf("begin fails with meta=%t brace=%c", isMeta, brace))
}

func TestScanFailOnBegin(t *testing.T) {
	for _, xsx := range []string{"(", "[", "{"} {
		s := NewScanner(beginFail, EndNop, AtomNop)
		err := s.ScanString(xsx)
		if err == nil {
			t.Error("expected error, got none")
		}
		if err.Error() != "@0:begin fails with meta=false brace="+xsx {
			t.Errorf("unexpected error message: '%s'", err.Error())
		}
	}
}

func ExampleScanEndFromNoToken() {
	for _, xsx := range []string{"( )", "[ ]", "{ }"} {
		s := NewTestScanner(true)
		mustExample(s.ScanString(xsx))
	}
	// Output:
	// begin: false (
	// end: false )
	// begin: false [
	// end: false ]
	// begin: false {
	// end: false }
}

func ExampleNopFuns() {
	s := NewTestScanner(false)
	mustExample(s.ScanString("(x)"))
	// Output:
}

func ExampleMetaAtom() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("\\foo"))
	// Output:
	// atom: true [foo] false
}

func ExampleMetachaInList() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("(\\)"))
	// Output:
	// begin: false (
	// atom: false [\] false
	// end: false )
}

func ExampleMetacharAtom1() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("\\"))
	// Output:
	// atom: false [\] false
}

func ExampleMetacharAtom2() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("\\ "))
	// Output:
	// atom: false [\] false
}

func ExampleMetacharAtom3() {
	s := NewTestScanner(true)
	mustExample(s.ScanString(" \\"))
	// Output:
	// atom: false [\] false
}

func ExampleStartTouchAtom1() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("foo(bar)"))
	// Output:
	// atom: false [foo] false
	// begin: false (
	// atom: false [bar] false
	// end: false )
}

func ExampleStartTouchAtom2() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("foo[bar]"))
	// Output:
	// atom: false [foo] false
	// begin: false [
	// atom: false [bar] false
	// end: false ]
}

func ExampleStartTouchAtom3() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("foo{bar}"))
	// Output:
	// atom: false [foo] false
	// begin: false {
	// atom: false [bar] false
	// end: false }
}

func ExampleAtomTouch1() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("foo\"bar\""))
	// Output:
	// atom: false [foo] false
	// atom: false [bar] true
}

func ExampleAtomTouch2() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("\"foo\"bar"))
	// Output:
	// atom: false [foo] true
	// atom: false [bar] false
}

func ExampleAtomTouch3() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("\"foo\"\"bar\""))
	// Output:
	// atom: false [foo] true
	// atom: false [bar] true
}

func ExampleAtomTouch4() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("foo\\bar"))
	// Output:
	// atom: false [foo] false
	// atom: true [bar] false
}

func ExampleEscMeta1() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("\\\\"))
	// Output:
	// atom: true [\] false
}

func ExampleEscMeta2() {
	s := NewTestScanner(true)
	mustExample(s.ScanString(" \\\\"))
	// Output:
	// atom: true [\] false
}

func ExampleEscMeta3() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("\\\\ "))
	// Output:
	// atom: true [\] false
}

func ExampleEscMetaAtom() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("\\\\foo"))
	// Output:
	// atom: true [\] false
	// atom: false [foo] false
}

func ExampleEscMetaStart() {
	s := NewTestScanner(true)
	mustExample(s.ScanString("\\\\(_)"))
	// Output:
	// atom: true [\] false
	// begin: false (
	// atom: false [_] false
	// end: false )
}

func twoStepScan(t *testing.T, txt string) {
	in := []byte(txt)
	var out bytes.Buffer
	scn := NewScanner(
		exampleBegin(&out),
		exampleEnd(&out),
		exampleAtom(&out),
	)
	err := scn.Scan(in)
	if err != nil {
		t.Fatal(err)
	}
	err = scn.Finish()
	if err != nil {
		t.Fatal(err)
	}
	expect := bytes.Repeat(out.Bytes(), 1)
	for split := 1; split < len(in); split++ {
		scn.Reset()
		out.Reset()
		err = scn.Scan(in[:split])
		if err != nil {
			t.Fatalf("%d[%s|%s]: %s", split, in[:split], in[split:], err)
		}
		err = scn.Scan(in[split:])
		if err != nil {
			t.Fatalf("%d[%s|%s]: %s", split, in[:split], in[split:], err)
		}
		err = scn.Finish()
		if err != nil {
			t.Fatalf("%d[%s|%s]: %s", split, in[:split], in[split:], err)
		}
		if !bytes.Equal(expect, out.Bytes()) {
			t.Fatalf("event sequence differs: %d[%s|%s]\n%s------------------\n%s",
				split,
				in[:split], in[split:],
				expect,
				out.String())
		}
	}
}

func TestScanner_2Suatom(t *testing.T) {
	twoStepScan(t, `foo`)
	twoStepScan(t, ` foo`)
	twoStepScan(t, `foo `)
	twoStepScan(t, ` foo `)
}

func TestScanner_2Smuatom(t *testing.T) {
	twoStepScan(t, `\foo`)
	twoStepScan(t, ` \foo`)
	twoStepScan(t, `\foo `)
	twoStepScan(t, ` \foo `)
}

func TestScanner_2Sqatom(t *testing.T) {
	twoStepScan(t, `"foo bar"`)
	twoStepScan(t, ` "foo bar"`)
	twoStepScan(t, `"foo bar" `)
	twoStepScan(t, ` "foo bar" `)
}

func TestScanner_2Smqatom(t *testing.T) {
	twoStepScan(t, `\"foo bar"`)
	twoStepScan(t, ` \"foo bar"`)
	twoStepScan(t, `\"foo bar" `)
	twoStepScan(t, ` \"foo bar" `)
}

func TestScanner_2Sqatomesc(t *testing.T) {
	twoStepScan(t, `"foo\\bar"`)
	twoStepScan(t, ` "foo\\bar"`)
	twoStepScan(t, `"foo\\bar" `)
	twoStepScan(t, ` "foo\\bar" `)
}

func TestScanner_2Semptylist(t *testing.T) {
	twoStepScan(t, `()`)
	twoStepScan(t, ` ( ) `)
	twoStepScan(t, `[]`)
	twoStepScan(t, ` [ ] `)
	twoStepScan(t, `{}`)
	twoStepScan(t, ` { } `)
}

func TestScanner_2Smemptylist(t *testing.T) {
	twoStepScan(t, `\()`)
	twoStepScan(t, ` \( ) `)
	twoStepScan(t, `\[]`)
	twoStepScan(t, ` \[ ] `)
	twoStepScan(t, `\{}`)
	twoStepScan(t, ` \{ } `)
}

func ExampleAtomWith2Escapes() {
	s := NewTestScanner(true)
	mustExample(s.ScanString(`"quote: \"; backslash: \\ !"`))
	// Output:
	// atom: false [quote: "; backslash: \ !] true
}

func TestScanner_2Sqatom3esc(t *testing.T) {
	twoStepScan(t, `"foo\"bar\\ba\\z"`)
}

func ExampleScanner_EscIn2ndSplit() {
	s := NewTestScanner(true)
	mustExample(s.Scan([]byte(`\"foo `)))
	mustExample(s.Scan([]byte(`e\\s`)))
	mustExample(s.Scan([]byte(`cape"`)))
	// Output:
	// atom: true [foo e\scape] true
}
