package table

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"

	"github.com/fractalqb/xsx"
	"github.com/fractalqb/xsx/gem"
)

func pullStr(str string) *xsx.PullParser {
	return xsx.NewPullParser(bufio.NewReader(bytes.NewBufferString(str)))
}

func TestReadDef_empty(t *testing.T) {
	input := pullStr("[]")
	tdef, err := ReadDef(input)
	if err != nil {
		t.Error(err)
	}
	if len(tdef) != 0 {
		t.Error("non-empty table defintiion")
	}
}

func assertColumn(t *testing.T, tdef Definition, col int, nm string, meta bool) {
	if col >= len(tdef) {
		t.Errorf("not enough hcolumns: %d, looking for index %d", len(tdef), col)
		return
	}
	cdef := tdef[col]
	if cdef.Name != nm {
		t.Errorf("col %d: expect name '%s', got '%s'", col, nm, cdef.Name)
	}
	if cdef.Meta != meta {
		t.Errorf("col %d: expect meta=%t, got %t", col, meta, cdef.Meta)
	}
}

func assertTags(t *testing.T, tdef Definition, col int, tags []gem.Expr) {
	cdef := tdef[col]
	if len(cdef.Tags) != len(tags) {
		t.Errorf("col %d: wring number of tags %d, expected %d", col, len(tags), len(cdef.Tags))
	}
	for i := 0; i < len(tags); i++ {
		if !reflect.DeepEqual(tags[i], cdef.Tags[i]) {
			t.Errorf("col %d: unexpected tag '%v', expect '%v'", col, cdef.Tags[i], tags[i])
		}
	}
}

func TestReadDef_1simple(t *testing.T) {
	input := pullStr("[foo]")
	tdef, err := ReadDef(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(tdef) != 1 {
		t.Fatalf("expected 1 column, go %d", len(tdef))
	}
	assertColumn(t, tdef, 0, "foo", false)
}

func TestReadDef_1simple_meta(t *testing.T) {
	input := pullStr("[\\foo]")
	tdef, err := ReadDef(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(tdef) != 1 {
		t.Fatalf("expected 1 column, go %d", len(tdef))
	}
	assertColumn(t, tdef, 0, "foo", true)
}

func TestReadDef_1default(t *testing.T) {
	input := pullStr("[(foo)]")
	tdef, err := ReadDef(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(tdef) != 1 {
		t.Fatalf("expected 1 column, go %d", len(tdef))
	}
	assertColumn(t, tdef, 0, "foo", false)
}

func TestReadDef_1default_meta(t *testing.T) {
	input := pullStr("[\\(foo)]")
	tdef, err := ReadDef(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(tdef) != 1 {
		t.Fatalf("expected 1 column, go %d", len(tdef))
	}
	assertColumn(t, tdef, 0, "foo", true)
}

func TestReadDef_1int(t *testing.T) {
	input := pullStr("[(foo int)]")
	tdef, err := ReadDef(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(tdef) != 1 {
		t.Fatalf("expected 1 column, go %d", len(tdef))
	}
	assertColumn(t, tdef, 0, "foo", false)
	assertTags(t, tdef, 0, []gem.Expr{&gem.Atom{Str: "int"}})
}

func TestReadDef_complex(t *testing.T) {
	input := pullStr("[foo (bar bool) (baz string) (quux float)]")
	tdef, err := ReadDef(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(tdef) != 4 {
		t.Fatalf("expected 1 column, go %d", len(tdef))
	}
	assertColumn(t, tdef, 0, "foo", false)
	assertTags(t, tdef, 0, nil)
	assertColumn(t, tdef, 1, "bar", false)
	assertTags(t, tdef, 1, []gem.Expr{&gem.Atom{Str: "bool"}})
	assertColumn(t, tdef, 2, "baz", false)
	assertTags(t, tdef, 2, []gem.Expr{&gem.Atom{Str: "string"}})
	assertColumn(t, tdef, 3, "quux", false)
	assertTags(t, tdef, 3, []gem.Expr{&gem.Atom{Str: "float"}})
}

func TestNextRow(t *testing.T) {
	input := pullStr(`[(foo int) bar (baz bool)]
	(0 word1 true)
	(1 "word 2" false)`)
	tdef, err := ReadDef(input)
	if err != nil {
		t.Fatal(err)
	}
	row, err := tdef.NextRow(input, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(row) != len(tdef) {
		t.Fatalf("1st row has %d columns, expected %d", len(row), len(tdef))
	}
	if row[0].(*gem.Atom).Str != "0" {
		t.Errorf("col 0: expect '0', got '%s'", row[0])
	}
	if row[1].(*gem.Atom).Str != "word1" {
		t.Errorf("col 1: expect 'word1', got '%s'", row[1])
	}
	if row[2].(*gem.Atom).Str != "true" {
		t.Errorf("col 2: expect 'true', got '%s'", row[2])
	}
	row, err = tdef.NextRow(input, row)
	if err != nil {
		t.Fatal(err)
	}
	if len(row) != len(tdef) {
		t.Fatalf("1st row has %d columns, expected %d", len(row), len(tdef))
	}
	if row[0].(*gem.Atom).Str != "1" {
		t.Errorf("col 0: expect '1', got '%s'", row[0])
	}
	if row[1].(*gem.Atom).Str != "word 2" {
		t.Errorf("col 1: expect 'word 2', got '%s'", row[1])
	}
	if row[2].(*gem.Atom).Str != "false" {
		t.Errorf("col 2: expect 'false', got '%s'", row[2])
	}
	row, err = tdef.NextRow(input, row)
	if row != nil {
		t.Errorf("received unexpected line: %v", row)
	}
	if err != xsx.PullEOI {
		t.Error("expected EOI, got", err)
	}
}
