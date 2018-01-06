package xsx

import (
	"bytes"
	"os"
	"testing"
)

func TestSingleQuote(t *testing.T) {
	var buf bytes.Buffer
	EscapeTo("\"", &buf)
	if buf.String() != "\\\"" {
		t.Errorf("expected '\\\"', got '%s'", buf.String())
	}
}

func TestSingleBackslash(t *testing.T) {
	var buf bytes.Buffer
	EscapeTo("\\", &buf)
	if buf.String() != "\\\\" {
		t.Errorf("expected '\\\\', got '%s'", buf.String())
	}
}

func TestNeedQuote(t *testing.T) {
	for _, c := range "\"\\ \t([{" {
		if !NeedQuote(string(c)) {
			t.Errorf("needs quote: %c", c)
		}
	}
	for _, c := range "abcDEF.-" {
		if NeedQuote(string(c)) {
			t.Errorf("does not need quote: %c", c)
		}
	}
}

func ExamplePrint() {
	p := Indenting(os.Stdout, "  ")
	Print(p, B('('), "foo", Nl{1, 1}, Bm('{'), "bar", 4711, End, Nl{1, -1}, End)
	// Output:
	// (foo
	//   \{bar 4711}
	// )
}

func ExamplePrinterCompact() {
	p := Compact(os.Stdout)
	p.Begin('(', false)
	p.Atom("foo", false, Qcond)
	p.Begin('{', true)
	p.Atom("bar", false, Qforce)
	p.Atom("baz", true, Qcond)
	p.End()
	p.Atom("quux", true, Qforce)
	p.End()
	// Output:
	// (foo\{"bar" \baz}\"quux")
}

func ExamplePrinterPretty() {
	p := Pretty(os.Stdout)
	p.Begin('(', false)
	p.Atom("foo", false, Qcond)
	p.Begin('{', true)
	p.Atom("bar", false, Qforce)
	p.Atom("...", false, Qcond)
	p.Atom("baz", true, Qcond)
	p.End()
	p.Atom("quux", true, Qforce)
	p.End()
	// Output:
	// (foo\{"bar" \baz}\"quux")
}
