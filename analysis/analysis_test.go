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
	ana, err = NewAnalyser("../test-package/source_src.go")
	if err != nil {
		panic(err)
	}
}

func (an *Analyser) printExpr(expr ast.Expr) string {
	var buf bytes.Buffer
	format.Node(&buf, an.pkg.Fset, expr)
	return buf.String()
}

func TestParseSource(t *testing.T) {
	if ana.sourceAbsPath == "" {
		t.Fatal()
	}

	if len(ana.commentsMap) == 0 {
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

	if len(ana.fetchSource()) == 0 {
		t.Fatal()
	}
}

func TestStartingOffset(t *testing.T) {
	ty := ana.ByName("startNoAtSubslice")
	if ana.commentsMap[ty].startingOffset != "2" {
		t.Fatal()
	}

	if u := ana.Tables[ty]; u.StartingOffset != 2 {
		t.Fatal()
	}
}

func TestAliases(t *testing.T) {
	if ty := ana.ByName("withAlias"); ana.printExpr(ana.forAliases[ty]["f"]) != "fl32" {
		t.Fatal()
	}

	u := ana.Tables[ana.ByName("withAlias")].Fields[0]
	if derived := u.Type.(DerivedFromBasic); derived.Name != "fl32" {
		t.Fatal()
	}
}

func TestOpaque(t *testing.T) {
	fi := ana.Tables[ana.ByName("WithOpaque")].Fields[1]
	if _, isOpaque := fi.Type.(Opaque); !isOpaque {
		t.Fatal()
	}
}

func TestInterfaces(t *testing.T) {
	if ty := ana.ByName("subtableITF"); len(ana.interfaces[ty.Underlying().(*types.Interface)]) != 2 {
		t.Fatal()
	}

	u := ana.Tables[ana.ByName("WithUnion")].Fields[2].Type.(Union)
	if len(u.UnionTag.(UnionTagExplicit).Flags) != 2 || len(u.Members) != 2 {
		t.Fatal(u)
	}
}

func TestConstructors(t *testing.T) {
	if ana.constructors["fl32"] != types.Typ[types.Uint32] {
		t.Fatal(ana.constructors["fl32"])
	}
}

func TestOffset(t *testing.T) {
	ty := ana.Tables[ana.ByName("WithOffset")]
	o1 := ty.Fields[1].Type
	o2 := ty.Fields[2].Type
	o3 := ty.Fields[6].Type
	if o1.(Offset).Size != Uint32 {
		t.Fatal(o1)
	}
	if o2.(Offset).Size != Uint32 {
		t.Fatal(o2)
	}
	if o3.(Offset).Size != Uint16 {
		t.Fatal(o3)
	}
}

func TestRawdata(t *testing.T) {
	ty := ana.Tables[ana.ByName("WithRawdata")]

	for _, fi := range ty.Fields[1:] {
		if !fi.Type.(Slice).IsRawData() {
			t.Fatal()
		}
	}

	startTo := ty.Fields[2]
	if startTo.Layout.SubsliceStart != AtStart {
		t.Fatal()
	}

	startToEnd := ty.Fields[4]
	if startToEnd.Type.(Slice).Count != ToEnd {
		t.Fatal()
	}

	startToOffset := ty.Fields[5]
	if startToOffset.Type.(Slice).Count != ToComputedField {
		t.Fatal()
	}
}

func TestExternalTypes(t *testing.T) {
	ty := ana.Tables[ana.ByName("withFromExternalFile")]
	ref := ty.Fields[0].Type.Origin().(*types.Named)
	if _, hasRef := ana.Tables[ref]; !hasRef {
		t.Fatalf("missing reference to %s", ref)
	}
}

func TestArray(t *testing.T) {
	ty := ana.Tables[ana.ByName("WithArray")]
	// a uint16
	// b [4]uint32
	// c [3]byte
	if size, _ := ty.IsFixedSize(); size != 2+4*4+3*1 {
		t.Fatal()
	}
}

func TestOffsetsArray(t *testing.T) {
	ty := ana.Tables[ana.ByName("WithOffsetArray")]
	sl := ty.Fields[0].Type.(Slice)
	offset, isOffset := sl.Elem.(Offset)
	if !isOffset {
		t.Fatalf("%T", sl.Elem)
	}
	if _, isStruct := offset.Target.(Struct); !isStruct {
		t.Fatalf("%T", offset.Target)
	}
}

func TestExternalArguments(t *testing.T) {
	ty := ana.Tables[ana.ByName("withArgument")]
	if len(ty.Arguments) != 2 {
		t.Fatal(ty.Arguments)
	}
}

func TestImplicitITF(t *testing.T) {
	ty := ana.Tables[ana.ByName("WithImplicitITF")]
	unionScheme, ok := ty.Fields[1].Type.(Union).UnionTag.(UnionTagImplicit)
	if !ok {
		t.Fatal()
	}
	if size, ok := unionScheme.Tag.IsFixedSize(); !ok || size != Uint16 {
		t.Fatal()
	}

	if len(ana.StandaloneUnions) != 1 {
		t.Fatal()
	}
}
