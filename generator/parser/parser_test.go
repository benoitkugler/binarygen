package parser

import (
	"fmt"
	pa "go/parser"
	"go/token"
	"testing"

	"github.com/benoitkugler/binarygen/analysis"
	gen "github.com/benoitkugler/binarygen/generator"
)

func assertParseBlock(t *testing.T, code string) {
	code = fmt.Sprintf(`package main 
	func main() {
		%s
	}`, code)
	_, err := pa.ParseFile(token.NewFileSet(), "", code, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_staticLengthCheck_code(t *testing.T) {
	cc := gen.Context{
		Type:  "lookup",
		Slice: "data",
	}
	tests := []analysis.BinarySize{1, 5, 12}
	for _, sl := range tests {
		got := staticLengthCheckAt(cc, sl)
		assertParseBlock(t, got)
	}
}

func Test_affineLengthCheck_code(t *testing.T) {
	cc := gen.Context{
		Type:   "lookup",
		Slice:  "data",
		Offset: gen.NewOffset("n", 0),
	}
	tests := []struct {
		count gen.Expression
		size  analysis.BinarySize
	}{
		{"L2", 4},
		{"L2", 5},
	}
	for _, sl := range tests {
		got := affineLengthCheckAt(cc, sl.count, sl.size)
		assertParseBlock(t, got)
	}
}

func Test_conditionalLengthCheck_code(t *testing.T) {
	cc := gen.Context{
		Type:  "lookup",
		Slice: "data",
	}
	tests := []conditionalLength{
		{
			"2",
			[]conditionalField{},
		},
		{
			"n",
			[]conditionalField{
				{"lookup", 4},
				{"name", 2},
			},
		},
	}
	for _, sl := range tests {
		got := conditionalLengthCheck(sl, cc)
		assertParseBlock(t, got)
	}
}

func TestCodeForLength(t *testing.T) {
	cc := gen.Context{
		Type:      "lookup",
		Slice:     "data",
		ObjectVar: "item",
		Offset:    gen.NewOffset("n", 0),
	}

	for _, ct := range [...]analysis.ArrayCount{
		analysis.FirstUint16, analysis.FirstUint32, analysis.ComputedField,
		analysis.ToEnd,
	} {
		_, code := codeForSliceCount(analysis.Slice{
			Count:     ct,
			CountExpr: "myVar",
		}, "dummy", &cc)
		assertParseBlock(t, code)
	}
}
