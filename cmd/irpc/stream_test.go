package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// generateFromSource runs the generator on a single file in a temporary module
func generateFromSource(t *testing.T, src string) error {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/streamtest\n\ngo 1.24\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "api.go")
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := newGenerator(file)
	return err
}

func TestStreamValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		wantErr []string
	}{
		{
			name:    "send-only parameter",
			method:  "In(ctx context.Context, a <-chan int, b chan<- int) error",
			wantErr: []string{"Api.In", `parameter "b" has type chan<- int`, "only receive-only channels (<-chan T)"},
		},
		{
			name:    "bidirectional parameter",
			method:  "In(ctx context.Context, a chan int) error",
			wantErr: []string{"Api.In", `parameter "a" has type chan int`, "use <-chan int"},
		},
		{
			name:    "send-only result",
			method:  "Out(ctx context.Context) (chan<- int, error)",
			wantErr: []string{"Api.Out", "result 0 has type chan<- int", "only receive-only channels (<-chan T) can be returned"},
		},
		{
			name:    "bidirectional result",
			method:  "Out(ctx context.Context) (chan int, error)",
			wantErr: []string{"Api.Out", "result 0 has type chan int", "return <-chan int"},
		},
		{
			name:    "missing context",
			method:  "In(a <-chan int) error",
			wantErr: []string{"Api.In", "must have a context.Context parameter"},
		},
		{
			name:    "missing error",
			method:  "In(ctx context.Context, a <-chan int) int",
			wantErr: []string{"Api.In", "must return error as the last result"},
		},
		{
			name:    "channel of channels",
			method:  "In(ctx context.Context, a <-chan chan int) error",
			wantErr: []string{"channel of channels"},
		},
		{
			name:    "channel inside a slice",
			method:  "In(ctx context.Context, a []<-chan int) error",
			wantErr: []string{"channels are supported only", "directly as parameters or results"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src := "package api\n\nimport \"context\"\n\ntype Api interface {\n\t" + tc.method + "\n}\n"
			err := generateFromSource(t, src)
			if err == nil {
				t.Fatalf("expected an error for %q", tc.method)
			}
			for _, want := range tc.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error doesn't contain %q:\n%v", want, err)
				}
			}
		})
	}
}

// several channels in the same method, in both directions, are fine
func TestStreamValidationOk(t *testing.T) {
	src := `package api

import "context"

type Api interface {
	Many(ctx context.Context, a <-chan int, b <-chan string, n int) (x int, c <-chan bool, d <-chan []byte, err error)
}
`
	if err := generateFromSource(t, src); err != nil {
		t.Fatal(err)
	}
}
