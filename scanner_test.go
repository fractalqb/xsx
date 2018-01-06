package xsx

import (
	"fmt"
	"testing"
)

func testBegin(scanPos uint64, meta bool, brace rune) error {
	return nil
}

func testEnd(scanPos uint64, brace rune) error {
	return nil
}

func testAtom(scanPos uint64, meta bool, atom string, quoted bool) error {
	return nil
}

func exampleBegin(scanPos uint64, meta bool, brace rune) error {
	_, err := fmt.Printf("begin: %t %c\n", meta, brace)
	return err
}

func exampleEnd(scanPos uint64, brace rune) error {
	_, err := fmt.Printf("end: %c\n", brace)
	return err
}

func exampleAtom(scanPos uint64, meta bool, atom string, quoted bool) error {
	_, err := fmt.Printf("atom: %t [%s] %t\n", meta, atom, quoted)
	return err
}

func NewTestScanner(example bool) *Scanner {
	if example {
		return NewScanner(exampleBegin, exampleEnd, exampleAtom)
	} else {
		return NewScanner(testBegin, testEnd, testAtom)
	}
}

func ExampleAtom() {
	p := NewTestScanner(true)
	p.PushString("foo", true)
	// Output:
	// atom: false [foo] false
}

func ExampleDoubleFinish() {
	p := NewTestScanner(true)
	p.PushString("foo", true)
	p.Finish()
	// Output:
	// atom: false [foo] false
}

func ExampleWsAtom() {
	p := NewTestScanner(true)
	p.PushString(" foo", true)
	// Output:
	// atom: false [foo] false
}

func ExampleAtomWs() {
	p := NewTestScanner(true)
	p.PushString("foo ", true)
	// Output:
	// atom: false [foo] false
}

func ExampleQAtom() {
	p := NewTestScanner(true)
	p.PushString("\"foo\"", true)
	// Output:
	// atom: false [foo] true
}

func ExampleWsQAtom() {
	p := NewTestScanner(true)
	p.PushString(" \"foo\"", true)
	// Output:
	// atom: false [foo] true
}

func ExampleQAtomWithEsc() {
	p := NewTestScanner(true)
	p.PushString("\"ab\\\\cd\"", true)
	// Output:
	// atom: false [ab\cd] true
}

func ExampleNil1() {
	p := NewTestScanner(true)
	p.PushString("()", true)
	p.PushString(" ()", true)
	p.PushString("( )", true)
	p.PushString("() ", true)
	p.PushString(" () ", true)
	p.PushString(" ( ) ", true)
	// Output:
	// begin: false (
	// end: )
	// begin: false (
	// end: )
	// begin: false (
	// end: )
	// begin: false (
	// end: )
	// begin: false (
	// end: )
	// begin: false (
	// end: )
}

func ExampleNil2() {
	p := NewTestScanner(true)
	p.PushString("[]", true)
	p.PushString(" []", true)
	p.PushString("[ ]", true)
	p.PushString("[] ", true)
	p.PushString(" [] ", true)
	p.PushString(" [ ] ", true)
	// Output:
	// begin: false [
	// end: ]
	// begin: false [
	// end: ]
	// begin: false [
	// end: ]
	// begin: false [
	// end: ]
	// begin: false [
	// end: ]
	// begin: false [
	// end: ]
}

func ExampleNil3() {
	p := NewTestScanner(true)
	p.PushString("{}", true)
	p.PushString(" {}", true)
	p.PushString("{ }", true)
	p.PushString("{} ", true)
	p.PushString(" {} ", true)
	p.PushString(" { } ", true)
	// Output:
	// begin: false {
	// end: }
	// begin: false {
	// end: }
	// begin: false {
	// end: }
	// begin: false {
	// end: }
	// begin: false {
	// end: }
	// begin: false {
	// end: }
}

func TestUnbalanced(t *testing.T) {
	p := NewTestScanner(false)
	err := p.PushString("(}", true)
	if err == nil {
		t.Error("expected unbalanced bracing error, got no error")
	} else if scnErr := err.(*ScanError); scnErr == nil {
		t.Errorf("expected unbalanced bracing error, got: %s", err)
	} else if scnErr.Position() != 2 {
		t.Errorf("unbalanced bracing in wrong position: %d", scnErr.Position())
	}
}

