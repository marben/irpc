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
		{in: []byte{127}, want: "[]byte{0x7f}"},
		{in: []byte{0, 255}, want: "[]byte{0x00,0xff}"},
		{in: []byte{17, 128, 0, 0, 0, 0, 0, 0}, want: "[]byte{0x11,0x80,0x00,0x00,0x00,0x00,0x00,0x00}"},
		{in: []byte{17, 128, 0, 0, 0, 0, 0, 0, 99}, want: "[]byte{\n0x11,0x80,0x00,0x00,0x00,0x00,0x00,0x00,\n0x63,\n}"},
	}

	for _, tc := range tests {
		res := byteSliceLiteral(tc.in)
		if tc.want != res {
			t.Fatalf("expected: %q . got: %q", tc.want, res)
		}
	}
}
