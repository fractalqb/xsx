package table

import (
	"fmt"
	"strconv"

	"github.com/fractalqb/xsx"
)

type ColType int

const (
	Bool ColType = iota
	Int
	Float
	String
)

func (ct ColType) String() string {
	switch ct {
	case Bool:
		return "bool"
	case Int:
		return "int"
	case Float:
		return "float"
	case String:
		return "string"
	default:
		return "<unsupported type: " + strconv.Itoa(int(ct)) + ">"
	}
}

type Column struct {
	Name string
	Type ColType
	Meta bool
}

type Definition []Column

func ReadDef(xrd *xsx.PullParser) (res Definition, err error) {
	if err = xrd.NextBegin("[", xsx.NoMeta); err != nil {
		return nil, err
	}
	var tok xsx.Token
	for tok, err = xrd.Next(); tok != xsx.TokEnd || xrd.LastBracket() != ']'; tok, err = xrd.Next() {
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
			Type: String,
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
	tok, err := xrd.Next()
	if err != nil {
		return Column{}, err
	}
	typ := String
	switch tok {
	case xsx.TokEnd:
		if xrd.LastBracket() != ')' {
			return Column{}, fmt.Errorf("in column '%s' expected ')', got '%c'",
				name,
				xrd.LastBracket())
		} else {
			return Column{Name: name, Type: String, Meta: metaCol}, nil
		}
	case xsx.TokAtom:
		if xrd.WasMeta() {
			return Column{}, fmt.Errorf("unsupported meta token in column '%s'", name)
		}
		switch xrd.Atom {
		case "bool":
			typ = Bool
		case "int":
			typ = Int
		case "float":
			typ = Float
		case "string":
			typ = String
		default:
			return Column{}, fmt.Errorf("unknown type '%s' for column '%s'",
				xrd.Atom,
				name)
		}
	default:
		return Column{}, fmt.Errorf("unexpected token '%s' in column '%s'", tok, name)
	}
	if err = xrd.NextEnd(")"); err != nil {
		return Column{}, fmt.Errorf("in column '%s' expected ')', got '%c'",
			name,
			xrd.LastBracket())
	}
	return Column{Name: name, Type: typ, Meta: metaCol}, nil
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

func (tdef Definition) NextRow(xrd *xsx.PullParser, row []string) ([]string, error) {
	if row == nil || len(row) < len(tdef) {
		row = make([]string, len(tdef))
	}
	if err := xrd.NextBegin("(", xsx.NoMeta); err == xsx.PullEOI {
		return nil, xsx.PullEOI
	}
	for i := 0; i < len(tdef); i++ {
		atom, err := xrd.NextAtom(xsx.NoMeta)
		if err != nil {
			return row[:i], err
		}
		row[i] = atom
	}
	if err := xrd.NextEnd(")"); err != nil {
		return row, err
	}
	return row, nil
}
