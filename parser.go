package xsx

// State is an interface for parser states that are changed by the XSX event
// callback methods of the state object.
type State interface {
	Begin(isMeta bool, brace byte)
	End(isMeta bool, brace byte)
	Atom(isMeta bool, atom []byte, quoted bool)
}

// Parser combines a Scanner with a parser state so that scanner methods can be
// called on the parser and the resulting state is accessible through the State
// field.
type Parser struct {
	*Scanner
	State State
}

func NewParser(p State) *Parser {
	scn := NewScanner(p.Begin, p.End, p.Atom)
	return &Parser{State: p, Scanner: scn}
}
