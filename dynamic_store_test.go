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
	"reflect"
	"testing"

	"github.com/oras-project/oras-credentials-go/internal/config"
	"oras.land/oras-go/v2/registry/remote/auth"
)

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
			cfg, err := config.LoadConfigFile(tt.configPath)
			if err != nil {
				t.Fatal("config.LoadConfigFile() error =", err)
			}
			ds := &dynamicStore{config: cfg}
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
			cfg, err := config.LoadConfigFile(tt.configPath)
			if err != nil {
				t.Fatal("config.LoadConfigFile() error =", err)
			}
			ds := &dynamicStore{config: cfg}
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
			cfg, err := config.LoadConfigFile(tt.configPath)
			if err != nil {
				t.Fatal("config.LoadConfigFile() error =", err)
			}
			ds := &dynamicStore{config: cfg}
			gotStore, err := ds.getStore(tt.serverAddress)
			if err != nil {
				t.Fatal("dynamicStore.getStore() error =", err)
			}
			if _, ok := gotStore.(*FileStore); !ok {
				t.Errorf("gotStore is not a file store")
			}
		})
	}
}
