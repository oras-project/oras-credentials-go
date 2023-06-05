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

package credentials

import (
	"context"
	"os/exec"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/oras-project/oras-credentials-go/internal/client"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	remoteCredentialsPrefix = "docker-credential-"
)

// nativeStore implements a credentials store using native keychain to keep
// credentials secure.
type nativeStore struct {
	// helperBinaryName string
	exe client.Executer
}

// NewNativeStore creates a new native store that uses a remote helper program to
// manage credentials.
//
// The argument of NewNativeStore can be the native keychains
// ("wincred" for Windows, "pass" for linux and "osxkeychain" for macOS),
// or any program that follows the docker-credentials-helper protocol.
//
// Reference:
//   - https://docs.docker.com/engine/reference/commandline/login#credentials-store
func NewNativeStore(helperSuffix string) Store {
	// return &nativeStore{
	// 	helperBinaryName: remoteCredentialsPrefix + helperSuffix,
	// }
	return &nativeStore{
		exe: client.NewExecuter(remoteCredentialsPrefix + helperSuffix),
	}
}

// Get retrieves credentials from the store for the given server.
func (ns *nativeStore) Get(ctx context.Context, serverAddress string) (auth.Credential, error) {
	var cred auth.Credential
	// dockerCred, err := client.Get(ctx, ns.helperBinaryName, serverAddress)
	dockerCred, err := client.Get(ctx, ns.exe, serverAddress)
	if err != nil {
		if credentials.IsErrCredentialsNotFound(err) {
			// do not return an error if the credentials are not in the keychain.
			return auth.EmptyCredential, nil
		}
		return auth.EmptyCredential, err
	}
	// bearer auth is used if the username is "<token>"
	if dockerCred.Username == "<token>" {
		cred.RefreshToken = dockerCred.Secret
	} else {
		cred.Username = dockerCred.Username
		cred.Password = dockerCred.Secret
	}
	return cred, nil
}

// Put saves credentials into the store.
func (ns *nativeStore) Put(ctx context.Context, serverAddress string, cred auth.Credential) error {
	// return client.Store(ctx, ns.helperBinaryName, serverAddress, cred)
	return client.Store(ctx, ns.exe, serverAddress, cred)
}

// Delete removes credentials from the store for the given server.
func (ns *nativeStore) Delete(ctx context.Context, serverAddress string) error {
	return client.Erase(ctx, ns.exe, serverAddress)
}

// getDefaultHelperSuffix returns the default credential helper suffix.
func getDefaultHelperSuffix() string {
	platformDefault := getPlatformDefaultHelperSuffix()
	if _, err := exec.LookPath(remoteCredentialsPrefix + platformDefault); err == nil {
		return platformDefault
	}
	return ""
}
