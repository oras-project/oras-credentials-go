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
	"errors"

	"oras.land/oras-go/v2/registry/remote/auth"
)

// FileStore implements a credentials store using the docker configuration file
// to keep the credentials in plain-text.
type FileStore struct {
	// DisablePut disables putting credentials in plaintext.
	// If DisablePut is set to true, Put() will return ErrPlaintextPutDisabled.
	DisablePut bool

	config *config
}

// ErrPlaintextPutDisabled is returned by Put() when DisablePut is set
// to true.
var ErrPlaintextPutDisabled = errors.New("putting plaintext credentials is disabled")

// NewFileStore creates a new file credentials store.
func NewFileStore(configPath string) (*FileStore, error) {
	cfg, err := loadConfigFile(configPath)
	if err != nil {
		return nil, err
	}
	return &FileStore{config: cfg}, nil
}

// Get retrieves credentials from the store for the given server address.
func (fs *FileStore) Get(_ context.Context, serverAddress string) (auth.Credential, error) {
	authCfg, err := fs.config.getAuthConfig(serverAddress)
	if err != nil {
		return auth.EmptyCredential, err
	}
	return authCfg.Credential()
}

// Put saves credentials into the store for the given server address.
// Returns ErrPlaintextPutDisabled if fs.DisablePut is set to true.
func (fs *FileStore) Put(_ context.Context, serverAddress string, cred auth.Credential) error {
	if fs.DisablePut {
		return ErrPlaintextPutDisabled
	}

	authCfg := newAuthConfig(cred)
	return fs.config.putAuthConfig(serverAddress, authCfg)
}

// Delete removes credentials from the store for the given server address.
func (fs *FileStore) Delete(_ context.Context, serverAddress string) error {
	return fs.config.deleteAuthConfig(serverAddress)
}
