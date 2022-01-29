package binarygen

import (
	"fmt"
	"go/format"
	"testing"
)

func Test_importSource(t *testing.T) {
	an, err := importSource("test-package/defs.go")
	if err != nil {
		t.Fatal(err)
	}
	if an.pkgName != "testpackage" {
		t.Fatalf("unexpected package name %s", an.pkgName)
	}

	obj := an.structDefs
	if len(obj) != 9 {
		t.Fatal("unexpected number of structs :", len(obj))
	}
}

func Test_generateParser(t *testing.T) {
	an, err := importSource("test-package/defs.go")
	if err != nil {
		t.Fatal(err)
	}

	code := ""
	for _, v := range an.structDefs {
		st := an.analyzeStruct(v)
		code += st.generateParser() + "\n"
	}

	out, err := format.Source([]byte(code))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(out))
}

func Test_Generate(t *testing.T) {
	err := Generate("test-package/defs.go")
	if err != nil {
		t.Fatal(err)
	}
}
