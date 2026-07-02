package main

import (
	"io"
	"os/exec"
)

type Process struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func (p *Process) Read(b []byte) (int, error) {
	return p.stdout.Read(b)
}

func (p *Process) Write(b []byte) (int, error) {
	return p.stdin.Write(b)
}

func (p *Process) Close() error {
	// stdin を閉じて終了を促す
	_ = p.stdin.Close()

	// プロセス終了待ち
	return p.cmd.Wait()
}

func (p *Process) Stderr() io.Reader {
	return p.stderr
}

func Proc(command string, args ...string) (*Process, error) {
	cmd := exec.Command(command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, err
	}

	return &Process{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}, nil
}
