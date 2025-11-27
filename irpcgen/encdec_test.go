package irpcgen

import (
	"bytes"
	"maps"
	"testing"
)

func TestEncDecMap(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	dec := NewDecoder(buf)

	m1 := map[int]string{1: "hhoho", 2: "ohloho"}

	if err := EncMap(enc, m1, "int", EncInt, "string", EncString); err != nil {
		t.Fatalf("EncMap: %+v", err)
	}
	enc.Flush()

	var r1 map[int]string
	if err := DecMap(dec, &r1, "int", DecInt, "string", DecString); err != nil {
		t.Fatalf("DecMap: %+v", err)
	}

	if !maps.Equal(m1, r1) {
		t.Fatalf("m1 != r1: %v != %v", m1, r1)
	}

	// nil map
	var m2 map[int]string
	if err := EncMap(enc, m2, "int", EncInt, "string", EncString); err != nil {
		t.Fatalf("EncMap: %+v", err)
	}
	enc.Flush()
	var r2 map[int]string
	if err := DecMap(dec, &r2, "int", DecInt, "string", DecString); err != nil {
		t.Fatalf("DecMap: %+v", err)
	}

	if r2 != nil {
		t.Fatalf("encoed nil map, but got: %v", r2)
	}
}
