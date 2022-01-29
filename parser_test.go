package binarygen

import (
	"fmt"
	"go/parser"
	"go/token"
	"testing"
)

func assertParseBlock(t *testing.T, code string) {
	code = fmt.Sprintf(`package main 
	func main() {
		%s
	}`, code)
	_, err := parser.ParseFile(token.NewFileSet(), "", code, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_staticLengthCheck_code(t *testing.T) {
	cc := codeContext{
		typeName:      "lookup",
		byteSliceName: "data",
	}
	tests := []int{1, 5, 12}
	for _, sl := range tests {
		got := staticLengthCheck(sl, cc)
		assertParseBlock(t, got)
	}
}

func Test_affineLengthCheck_code(t *testing.T) {
	cc := codeContext{
		typeName:      "lookup",
		byteSliceName: "data",
	}
	tests := []affine{
		{"", "L2", 4},
		{"offset", "", 0},
		{"2", "L2", 5},
	}
	for _, sl := range tests {
		got := affineLengthCheck(sl, cc)
		assertParseBlock(t, got)
	}
}

func Test_conditionalLengthCheck_code(t *testing.T) {
	cc := codeContext{
		typeName:      "lookup",
		byteSliceName: "data",
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
