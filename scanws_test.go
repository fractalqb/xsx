package xsx

import (
	"bytes"
	"fmt"
)

type WsScan bytes.Buffer

func (s *WsScan) cbBegin(meta bool, brace byte) {
	_, err := fmt.Printf("begin: %t %c (%s)\n", meta, brace, (*bytes.Buffer)(s).String())
	if err != nil {
		panic(err)
	}
}

func (s *WsScan) cbEnd(meta bool, brace byte) {
	_, err := fmt.Printf("end: %c (%s)\n", brace, (*bytes.Buffer)(s).String())
	if err != nil {
		panic(err)
	}
}

func (s *WsScan) cbAtom(meta bool, atom []byte, quoted bool) {
	_, err := fmt.Printf("atom: %t [%s] %t (%s)\n",
		meta, atom, quoted, (*bytes.Buffer)(s).String())
	if err != nil {
		panic(err)
	}
}

func ExampleWsBuf_wsBeforeAtom() {
	wsc := bytes.NewBuffer(nil)
	scn := NewScanner(
		(*WsScan)(wsc).cbBegin,
		(*WsScan)(wsc).cbEnd,
		(*WsScan)(wsc).cbAtom)
	scn.WsBuf = wsc
	scn.ScanString("  foo")
	// Output:
	// atom: false [foo] false (  )
}

func ExampleWsBuf_wsAfterAtom() {
	wsc := bytes.NewBuffer(nil)
	scn := NewScanner(
		(*WsScan)(wsc).cbBegin,
		(*WsScan)(wsc).cbEnd,
		(*WsScan)(wsc).cbAtom)
	scn.WsBuf = wsc
	scn.ScanString("foo  ")
	fmt.Printf("ws: (%s)", wsc.String())
	// Output:
	// atom: false [foo] false ()
	// ws: (  )
}
