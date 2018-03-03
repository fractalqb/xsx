package xsx

import (
	"bytes"
	"fmt"
)

type WsScan bytes.Buffer

func (s *WsScan) begin(meta bool, brace rune) error {
	_, err := fmt.Printf("begin: %t %c (%s)\n", meta, brace, (*bytes.Buffer)(s).String())
	return err
}

func (s *WsScan) end(brace rune) error {
	_, err := fmt.Printf("end: %c (%s)\n", brace, (*bytes.Buffer)(s).String())
	return err
}

func (s *WsScan) atom(meta bool, atom string, quoted bool) error {
	_, err := fmt.Printf("atom: %t [%s] %t (%s)\n",
		meta, atom, quoted, (*bytes.Buffer)(s).String())
	return err
}

func ExampleWsBuf_wsBeforeAtom() {
	wsc := bytes.NewBuffer(nil)
	scn := NewScanner((*WsScan)(wsc).begin, (*WsScan)(wsc).end, (*WsScan)(wsc).atom)
	scn.WsBuf = wsc
	scn.PushString("  foo", true)
	// Output:
	// atom: false [foo] false (  )
}

func ExampleWsBuf_wsAfterAtom() {
	wsc := bytes.NewBuffer(nil)
	scn := NewScanner((*WsScan)(wsc).begin, (*WsScan)(wsc).end, (*WsScan)(wsc).atom)
	scn.WsBuf = wsc
	scn.PushString("foo  ", true)
	fmt.Printf("ws: (%s)", wsc.String())
	// Output:
	// atom: false [foo] false ()
	// ws: (  )
}
