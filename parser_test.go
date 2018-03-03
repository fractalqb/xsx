package xsx

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"testing"
)

const (
	scnBegin = iota
	scnEnd
	scnAtom
)

var scnNames = [3]string{"BEGIN", "END", "ATOM"}

type event struct {
	sTok int
	tok  string
	meta bool
	quot bool
}

func EvtBeg(brace rune, meta bool) event {
	return event{scnBegin, string(brace), meta, false}
}

func EvtEnd(brace rune) event {
	return event{scnEnd, string(brace), false, false}
}

func EvtAtm(txt string, quoted bool, meta bool) event {
	return event{scnAtom, txt, meta, quoted}
}

type TestParser struct {
	events []event
	evtPtr int
}

func NewTestParser(events ...event) *TestParser {
	return &TestParser{events: events}
}

func (p *TestParser) checkEvtPtr() (err error) {
	if p.events == nil {
		err = fmt.Errorf("too many events: %d, exected none", p.evtPtr)
	} else if p.evtPtr >= len(p.events) {
		err = fmt.Errorf("too many events: %d, exected %d", p.evtPtr, len(p.events))
	}
	return err
}

func (p *TestParser) Begin(isMeta bool, brace rune) (err error) {
	p.checkEvtPtr()
	xpct := p.events[p.evtPtr]
	p.evtPtr++
	if xpct.sTok != scnBegin {
		err = fmt.Errorf("wrong scanner event: %s, expetced %s",
			scnNames[scnBegin],
			scnNames[xpct.sTok])
	}
	if xpct.tok != string(brace) {
		err = fmt.Errorf("wrong brace: %c, expected %s",
			brace,
			xpct.tok)
	}
	if xpct.meta != isMeta {
		err = fmt.Errorf("wrong meta on %c: %t", brace, isMeta)
	}
	return err
}

func (p *TestParser) End(brace rune) (err error) {
	p.checkEvtPtr()
	xpct := p.events[p.evtPtr]
	p.evtPtr++
	if xpct.sTok != scnEnd {
		err = fmt.Errorf("wrong scanner event: %s, expetced %s",
			scnNames[scnEnd],
			scnNames[xpct.sTok])
	}
	if xpct.tok != string(brace) {
		err = fmt.Errorf("wrong brace: %c, expected %s",
			brace,
			xpct.tok)
	}
	return err
}

func (p *TestParser) Atom(isMeta bool, atom string, quoted bool) (err error) {
	p.checkEvtPtr()
	xpct := p.events[p.evtPtr]
	p.evtPtr++
	if xpct.sTok != scnAtom {
		err = fmt.Errorf("wrong scanner event: %s, expetced %s",
			scnNames[scnAtom],
			scnNames[xpct.sTok])
	}
	if xpct.tok != atom {
		err = fmt.Errorf("wrong atom: %s, expected %s",
			atom,
			xpct.tok)
	}
	if xpct.quot != quoted {
		err = fmt.Errorf("wrong quotation for atom '%s': %t",
			atom,
			quoted)
	}
	if xpct.meta != isMeta {
		err = fmt.Errorf("wrong meta for atom '%s': %t", atom, isMeta)
	}
	return err
}

func ExampleParserExample() {
	p := NewParser(NewTestParser(
		EvtBeg('(', false),
		EvtAtm("this", false, false),
		EvtAtm("is", false, false),
		EvtAtm("a", false, false),
		EvtAtm("test", false, false),
		EvtEnd(')')))
	p.PushString("(this is a test)", true)
	//	p.State.(*TestParser)
	//	if ok {
	//		fmt.Println(tp.evtPtr)
	//	}
}

func TestParserRead(t *testing.T) {
	p := NewParser(NewTestParser(
		EvtBeg('(', false),
		EvtAtm("this", false, false),
		EvtAtm("is", false, false),
		EvtAtm("a", false, false),
		EvtAtm("test", false, false),
		EvtEnd(')'),
	))
	txt := bytes.NewBufferString("(this is a test)")
	err := p.Read(bufio.NewReader(txt), true)
	if err != io.EOF {
		t.Error(err)
	}
}
