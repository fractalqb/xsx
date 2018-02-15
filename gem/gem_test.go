package gem

import (
	"testing"

	"github.com/stvp/assert"
)

func TestBrace_Opening(t *testing.T) {
	assert.Equal(t, '(', Paren.Opening())
	assert.Equal(t, '[', Square.Opening())
	assert.Equal(t, '{', Curly.Opening())
}

func TestBrace_Closing(t *testing.T) {
	assert.Equal(t, ')', Paren.Closing())
	assert.Equal(t, ']', Square.Closing())
	assert.Equal(t, '}', Curly.Closing())
}

func TestAtom_Meta(t *testing.T) {
	a := Atom{Str: "foo"}
	quoted := a.Quoted()
	assert.False(t, a.Meta())
	a.SetMeta(true)
	assert.True(t, a.Meta())
	assert.Equal(t, quoted, a.Quoted())
	a.SetMeta(true)
	assert.True(t, a.Meta())
	assert.Equal(t, quoted, a.Quoted())
}

func TestSequence_Meta(t *testing.T) {
	s := Sequence{}
	brace := s.Brace()
	assert.False(t, s.Meta())
	s.SetMeta(true)
	assert.True(t, s.Meta())
	assert.Equal(t, brace, s.Brace())
	s.SetMeta(true)
	assert.True(t, s.Meta())
	assert.Equal(t, brace, s.Brace())
}

func TestSequence_Brace(t *testing.T) {
	seq := Sequence{}
	seq.SetBrace(Paren)
	assert.Equal(t, Paren, seq.Brace())
	seq.SetBrace(Square)
	assert.Equal(t, Square, seq.Brace())
	seq.SetBrace(Curly)
	assert.Equal(t, Curly, seq.Brace())
}

func TestFromRune(t *testing.T) {
	assert.Equal(t, Paren, FromRune('('))
	assert.Equal(t, Paren, FromRune(')'))
	assert.Equal(t, Square, FromRune('['))
	assert.Equal(t, Square, FromRune(']'))
	assert.Equal(t, Curly, FromRune('{'))
	assert.Equal(t, Curly, FromRune('}'))
	assert.Equal(t, Undef, FromRune('x'))
}
