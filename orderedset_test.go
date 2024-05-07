package main

import "testing"

func TestOrderedSetInt(t *testing.T) {
	os := newOrderedSet[int]()
	os.add(3)
	os.add(1)
	os.add(1)
	os.add(3)
	os.add(2)
	os.add(2)
	out := os.getAll()
	if len(out) != 3 {
		t.Fatalf("unexpected length: %d", len(out))
	}
	if out[0] != 3 || out[1] != 1 || out[2] != 2 {
		t.Fatalf("unexpected slice: %v", out)
	}
}

func TestOederedSetVariadic(t *testing.T) {
	a := []int{1, 3, 2, 1, 2, 3}
	s := newOrderedSet[int]()
	s.add(a...)
	out := s.getAll()
	if len(out) != 3 ||
		out[0] != 1 ||
		out[1] != 3 ||
		out[2] != 2 {
		t.Fatalf("unexpected output: %v", out)
	}
}
