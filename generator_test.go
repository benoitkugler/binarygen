package binarygen

import (
	"fmt"
	"go/format"
	"testing"
	"time"
)

func Test_importSource(t *testing.T) {
	ti := time.Now()
	an, err := importSource("test-package/defs.go")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("code loaded in %s\n", time.Since(ti))

	if an.pkgName != "testpackage" {
		t.Fatalf("unexpected package name %s", an.pkgName)
	}

	an.performAnalysis()

	obj := an.structDefs
	if len(obj) != 13 {
		t.Fatal("unexpected number of struct definitions:", len(obj))
	}

	if layouts := an.structLayouts; len(layouts) != len(obj) {
		t.Fatal("unexpected number of struct layouts:", len(layouts))
	}
}

func Test_generateParser(t *testing.T) {
	an, err := importSource("test-package/defs.go")
	if err != nil {
		t.Fatal(err)
	}

	an.performAnalysis()

	buffer := newDeclarationBuffer()
	for _, st := range an.structLayouts {
		for _, decl := range st.generateParser() {
			buffer.add(decl)
		}
	}

	code := buffer.code()

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
