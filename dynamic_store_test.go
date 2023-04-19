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
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"oras.land/oras-go/v2/registry/remote/auth"
)

func Test_dynamicStore_interface(t *testing.T) {
	var ds interface{} = &dynamicStore{}
	if _, ok := ds.(Store); !ok {
		t.Error("&dynamicStore{} does not conform Store")
	}
}

func Test_dynamicStore_Get_fileStore(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name          string
		configPath    string
		serverAddress string
		want          auth.Credential
		wantErr       bool
	}{
		{
			name:          "registry3.example.com",
			configPath:    "testdata/credHelpers_config.json",
			serverAddress: "registry3.example.com",
			want: auth.Credential{
				Username: "foo",
				Password: "bar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, err := NewStore(tt.configPath, StoreOptions{})
			if err != nil {
				t.Fatal("NewStore() error =", err)
			}
			got, err := ds.Get(ctx, tt.serverAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("dynamicStore.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dynamicStore.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dynamicStore_Put_AllowPlainTextPut(t *testing.T) {
	// prepare test content
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	serverAddr := "newtest.example.com"
	cred := auth.Credential{
		Username: "username",
		Password: "password",
	}
	ctx := context.Background()

	cfg := testConfig{
		AuthConfigs: map[string]testAuthConfig{
			"test.example.com": {},
		},
		SomeConfigField: 123,
	}
	jsonStr, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, jsonStr, 0666); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// test default option
	ds, err := NewStore(configPath, StoreOptions{})
	if err != nil {
		t.Fatal("NewStore() error =", err)
	}
	err = ds.Put(ctx, serverAddr, cred)
	if wantErr := ErrPlaintextPutDisabled; !errors.Is(err, wantErr) {
		t.Errorf("dynamicStore.Put() error = %v, wantErr %v", err, wantErr)
	}

	// test AllowPlainTextPut = true
	ds, err = NewStore(configPath, StoreOptions{AllowPlaintextPut: true})
	if err != nil {
		t.Fatal("NewStore() error =", err)
	}
	if err := ds.Put(ctx, serverAddr, cred); err != nil {
		t.Error("dynamicStore.Put() error =", err)
	}

	// verify config file
	configFile, err := os.Open(configPath)
	if err != nil {
		t.Fatalf("failed to open config file: %v", err)
	}
	defer configFile.Close()
	var gotCfg testConfig
	if err := json.NewDecoder(configFile).Decode(&gotCfg); err != nil {
		t.Fatalf("failed to decode config file: %v", err)
	}
	wantCfg := testConfig{
		AuthConfigs: map[string]testAuthConfig{
			"test.example.com": {},
			serverAddr: {
				Auth: "dXNlcm5hbWU6cGFzc3dvcmQ=",
			},
		},
		SomeConfigField: cfg.SomeConfigField,
	}
	if !reflect.DeepEqual(gotCfg, wantCfg) {
		t.Errorf("Decoded config = %v, want %v", gotCfg, wantCfg)
	}
}

func Test_dynamicStore_getHelperSuffix(t *testing.T) {
	tests := []struct {
		name          string
		configPath    string
		serverAddress string
		want          string
	}{
		{
			name:          "Get cred helper: registry_helper1",
			configPath:    "testdata/credHelpers_config.json",
			serverAddress: "registry1.example.com",
			want:          "registry1-helper",
		},
		{
			name:          "Get cred helper: registry_helper2",
			configPath:    "testdata/credHelpers_config.json",
			serverAddress: "registry2.example.com",
			want:          "registry2-helper",
		},
		{
			name:          "Empty cred helper configured",
			configPath:    "testdata/credHelpers_config.json",
			serverAddress: "registry3.example.com",
			want:          "",
		},
		{
			name:          "No cred helper and creds store configured",
			configPath:    "testdata/credHelpers_config.json",
			serverAddress: "whatever.example.com",
			want:          "",
		},
		{
			name:          "Choose cred helper over creds store",
			configPath:    "testdata/credsStore_config.json",
			serverAddress: "test.example.com",
			want:          "test-helper",
		},
		{
			name:          "No cred helper configured, choose cred store",
			configPath:    "testdata/credsStore_config.json",
			serverAddress: "whatever.example.com",
			want:          "teststore",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewStore(tt.configPath, StoreOptions{})
			if err != nil {
				t.Fatal("NewStore() error =", err)
			}
			ds := store.(*dynamicStore)
			if got := ds.getHelperSuffix(tt.serverAddress); got != tt.want {
				t.Errorf("dynamicStore.getHelperSuffix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dynamicStore_getStore_nativeStore(t *testing.T) {
	tests := []struct {
		name          string
		configPath    string
		serverAddress string
	}{
		{
			name:          "Cred helper configured for registry1.example.com",
			configPath:    "testdata/credHelpers_config.json",
			serverAddress: "registry1.example.com",
		},
		{
			name:          "Cred helper configured for registry2.example.com",
			configPath:    "testdata/credHelpers_config.json",
			serverAddress: "registry2.example.com",
		},
		{
			name:          "Cred helper configured for test.example.com",
			configPath:    "testdata/credsStore_config.json",
			serverAddress: "test.example.com",
		},
		{
			name:          "No cred helper configured, use creds store",
			configPath:    "testdata/credsStore_config.json",
			serverAddress: "whaterver.example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewStore(tt.configPath, StoreOptions{})
			if err != nil {
				t.Fatal("NewStore() error =", err)
			}
			ds := store.(*dynamicStore)
			gotStore, err := ds.getStore(tt.serverAddress)
			if err != nil {
				t.Fatal("dynamicStore.getStore() error =", err)
			}
			if _, ok := gotStore.(*NativeStore); !ok {
				t.Errorf("gotStore is not a native store")
			}
		})
	}
}

func Test_dynamicStore_getStore_fileStore(t *testing.T) {
	tests := []struct {
		name          string
		configPath    string
		serverAddress string
	}{
		{
			name:          "Empty cred helper configured for registry3.example.com",
			configPath:    "testdata/credHelpers_config.json",
			serverAddress: "registry3.example.com",
		},
		{
			name:          "No cred helper configured",
			configPath:    "testdata/credHelpers_config.json",
			serverAddress: "whatever.example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewStore(tt.configPath, StoreOptions{})
			if err != nil {
				t.Fatal("NewStore() error =", err)
			}
			ds := store.(*dynamicStore)
			gotStore, err := ds.getStore(tt.serverAddress)
			if err != nil {
				t.Fatal("dynamicStore.getStore() error =", err)
			}
			gotFS1, ok := gotStore.(*FileStore)
			if !ok {
				t.Errorf("gotStore is not a file store")
			}

			// get again, the two file stores should be based on the same config instance
			gotStore, err = ds.getStore(tt.serverAddress)
			if err != nil {
				t.Fatal("dynamicStore.getStore() error =", err)
			}
			gotFS2, ok := gotStore.(*FileStore)
			if !ok {
				t.Errorf("gotStore is not a file store")
			}
			if gotFS1.config != gotFS2.config {
				t.Errorf("gotFS1 and gotFS2 are not based on the same config")
			}
		})
	}
}
