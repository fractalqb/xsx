package table

import (
	"errors"
	"fmt"

	"git.fractalqb.de/fractalqb/xsx"
	"git.fractalqb.de/fractalqb/xsx/gem"
)

type Column struct {
	Name string
	Meta bool
	Tags []gem.Expr
}

type Definition []Column

func ReadDef(xrd *xsx.PullParser) (res Definition, err error) {
	if err = xrd.NextBegin("[", xsx.NoMeta); err != nil {
		return nil, err
	}
	var tok xsx.Token
	for tok, err = xrd.Next(); tok != xsx.TokEnd || xrd.LastBrace() != ']'; tok, err = xrd.Next() {
		if err != nil {
			return res, err
		}
		if col, err := readCol(xrd); err != nil {
			return res, err
		} else {
			res = append(res, col)
		}
	}
	return res, nil
}

func readCol(xrd *xsx.PullParser) (res Column, err error) {
	if xrd.LastToken() == xsx.TokAtom {
		res = Column{
			Name: xrd.Atom,
			Meta: xrd.WasMeta()}
		return res, nil
	}
	if err = xrd.ExpectBegin("(", xsx.AllowMeta); err != nil {
		return Column{}, err
	}
	metaCol := xrd.WasMeta()
	name, err := xrd.NextAtom(xsx.NoMeta)
	if err != nil {
		return Column{}, err
	}
	var tags []gem.Expr
	for tok, err := xrd.Next(); tok != xsx.TokEnd; tok, err = xrd.Next() {
		switch {
		case err != nil:
			return Column{}, err
		case tok == xsx.TokEOI:
			return Column{}, errors.New("premature end of input")
		}
		tag, err := gem.ReadCurrent(xrd)
		if err != nil {
			return Column{}, err
		}
		tags = append(tags, tag)
	}
	if err = xrd.ExpectEnd(")"); err != nil {
		return Column{}, fmt.Errorf("in column '%s' expected ')', got '%c'",
			name,
			xrd.LastBrace())
	}
	return Column{Name: name, Tags: tags, Meta: metaCol}, nil
}

// ColIndex returns the index of the column with name 'colName', if any.
// Otherwise it returns -1.
func (tdef Definition) ColIndex(colName string) int {
	for i, c := range tdef {
		if c.Name == colName {
			return i
		}
	}
	return -1
}

func (tdef Definition) NextRow(xrd *xsx.PullParser, row []gem.Expr) ([]gem.Expr, error) {
	if row == nil || len(row) < len(tdef) || 3*len(row) < cap(row) {
		row = make([]gem.Expr, len(tdef))
	}
	for {
		if err := xrd.NextBegin("(", xsx.AllowMeta); err == xsx.PullEOI {
			return nil, xsx.PullEOI
		} else if err != nil {
			return nil, err
		}
		if xrd.WasMeta() {
			if err := xrd.SkipMeta(); err != nil {
				return nil, err
			}
		} else {
			break
		}
	}
	for i := 0; i < len(tdef); i++ {
		elem, err := gem.ReadNext(xrd)
		if err != nil {
			return row[:i], err
		}
		row[i] = elem
	}
	if err := xrd.NextEnd(")"); err != nil {
		return row, err
	}
	return row, nil
}
