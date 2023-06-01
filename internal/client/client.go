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

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker-credential-helpers/credentials"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func createCmd(ctx context.Context, commandName string, arg string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, commandName, arg)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	return cmd
}

// isValidCredsMessage checks if 'msg' contains invalid credentials error message.
// It returns whether the logs are free of invalid credentials errors and the error if it isn't.
// error values can be errCredentialsMissingServerURL or errCredentialsMissingUsername.
func isValidCredsMessage(msg string) error {
	if credentials.IsCredentialsMissingServerURLMessage(msg) {
		return credentials.NewErrCredentialsMissingServerURL()
	}
	if credentials.IsCredentialsMissingUsernameMessage(msg) {
		return credentials.NewErrCredentialsMissingUsername()
	}
	return nil
}

// Store uses an external program to save credentials.
func Store(ctx context.Context, binary string, serverAddress string, cred auth.Credential) error {
	dockerCred := &credentials.Credentials{
		ServerURL: serverAddress,
		Username:  cred.Username,
		Secret:    cred.Password,
	}
	if cred.RefreshToken != "" {
		dockerCred.Username = "<token>"
		dockerCred.Secret = cred.RefreshToken
	}

	// should I use *cmd (pointer) here?
	cmd := createCmd(ctx, binary, "store")
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(dockerCred); err != nil {
		return err
	}
	cmd.Stdin = buffer

	out, err := cmd.Output()
	if err != nil {
		t := strings.TrimSpace(string(out))

		if isValidErr := isValidCredsMessage(t); isValidErr != nil {
			err = isValidErr
		}

		return fmt.Errorf("error storing credentials - err: %v, out: `%s`", err, t)
	}

	return nil
}

// Get executes an external program to get the credentials from a native store.
func Get(ctx context.Context, binary string, serverURL string) (*credentials.Credentials, error) {
	cmd := createCmd(ctx, binary, "get")
	cmd.Stdin = strings.NewReader(serverURL)
	out, err := cmd.Output()
	if err != nil {
		t := strings.TrimSpace(string(out))
		if credentials.IsErrCredentialsNotFoundMessage(t) {
			return nil, credentials.NewErrCredentialsNotFound()
		}
		if isValidErr := isValidCredsMessage(t); isValidErr != nil {
			err = isValidErr
		}
		return nil, fmt.Errorf("error getting credentials - err: %v, out: `%s`", err, t)
	}
	resp := &credentials.Credentials{
		ServerURL: serverURL,
	}
	if err := json.NewDecoder(bytes.NewReader(out)).Decode(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Erase executes a program to remove the server credentials from the native store.
func Erase(ctx context.Context, binary string, serverURL string) error {
	cmd := createCmd(ctx, binary, "erase")
	cmd.Stdin = strings.NewReader(serverURL)
	out, err := cmd.Output()
	if err != nil {
		t := strings.TrimSpace(string(out))
		if isValidErr := isValidCredsMessage(t); isValidErr != nil {
			err = isValidErr
		}
		return fmt.Errorf("error erasing credentials - err: %v, out: `%s`", err, t)
	}
	return nil
}

// List executes a program to list server credentials in the native store.
func List(ctx context.Context, binary string) (map[string]string, error) {
	cmd := createCmd(ctx, binary, "list")
	cmd.Stdin = strings.NewReader("unused")
	out, err := cmd.Output()
	if err != nil {
		t := strings.TrimSpace(string(out))
		if isValidErr := isValidCredsMessage(t); isValidErr != nil {
			err = isValidErr
		}
		return nil, fmt.Errorf("error listing credentials - err: %v, out: `%s`", err, t)
	}
	var resp map[string]string
	if err = json.NewDecoder(bytes.NewReader(out)).Decode(&resp); err != nil {
		return nil, err
	}
	return resp, nil
}
