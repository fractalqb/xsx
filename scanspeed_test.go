package xsx

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"testing"
)

type GenXsx interface {
	Next() (c rune, last bool)
}

func ForEach(gen GenXsx, f func(c rune)) {
	r, last := gen.Next()
	f(r)
	for !last {
		r, last = gen.Next()
		f(r)
	}
}

type GenAtom struct {
	meta  bool
	quote int
	len   int
	esced rune
}

func Atom() *GenAtom {
	res := &GenAtom{
		meta: rand.Intn(3) == 0,
		len:  rand.Intn(12),
	}
	if res.len == 0 || rand.Intn(2) > 0 {
		res.quote = 2
	}
	return res
}

var alphaNoq = bytes.Runes([]byte("0123456789abcdefghijklmnopqrstuvwxyz-_"))
var alphaQuo = bytes.Runes([]byte("\\\" 0123456789abcdefghijklmnopqrstuvwxyz-_"))

func (g *GenAtom) Next() (r rune, last bool) {
	if g.meta {
		g.meta = false
		return '\\', false
	}
	if g.quote > 1 {
		g.quote = 1
		return '"', false
	}
	if g.len <= 0 {
		if g.quote > 0 {
			g.quote = 0
			return '"', true
		}
		panic("requesting from exhausted atom generator")
	}
	if g.esced != 0 {
		r = g.esced
		g.esced = 0
		g.len--
		return r, g.len <= 0
	}
	if g.quote > 0 {
		r = alphaQuo[rand.Intn(len(alphaQuo))]
		switch r {
		case '"', '\\':
			g.esced = r
			return '\\', false
		default:
			g.len--
			return r, false
		}
	} else {
		r = alphaNoq[rand.Intn(len(alphaNoq))]
		g.len--
		return r, g.len <= 0
	}
}

type GenSeq struct {
	meta  bool
	depth int
	o, c  rune
	elm   int
	sepws int
	gelm  GenXsx
}

func Seq(maxDepth int) *GenSeq {
	res := &GenSeq{
		meta: rand.Intn(4) == 0,
		elm:  rand.Intn(24),
	}
	if maxDepth <= 1 {
		res.depth = 0
	} else {
		res.depth = rand.Intn(maxDepth)
	}
	switch rand.Intn(3) {
	case 0:
		res.o, res.c = '(', ')'
	case 1:
		res.o, res.c = '[', ']'
	case 2:
		res.o, res.c = '{', '}'
	}
	return res
}

func (g *GenSeq) Next() (r rune, last bool) {
	if g.meta {
		g.meta = false
		return '\\', false
	}
	if g.o != 0 {
		r = g.o
		g.o = 0
		g.sepws = rand.Intn(2)
		return r, false
	}
	if g.sepws > 0 {
		g.sepws--
		return ' ', false
	}
	if g.gelm != nil {
		r, last = g.gelm.Next()
		if last {
			g.gelm = nil
			g.elm--
			g.sepws = 1 + rand.Intn(3)
		}
		return r, false
	}
	if g.elm <= 0 {
		return g.c, true
	} else {
		if g.depth > 0 {
			switch rand.Intn(2) {
			case 0:
				g.gelm = Atom()
			case 1:
				g.gelm = Seq(g.depth - 1)
			}
		} else {
			g.gelm = Atom()
		}
		r, last = g.gelm.Next()
		if last {
			g.gelm = nil
			g.elm--
			g.sepws = 1 + rand.Intn(3)
		}
		return r, false
	}
}

func BenchmarkScanner_noWsBuf(b *testing.B) {
	buf := bytes.NewBuffer(nil)
	rand.Seed(4712)
	gen := Seq(5)
	ForEach(gen, func(c rune) {
		buf.WriteRune(c)
	})
	gen = nil
	txt := buf.Bytes()
	buf = nil
	runtime.GC()
	scn := NewScanner(BeginNop, EndNop, AtomNop)
	//os.Stderr.Write(txt)
	fmt.Fprintf(os.Stderr, "message size: %d x %d\n", b.N, len(txt))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scn.Read(bytes.NewReader(txt))
		scn.Reset()
	}
}

func BenchmarkScanner_withWsBuf(b *testing.B) {
	buf := bytes.NewBuffer(nil)
	rand.Seed(4712)
	gen := Seq(5)
	ForEach(gen, func(c rune) {
		buf.WriteRune(c)
	})
	gen = nil
	txt := buf.Bytes()
	buf = nil
	runtime.GC()
	scn := NewScanner(BeginNop, EndNop, AtomNop)
	scn.WsBuf = &bytes.Buffer{}
	//os.Stderr.Write(txt)
	fmt.Fprintf(os.Stderr, "message size: %d x %d\n", b.N, len(txt))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scn.Read(bytes.NewReader(txt))
		scn.Reset()
	}
}
