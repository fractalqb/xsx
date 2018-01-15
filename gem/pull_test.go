package gem

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/fractalqb/xsx"
	"github.com/stvp/assert"
)

func TestPullNext_atom(t *testing.T) {
	txt := bytes.NewBufferString("foo")
	p := xsx.NewPullParser(bufio.NewReader(txt))
	gem, err := ReadNext(p)
	assert.Nil(t, err)
	atom, ok := gem.(*Atom)
	assert.True(t, ok)
	assert.False(t, atom.Meta())
	assert.False(t, atom.Quoted())
	assert.Equal(t, "foo", atom.Str)
}

func TestPullNext_sequence(t *testing.T) {
	txt := bytes.NewBufferString("(foo \"bar\" \\baz)")
	p := xsx.NewPullParser(bufio.NewReader(txt))
	gem, err := ReadNext(p)
	assert.Nil(t, err)
	seq, ok := gem.(*Sequence)
	assert.True(t, ok)
	assert.False(t, seq.Meta())
	assert.Equal(t, Paren, seq.Brace())
	assert.Equal(t, 3, len(seq.Elems))
}
