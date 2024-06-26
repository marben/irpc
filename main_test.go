package main

import "testing"

func TestGeneratedFileName(t *testing.T) {
	out, err := generatedFileName("test.go")
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if out != "test_irpc.go" {
		t.Fatalf("unexpected generated filename: '%s'", out)
	}
}
