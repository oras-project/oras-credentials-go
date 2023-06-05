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
	"strings"

	"github.com/docker/docker-credential-helpers/credentials"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// Store uses an external program to save credentials.
func Store(ctx context.Context, exe Executer, serverAddress string, cred auth.Credential) error {
	dockerCred := &credentials.Credentials{
		ServerURL: serverAddress,
		Username:  cred.Username,
		Secret:    cred.Password,
	}
	if cred.RefreshToken != "" {
		dockerCred.Username = "<token>"
		dockerCred.Secret = cred.RefreshToken
	}
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(dockerCred); err != nil {
		return err
	}
	exe.Execute(ctx, buffer, "store")
	return nil
}

// Get executes an external program to get the credentials from a native store.
func Get(ctx context.Context, exe Executer, serverURL string) (*credentials.Credentials, error) {
	out, err := exe.Execute(ctx, strings.NewReader(serverURL), "get")
	if err != nil {
		return nil, err // some handling needed
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
func Erase(ctx context.Context, exe Executer, serverURL string) error {
	exe.Execute(ctx, strings.NewReader(serverURL), "erase")
	return nil
}
