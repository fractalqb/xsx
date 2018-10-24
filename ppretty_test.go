package xsx

import (
	"os"
)

func ExamplePrinterPretty() {
	p := Pretty(os.Stdout, "  ")
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
	// (
	//   foo
	//   \{
	//     "bar"
	//     ...
	//     \baz
	//   }
	//   \"quux"
	// )
}
