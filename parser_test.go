package xsx

import (
	"fmt"
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

func (p *TestParser) Begin(scanPos uint64, isMeta bool, brace rune) (err error) {
	p.checkEvtPtr()
	xpct := p.events[p.evtPtr]
	p.evtPtr++
	if xpct.sTok != scnBegin {
		err = fmt.Errorf("@%d:wrong scanner event: %s, expetced %s",
			scanPos,
			scnNames[scnBegin],
			scnNames[xpct.sTok])
	}
	if xpct.tok != string(brace) {
		err = fmt.Errorf("@%d:wrong brace: %c, expected %s",
			scanPos,
			brace,
			xpct.tok)
	}
	if xpct.meta != isMeta {
		err = fmt.Errorf("@%d:wrong meta on %c: %t", scanPos, brace, isMeta)
	}
	return err
}

func (p *TestParser) End(scanPos uint64, brace rune) (err error) {
	p.checkEvtPtr()
	xpct := p.events[p.evtPtr]
	p.evtPtr++
	if xpct.sTok != scnEnd {
		err = fmt.Errorf("@%d:wrong scanner event: %s, expetced %s",
			scanPos,
			scnNames[scnEnd],
			scnNames[xpct.sTok])
	}
	if xpct.tok != string(brace) {
		err = fmt.Errorf("@%d:wrong brace: %c, expected %s",
			scanPos,
			brace,
			xpct.tok)
	}
	return err
}

func (p *TestParser) Atom(scanPos uint64, isMeta bool, atom string, quoted bool) (err error) {
	p.checkEvtPtr()
	xpct := p.events[p.evtPtr]
	p.evtPtr++
	if xpct.sTok != scnAtom {
		err = fmt.Errorf("@%d:wrong scanner event: %s, expetced %s",
			scanPos,
			scnNames[scnAtom],
			scnNames[xpct.sTok])
	}
	if xpct.tok != atom {
		err = fmt.Errorf("@%d:wrong atom: %s, expected %s",
			scanPos,
			atom,
			xpct.tok)
	}
	if xpct.quot != quoted {
		err = fmt.Errorf("@%d:wrong quotation for atom '%s': %t",
			scanPos,
			atom,
			quoted)
	}
	if xpct.meta != isMeta {
		err = fmt.Errorf("@%d:wrong meta for atom '%s': %t", scanPos, atom, isMeta)
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
