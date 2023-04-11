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

	"oras.land/oras-go/v2/registry/remote/auth"
)

// StoreOptions provides options for NewStore.
type StoreOptions struct {
	// AllowPlainText allows saving credentials in plain text in configuration file.
	AllowPlainText bool
}

const (
	configFieldCredStore   = "credsStore"
	configFieldCredHelpers = "credHelpers"
)

type dynamicStore struct {
	configPath  string
	credStore   string
	credHelpers map[string]string
	fileStore   *FileStore
}

// TODO: when to use default store?
func NewStore(configPath string, opts StoreOptions) (Store, error) {
	ds := &dynamicStore{
		configPath: configPath,
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// TODO: no error?
			return ds, nil
		}
		return nil, fmt.Errorf("failed to open config file at %s: %w", configPath, err)
	}
	defer configFile.Close()

	// init file store
	ds.fileStore, err = NewFileStore(ds.configPath)
	if err != nil {
		return nil, err
	}
	if !opts.AllowPlainText {
		ds.fileStore.DisableSave = true
	}

	var content map[string]json.RawMessage
	if err := json.NewDecoder(configFile).Decode(&content); err != nil {
		return nil, fmt.Errorf("failed to decode config file at %s: %w: %v", configPath, ErrInvalidConfigFormat, err)
	}
	ds.credStore = string(content[configFieldCredStore])
	credHelperBytes, ok := content[configFieldCredHelpers]
	if !ok {
		return ds, nil
	}
	if err := json.Unmarshal(credHelperBytes, &ds.credHelpers); err != nil {
		return nil, fmt.Errorf("failed to decode config file at %s: %w: %v", configPath, ErrInvalidConfigFormat, err)
	}
	return ds, nil
}

// Get retrieves credentials from the store for the given server address.
func (ds *dynamicStore) Get(ctx context.Context, serverAddress string) (auth.Credential, error) {
	store, err := ds.getStore(serverAddress)
	if err != nil {
		return auth.EmptyCredential, nil
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

func (ds *dynamicStore) getStore(serverAddress string) (Store, error) {
	if helper, ok := ds.credHelpers[serverAddress]; ok {
		return NewNativeStore(helper), nil
	}
	if ds.credStore != "" {
		return NewNativeStore(ds.credStore), nil
	}
	return ds.fileStore, nil
}
