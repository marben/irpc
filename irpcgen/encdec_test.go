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
	var r2 map[int]string = make(map[int]string)
	if err := DecMap(dec, &r2, "int", DecInt, "string", DecString); err != nil {
		t.Fatalf("DecMap: %+v", err)
	}

	if r2 != nil {
		t.Fatalf("encoded nil map, but got: %v", r2)
	}
}

func TestEncDecPointer(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	dec := NewDecoder(buf)

	var val int = 2
	p1 := &val

	if err := EncPointer(enc, p1, "int", EncInt); err != nil {
		t.Fatalf("encode *int: %+v", err)
	}
	enc.Flush()

	var p1_rtn *int
	if err := DecPointer(dec, &p1_rtn, "int", DecInt); err != nil {
		t.Fatalf("decode *int: %+v", err)
	}

	if *p1 != *p1_rtn {
		t.Fatalf("%d != %d", *p1, *p1_rtn)
	}
}

func TestEncDecPreinitializedPointer(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	dec := NewDecoder(buf)

	// nil pointer
	var val int = 2
	var p *int = nil
	if err := EncPointer(enc, p, "int", EncInt); err != nil {
		t.Fatalf("encode second *int: %+v", err)
	}
	enc.Flush()

	// pre-initialize to check if the Decoder corretly sets it to nil
	var p_rtn = &val
	if err := DecPointer(dec, &p_rtn, "int", DecInt); err != nil {
		t.Fatalf("decode  *int: %v", err)
	}

	if p_rtn != nil {
		t.Fatalf("rtn was supposed to be nil, but is: %v", p_rtn)
	}
	t.Logf("val == %d", val)
}
