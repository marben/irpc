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

func TestByteSliceLiteral(t *testing.T) {
	type test struct {
		in   []byte
		want string
	}
	tests := []test{
		{in: nil, want: "[]byte{}"},
		{in: []byte{127}, want: "[]byte{127}"},
		{in: []byte{0, 255}, want: "[]byte{0, 255}"},
		{in: []byte{17, 128, 0}, want: "[]byte{17, 128, 0}"},
	}

	for _, tc := range tests {
		res := byteSliceLiteral(tc.in)
		if tc.want != res {
			t.Fatalf("expected: %q . got: %q", tc.want, res)
		}
	}
}
