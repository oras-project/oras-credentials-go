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

package executer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
)

// doc please
type Executer interface {
	Execute(ctx context.Context, input io.Reader, action string) ([]byte, error)
}

// doc please
type executable struct {
	name string
}

// doc please
func NewExecuter(name string) Executer {
	return &executable{
		name: name,
	}
}

// doc please
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
