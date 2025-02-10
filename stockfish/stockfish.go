package stockfish

import (
	"fmt"
	"io"
	"os/exec"
)

type Engine struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func NewEngine() (*Engine, error) {
	cmd := exec.Command("stockfish")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start stockfish: %w", err)
	}

	return &Engine{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
	}, nil
}

func (e *Engine) Close() error {
	if err := e.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill stockfish process: %w", err)
	}
	return e.cmd.Wait()
}
