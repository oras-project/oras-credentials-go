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

	"github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	remoteCredentialsPrefix = "docker-credential-"
	emptyUsername           = "<token>"
)

// nativeAuthStore implements a credentials store using native keychain to keep
// credentials secure.
type NativeStore struct {
	programFunc client.ProgramFunc
}

// NewNativeStore creates a new native store that uses a remote helper program to
// manage credentials.
func NewNativeStore(helperSuffix string) Store {
	return &NativeStore{
		programFunc: client.NewShellProgramFunc(remoteCredentialsPrefix + helperSuffix),
	}
}

// Get retrieves credentials from the store for the given server
func (ns *NativeStore) Get(_ context.Context, serverAddress string) (auth.Credential, error) {
	var cred auth.Credential
	dockerCred, err := client.Get(ns.programFunc, serverAddress)
	if err != nil {
		if credentials.IsErrCredentialsNotFound(err) {
			// do not return an error if the credentials are not in the keychain.
			return auth.EmptyCredential, nil
		}
		return auth.EmptyCredential, err
	}
	// bearer auth is used if the username is emptyUsername
	if dockerCred.Username == emptyUsername {
		cred.RefreshToken = dockerCred.Secret
	} else {
		cred.Username = dockerCred.Username
		cred.Password = dockerCred.Secret
	}
	return cred, nil
}

// Put saves credentials into the store
func (ns *NativeStore) Put(_ context.Context, serverAddress string, cred auth.Credential) error {
	dockerCred := &credentials.Credentials{
		ServerURL: serverAddress,
		Username:  cred.Username,
		Secret:    cred.Password,
	}
	if cred.RefreshToken != "" {
		dockerCred.Username = emptyUsername
		dockerCred.Secret = cred.RefreshToken
	}
	return client.Store(ns.programFunc, dockerCred)
}

// Delete removes credentials from the store for the given server
func (ns *NativeStore) Delete(_ context.Context, serverAddress string) error {
	return client.Erase(ns.programFunc, serverAddress)
}
