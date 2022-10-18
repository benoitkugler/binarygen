package analysis

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/types"
	"testing"
)

func (an *Analyser) byName(name string) *types.Named {
	return an.pkg.Types.Scope().Lookup(name).Type().(*types.Named)
}

func (an *Analyser) printExpr(expr ast.Expr) string {
	var buf bytes.Buffer
	format.Node(&buf, an.pkg.Fset, expr)
	return buf.String()
}

func TestParseSource(t *testing.T) {
	an, err := NewAnalyser("../test-package/defs.go")
	if err != nil {
		t.Fatal(err)
	}

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

	if ty := an.byName("startNoAtSubslice"); an.commentsMap[ty].startingOffset != "2" {
		t.Fatal()
	}

	if ty := an.byName("lookup"); an.printExpr(an.forAliases[ty]["w"]) != "fl32" {
		t.Fatal()
	}

	if ty := an.byName("subtable"); len(an.interfaces[ty.Underlying().(*types.Interface)]) != 2 {
		t.Fatal()
	}
}
