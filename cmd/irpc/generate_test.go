package main

import "testing"

func TestGenerateNewFuncName(t *testing.T) {
	type test struct {
		in   string
		want string
	}

	tests := []test{
		{in: "abc", want: "newAbc"},
		{in: "a", want: "newA"},
		{in: "Abc", want: "NewAbc"},
		{in: "čBc", want: "newČBc"},
		{in: "_čBc", want: "new_čBc"},
	}

	for _, tc := range tests {
		got := generateStructConstructorName(tc.in)
		if got != tc.want {
			t.Fatalf("input: '%s'. got: '%s'. want: '%s'", tc.in, got, tc.want)
		}
	}
}
