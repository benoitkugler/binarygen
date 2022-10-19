package analysis

import "testing"

func TestScopes(t *testing.T) {
	l := an.Tables[an.byName("lookup")].Scopes()
	if len(l) != 1 {
		t.Fatal(l)
	}

	l = an.Tables[an.byName("simpleSubtable")].Scopes()
	if len(l) != 3 {
		t.Fatal(l)
	}
}
