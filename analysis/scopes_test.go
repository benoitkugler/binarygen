package analysis

import "testing"

func TestScopes(t *testing.T) {
	ta := ana.Tables[ana.byName("singleScope")]
	if _, is := ta.IsFixedSize(); !is {
		t.Fatal()
	}
	if len(ta.Scopes()) != 1 {
		t.Fatal(ta.Scopes())
	}

	l := ana.Tables[ana.byName("multipleScopes")].Scopes()
	if len(l) != 3 {
		t.Fatal(l)
	}
}
