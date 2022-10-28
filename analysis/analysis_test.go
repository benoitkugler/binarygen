package analysis

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/types"
	"testing"
)

var ana Analyser

func init() {
	var err error
	ana, err = NewAnalyser("../test-package/defs.go")
	if err != nil {
		panic(err)
	}
}

func (an *Analyser) byName(name string) *types.Named {
	return an.pkg.Types.Scope().Lookup(name).Type().(*types.Named)
}

func (an *Analyser) printExpr(expr ast.Expr) string {
	var buf bytes.Buffer
	format.Node(&buf, an.pkg.Fset, expr)
	return buf.String()
}

func TestParseSource(t *testing.T) {
	if ana.absSourcePath == "" {
		t.Fatal()
	}

	if len(ana.commentsMap) == 0 {
		t.Fatal()
	}

	if len(ana.Sources) == 0 {
		t.Fatal()
	}

	if len(ana.interfaces) == 0 {
		t.Fatal()
	}

	if len(ana.forAliases) == 0 {
		t.Fatal()
	}

	if len(ana.constructors) == 0 {
		t.Fatal()
	}
}

func TestStartingOffset(t *testing.T) {
	if ty := ana.byName("startNoAtSubslice"); ana.commentsMap[ty].startingOffset != "2" {
		t.Fatal()
	}
}

func TestAliases(t *testing.T) {
	if ty := ana.byName("lookup"); ana.printExpr(ana.forAliases[ty]["w"]) != "fl32" {
		t.Fatal()
	}

	u := ana.Tables[ana.byName("lookup")].Fields[9]
	if derived := u.Type.(DerivedFromBasic); derived.Name != "fl32" {
		t.Fatal()
	}
}

func TestInterfaces(t *testing.T) {
	if ty := ana.byName("subtable"); len(ana.interfaces[ty.Underlying().(*types.Interface)]) != 2 {
		t.Fatal()
	}

	u := ana.Tables[ana.byName("withUnion")].Fields[2].Type.(Union)
	if len(u.Flags) != 2 || len(u.Members) != 2 {
		t.Fatal(u)
	}
}

func TestConstructors(t *testing.T) {
	if ana.constructors["fl32"] != types.Typ[types.Uint32] {
		t.Fatal(ana.constructors["fl32"])
	}
}

func TestOffset(t *testing.T) {
	ty := ana.Tables[ana.byName("withOffset")]
	o1 := ty.Fields[1].Type
	o2 := ty.Fields[2].Type
	o3 := ty.Fields[6].Type
	if o1.(Offset).size != Uint32 {
		t.Fatal(o1)
	}
	if o2.(Offset).size != Uint32 {
		t.Fatal(o2)
	}
	if o3.(Offset).size != Uint16 {
		t.Fatal(o3)
	}
}

func TestRawdata(t *testing.T) {
	ty := ana.Tables[ana.byName("complexeSubtable")]

	rawDataField := ty.Fields[len(ty.Fields)-1]
	if rawDataField.Layout.SubsliceStart != AtStart {
		t.Fatal()
	}
	if !rawDataField.Type.(Slice).IsRawData() {
		t.Fatal()
	}

	if sliceField := ty.Fields[3].Type.(Slice); sliceField.IsRawData() {
		t.Fatal()
	}
}
