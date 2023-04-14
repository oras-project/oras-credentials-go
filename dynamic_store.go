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

	"oras.land/oras-go/v2/registry/remote/auth"
)

// credentialConfig contains the config fields related to credentials.
// Reference: https://github.com/docker/cli/blob/v24.0.0-beta.1/cli/config/configfile/file.go#L28-L29
type credentialConfig struct {
	CredentialsStore  string            `json:"credsStore,omitempty"`
	CredentialHelpers map[string]string `json:"credHelpers,omitempty"`
}

// dynamicStore dynamically determines which store to use based on the settings
// in the config file.
type dynamicStore struct {
	config  *config
	options StoreOptions
}

// StoreOptions provides options for NewStore.
type StoreOptions struct {
	// AllowPlaintextPut allows saving credentials in plaintext in the config
	// file.
	//   - If AllowPlaintextPut is set to false (default value), Put() will
	//     return an error when native store is not available.
	//   - If AllowPlaintextPut is set to true, Put() will save credentials in
	//     plaintext in the config file when native store is not available.
	AllowPlaintextPut bool
}

// NewStore returns a store based on given config file.
func NewStore(configPath string, opts StoreOptions) (Store, error) {
	cfg, err := loadConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	return &dynamicStore{
		config:  cfg,
		options: opts,
	}, nil
}

// Get retrieves credentials from the store for the given server address.
func (ds *dynamicStore) Get(ctx context.Context, serverAddress string) (auth.Credential, error) {
	store, err := ds.getStore(serverAddress)
	if err != nil {
		return auth.EmptyCredential, err
	}
	return store.Get(ctx, serverAddress)
}

// Put saves credentials into the store for the given server address.
// Returns ErrPlaintextPutDisabled if native store is not available and
// StoreOptions.AllowPlaintextPut is set to false.
func (ds *dynamicStore) Put(ctx context.Context, serverAddress string, cred auth.Credential) error {
	store, err := ds.getStore(serverAddress)
	if err != nil {
		return err
	}
	return store.Put(ctx, serverAddress, cred)
}

// Delete removes credentials from the store for the given server address.
func (ds *dynamicStore) Delete(ctx context.Context, serverAddress string) error {
	store, err := ds.getStore(serverAddress)
	if err != nil {
		return err
	}
	return store.Delete(ctx, serverAddress)
}

// getHelperSuffix returns the credential helper suffix for the given server
// address.
func (ds *dynamicStore) getHelperSuffix(serverAddress string) string {
	// 1. Look for a server-specific credential helper first
	if helper := ds.config.CredentialHelpers[serverAddress]; helper != "" {
		return helper
	}
	// 2. Then look for the configured native store
	return ds.config.CredentialsStore
}

// getStore returns a store for the given server address.
func (ds *dynamicStore) getStore(serverAddress string) (Store, error) {
	if helper := ds.getHelperSuffix(serverAddress); helper != "" {
		return NewNativeStore(helper), nil
	}

	return newFileStore(ds.config)
}
