package analysis

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/types"
	"testing"
)

var an Analyser

func init() {
	var err error
	an, err = NewAnalyser("../test-package/defs.go")
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
	if an.sourcePath == "" {
		t.Fatal()
	}

	if len(an.commentsMap) == 0 {
		t.Fatal()
	}

	if len(an.Sources) == 0 {
		t.Fatal()
	}

	if len(an.interfaces) == 0 {
		t.Fatal()
	}

	if len(an.forAliases) == 0 {
		t.Fatal()
	}

	if len(an.constructors) == 0 {
		t.Fatal()
	}
}

func TestStartingOffset(t *testing.T) {
	if ty := an.byName("startNoAtSubslice"); an.commentsMap[ty].startingOffset != "2" {
		t.Fatal()
	}
}

func TestAliases(t *testing.T) {
	if ty := an.byName("lookup"); an.printExpr(an.forAliases[ty]["w"]) != "fl32" {
		t.Fatal()
	}

	u := an.Tables[an.byName("lookup")].Fields[9]
	if derived := u.Type.(DerivedFromBasic); derived.name != "fl32" {
		t.Fatal()
	}
}

func TestInterfaces(t *testing.T) {
	if ty := an.byName("subtable"); len(an.interfaces[ty.Underlying().(*types.Interface)]) != 2 {
		t.Fatal()
	}

	u := an.Tables[an.byName("withUnion")].Fields[2].Type.(Union)
	if len(u.Flags) != 2 || len(u.Members) != 2 {
		t.Fatal(u)
	}
}

func TestConstructors(t *testing.T) {
	if an.constructors["fl32"] != types.Typ[types.Uint32] {
		t.Fatal(an.constructors["fl32"])
	}
}

func TestOffset(t *testing.T) {
	ty := an.Tables[an.byName("withOffset")]
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
	ty := an.Tables[an.byName("complexeSubtable")]

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
