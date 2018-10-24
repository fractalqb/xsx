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

func ExampleWrite() {
	p := Indenting(os.Stdout, "  ")
	mustExample(Write(p, B('('), "foo", Nl{1, 1}, Bm('{'), "bar", 4711, End, Nl{1, -1}, End))
	// Output:
	// (foo
	//   \{bar 4711}
	// )
}
