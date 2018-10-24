package xsx

import (
	"os"
)

func ExamplePrinterCompact() {
	p := Compact(os.Stdout)
	mustExample(p.Begin('(', false))
	mustExample(p.Atom("foo", false, Qcond))
	mustExample(p.Begin('{', true))
	mustExample(p.Atom("bar", false, Qforce))
	mustExample(p.Atom("baz", true, Qcond))
	mustExample(p.End())
	mustExample(p.Atom("quux", true, Qforce))
	mustExample(p.End())
	// Output:
	// (foo\{"bar" \baz}\"quux")
}
