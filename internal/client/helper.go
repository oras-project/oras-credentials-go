package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
)

type Executer interface {
	Execute(ctx context.Context, input io.Reader, action string) ([]byte, error)
}

type executable struct {
	name string
}

func NewExecuter(name string) Executer {
	return &executable{
		name: name,
	}
}

func (c *executable) Execute(ctx context.Context, input io.Reader, action string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, c.name, action)
	cmd.Stdin = input
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			if errMessage := string(bytes.TrimSpace(output)); errMessage != "" {
				err = errors.New(errMessage)
			}
		}
		return nil, err
	}
	return output, nil
}
