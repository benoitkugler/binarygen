package analysis

import "testing"

func TestScopes(t *testing.T) {
	l := ana.Tables[ana.byName("lookup")].Scopes()
	if len(l) != 1 {
		t.Fatal(l)
	}

	l = ana.Tables[ana.byName("simpleSubtable")].Scopes()
	if len(l) != 3 {
		t.Fatal(l)
	}
}
