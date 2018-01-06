package xsx

// State is an interface for parser states that are changed by the XSX event
// callback methods of the state object.
type State interface {
	Begin(scanPos uint64, isMeta bool, brace rune) error
	End(scanPos uint64, brace rune) error
	Atom(scanPos uint64, isMeta bool, atom string, quoted bool) error
}

// Parser combines a Scanner with a parser state so that scanner methods can be
// called on the parser and the resulting state is accessible through the State
// field.
type Parser struct {
	*Scanner
	State State
}

func NewParser(p State) *Parser {
	scn := NewScanner(
		func(scanPos uint64, isMeta bool, brace rune) error {
			return p.Begin(scanPos, isMeta, brace)
		},
		func(scanPos uint64, brace rune) error {
			return p.End(scanPos, brace)
		},
		func(scanPos uint64, isMeta bool, atom string, quoted bool) error {
			return p.Atom(scanPos, isMeta, atom, quoted)
		})
	return &Parser{State: p, Scanner: scn}
}
