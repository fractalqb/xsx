package gem

import (
	"testing"

	"git.fractalqb.de/fractalqb/xsx"
	"github.com/stvp/assert"
)

func TestParse_complex(t *testing.T) {
	var pstat State
	scn := xsx.NewParser(&pstat)
	scn.PushString(`foo
	(bar \[baz])
	\4711`, true)
	assert.Equal(t, 3, len(pstat.Results))
	a, ok := pstat.Results[0].(*Atom)
	assert.True(t, ok)
	assert.False(t, a.Meta())
	assert.False(t, a.Quoted())
	assert.Equal(t, "foo", a.Str)
	s, ok := pstat.Results[1].(*Sequence)
	assert.True(t, ok)
	assert.False(t, s.Meta())
	assert.Equal(t, Paren, s.Brace())
	assert.Equal(t, 2, len(s.Elems))
}
