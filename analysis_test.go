package binarygen

import (
	"testing"
)

func TestAlias(t *testing.T) {
	an, err := importSource("test-package/defs.go")
	if err != nil {
		t.Fatal(err)
	}

	an.performAnalysis()
	if an.structLayouts["lookup"].fields[9].type_.(withConstructor).name_ != "fl32" {
		t.Fatal()
	}
}

func TestInterfaces(t *testing.T) {
	an, err := importSource("test-package/defs.go")
	if err != nil {
		t.Fatal(err)
	}

	an.performAnalysis()

	u := an.structLayouts["withUnion"].fields[2].type_.(union)
	if len(u.flags) != 2 || len(u.members) != 2 {
		t.Fatal(u.flags, u.members)
	}
}
