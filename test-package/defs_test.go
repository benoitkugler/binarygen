package testpackage

import (
	"bytes"
	"crypto/rand"
	"reflect"
	"testing"
)

func Test(t *testing.T) {
	var buf [1000]byte
	rand.Read(buf[:])

	lk, err := parseLookup(buf[:])
	if err != nil {
		t.Fatal(err)
	}

	got := lk.appendTo(nil)

	origin := buf[:len(got)]
	if !bytes.Equal(origin, got) {
		t.Fatalf("%v !=\n %v", origin, got)
	}

	var ss simpleSubtable
	ss.lookups = make([]lookup, 5)
	for i := range ss.lookups {
		rand.Read(buf[:])
		ss.lookups[i].mustParse(buf[:])
	}
	ss.array2 = []uint32{1, 2, 3}

	out := ss.appendTo(nil)

	ss2, err := parseSimpleSubtable(out)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ss, ss2) {
		t.Fatalf("expected %v, got\n %v", ss, ss2)
	}
}
