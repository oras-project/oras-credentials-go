/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package executer is an abstraction for the docker credential helper protocol
// binaries. It is used by nativeStore to interact with installed binaries.
package executer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
)

// Executer is an interface that simulates an executable binary.
type Executer interface {
	Execute(ctx context.Context, input io.Reader, action string) ([]byte, error)
}

// executable implements the Executer interface.
type executable struct {
	name string
}

// New returns a new Executer instance.
func New(name string) Executer {
	return &executable{
		name: name,
	}
}

// Execute operates on an executable binary and supports context.
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
