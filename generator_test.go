package binarygen

import (
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

	obj := an.structs
	if len(obj) != 5 {
		t.Fatal("unexpected number of structs :", len(obj))
	}
}

func Test_generateParser(t *testing.T) {
	err := Generate("test-package/defs.go")
	if err != nil {
		t.Fatal(err)
	}
}
