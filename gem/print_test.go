package gem

import (
	"os"

	"github.com/fractalqb/xsx"
)

func ExamplePrint_atom() {
	pr := xsx.Compact(os.Stdout)
	gem := Atom{Str: "foo"}
	Print(pr, &gem)
	// Output:
	// foo
}

func ExamplePrint_qatom() {
	pr := xsx.Compact(os.Stdout)
	gem := Atom{Str: "foo bar"}
	Print(pr, &gem)
	// Output:
	// "foo bar"
}

func ExamplePrint_eseq() {
	pr := xsx.Compact(os.Stdout)
	gem := Sequence{}
	gem.SetBrace(Square)
	Print(pr, &gem)
	// Output:
	// []
}

func ExamplePrint() {
	pr := xsx.Compact(os.Stdout)
	gem := Sequence{Elems: []Expr{
		&Atom{Str: "foo"},
		&Sequence{Elems: []Expr{&Atom{Str: "bar"}}},
		&Atom{Str: "baz"},
	}}
	gem.SetBrace(Square)
	gem.Elems[0].(*Atom).SetQuoted(true)
	gem.Elems[1].SetMeta(true)
	Print(pr, &gem)
	// Output:
	// ["foo"\(bar)baz]
}
