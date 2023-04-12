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
	"encoding/json"
	"fmt"
	"os"
	"sync"

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
	credentialConfig
	configPath    string
	options       StoreOptions
	fileStore     *FileStore
	fileStoreOnce sync.Once
}

// LoadConfig loads the config file for ds.
func (ds *dynamicStore) LoadConfig() error {
	configFile, err := os.Open(ds.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// the config file can be created when needed
			return nil
		}
		return fmt.Errorf("failed to open config file at %s: %w", ds.configPath, err)
	}
	defer configFile.Close()

	// decode credential config if the config file exists
	if err := json.NewDecoder(configFile).Decode(&ds.credentialConfig); err != nil {
		return fmt.Errorf("failed to decode config file at %s: %w: %v", ds.configPath, ErrInvalidConfigFormat, err)
	}
	return nil
}

// StoreOptions provides options for NewStore.
type StoreOptions struct {
	// AllowPlaintext allows saving credentials in plaintext in the config file.
	AllowPlaintext bool
}

// NewStore returns a store based on given config file.
func NewStore(configPath string, opts StoreOptions) (Store, error) {
	ds := &dynamicStore{
		configPath: configPath,
		options:    opts,
	}

	if err := ds.LoadConfig(); err != nil {
		return nil, err
	}
	return ds, nil
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
	if helper := ds.CredentialHelpers[serverAddress]; helper != "" {
		return helper
	}
	// 2. Then look for the configured native store
	return ds.CredentialsStore
}

// getStore returns a store for the given server address.
func (ds *dynamicStore) getStore(serverAddress string) (Store, error) {
	if helper := ds.getHelperSuffix(serverAddress); helper != "" {
		return NewNativeStore(helper), nil
	}

	// lazy loading file store when no native store is available
	var err error
	ds.fileStoreOnce.Do(func() {
		ds.fileStore, err = NewFileStore(ds.configPath)
		if err != nil {
			return
		}
		if !ds.options.AllowPlaintext {
			ds.fileStore.DisableSave = true
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize file store: %w", err)
	}
	return ds.fileStore, nil
}