func TestTokenTouchesUnbalanced(t *testing.T) {
	p := NewTestScanner(false)
	err := p.PushString("(foo}", true)
	if err == nil {
		t.Error("expected unbalanced bracing error, got no error")
	} else if scnErr := err.(*ScanError); scnErr == nil {
		t.Errorf("expected unbalanced bracing error, got: %s", err)
	} else if scnErr.Position() != 5 {
		t.Errorf("unbalanced bracing in wrong position: %d", scnErr.Position())
	}
}

func TestPrematureEndOfString(t *testing.T) {
	p := NewTestScanner(false)
	err := p.PushString("\"foo", true)
	if err == nil {
		t.Error("premature end in quoted atom error, got one")
	} else if scnErr := err.(*ScanError); scnErr == nil {
		t.Errorf("expected scan error, got: %s", err)
	} else if scnErr.pos != 4 {
		t.Errorf("scan error in wrong position: %d", scnErr.Position())
	}
}

func ExampleMetaAtom() {
	s := NewTestScanner(true)
	s.PushString("\\foo", true)
	// Output:
	// atom: true [foo] false
}

func ExampleMetachaInList() {
	s := NewTestScanner(true)
	s.PushString("(\\)", true)
	// Output:
	// begin: false (
	// atom: false [\] false
	// end: )
}

func ExampleMetacharAtom1() {
	s := NewTestScanner(true)
	s.PushString("\\", true)
	// Output:
	// atom: false [\] false
}

func ExampleMetacharAtom2() {
	s := NewTestScanner(true)
	s.PushString("\\ ", true)
	// Output:
	// atom: false [\] false
}

func ExampleMetacharAtom3() {
	s := NewTestScanner(true)
	s.PushString(" \\", true)
	// Output:
	// atom: false [\] false
}

func ExampleStartTouchAtom1() {
	s := NewTestScanner(true)
	s.PushString("foo(bar)", true)
	// Output:
	// atom: false [foo] false
	// begin: false (
	// atom: false [bar] false
	// end: )
}

func ExampleStartTouchAtom2() {
	s := NewTestScanner(true)
	s.PushString("foo[bar]", true)
	// Output:
	// atom: false [foo] false
	// begin: false [
	// atom: false [bar] false
	// end: ]
}

func ExampleStartTouchAtom3() {
	s := NewTestScanner(true)
	s.PushString("foo{bar}", true)
	// Output:
	// atom: false [foo] false
	// begin: false {
	// atom: false [bar] false
	// end: }
}

func ExampleAtomTouch1() {
	s := NewTestScanner(true)
	s.PushString("foo\"bar\"", true)
	// Output:
	// atom: false [foo] false
	// atom: false [bar] true
}

func ExampleAtomTouch2() {
	s := NewTestScanner(true)
	s.PushString("\"foo\"bar", true)
	// Output:
	// atom: false [foo] true
	// atom: false [bar] false
}

func ExampleAtomTouch3() {
	s := NewTestScanner(true)
	s.PushString("\"foo\"\"bar\"", true)
	// Output:
	// atom: false [foo] true
	// atom: false [bar] true
}

func ExampleAtomTouch4() {
	s := NewTestScanner(true)
	s.PushString("foo\\bar", true)
	// Output:
	// atom: false [foo] false
	// atom: true [bar] false
}

func ExampleEscMeta1() {
	s := NewTestScanner(true)
	s.PushString("\\\\", true)
	// Output:
	// atom: false [\] false
}

func ExampleEscMeta2() {
	s := NewTestScanner(true)
	s.PushString(" \\\\", true)
	// Output:
	// atom: false [\] false
}

func ExampleEscMeta3() {
	s := NewTestScanner(true)
	s.PushString("\\\\ ", true)
	// Output:
	// atom: false [\] false
}

func ExampleEscMetaAtom() {
	s := NewTestScanner(true)
	s.PushString("\\\\foo", true)
	// Output:
	// atom: false [\foo] false
}

func ExampleEscMetaStart() {
	s := NewTestScanner(true)
	s.PushString("\\\\(", true)
	// Output:
	// atom: false [\] false
	// begin: false (
}
