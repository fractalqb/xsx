package xsx

// State is an interface for parser states that are changed by the XSX event
// callback methods of the state object.
type State interface {
	Begin(isMeta bool, brace rune) error
	End(brace rune) error
	Atom(isMeta bool, atom string, quoted bool) error
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
		func(isMeta bool, brace rune) error {
			return p.Begin(isMeta, brace)
		},
		func(brace rune) error {
			return p.End(brace)
		},
		func(isMeta bool, atom string, quoted bool) error {
			return p.Atom(isMeta, atom, quoted)
		})
	return &Parser{State: p, Scanner: scn}
}
