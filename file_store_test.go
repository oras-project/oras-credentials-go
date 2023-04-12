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

type testAuthConfig struct {
	SomeAuthField string `json:"some_auth_field"`
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	Auth          string `json:"auth,omitempty"`

	// IdentityToken is used to authenticate the user and get
	// an access token for the registry.
	IdentityToken string `json:"identitytoken,omitempty"`
	// RegistryToken is a bearer token to be sent to a registry
	RegistryToken string `json:"registrytoken,omitempty"`
}

type testConfig struct {
	SomeConfigField int                       `json:"some_config_field"`
	AuthConfigs     map[string]testAuthConfig `json:"auths"`
}

func TestNewFileStore_badPath(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		configPath string
		wantErr    bool
	}{
		{
			name:       "Path is a directory",
			configPath: tempDir,
			wantErr:    true,
		},
		{
			name:       "Empty file name",
			configPath: filepath.Join(tempDir, ""),
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFileStore(tt.configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFileStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestNewFileStore_badFormat(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		wantErr    error
	}{
		{
			name:       "Bad JSON format",
			configPath: "testdata/bad_config",
			wantErr:    ErrInvalidConfigFormat,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFileStore(tt.configPath)
			if !errors.Is(err, ErrInvalidConfigFormat) {
				t.Errorf("NewFileStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestFileStore_Get_validConfig(t *testing.T) {
	ctx := context.Background()
	fs, err := NewFileStore("testdata/valid_config.json")
	if err != nil {
		t.Fatal("NewFileStore() error =", err)
	}

	tests := []struct {
		name          string
		serverAddress string
		want          auth.Credential
		wantErr       bool
	}{
		{
			name:          "Username and password",
			serverAddress: "registry1.example.com",
			want: auth.Credential{
				Username: "username",
				Password: "password",
			},
		},
		{
			name:          "Identity token",
			serverAddress: "registry2.example.com",
			want: auth.Credential{
				RefreshToken: "identity_token",
			},
		},
		{
			name:          "Registry token",
			serverAddress: "registry3.example.com",
			want: auth.Credential{
				AccessToken: "registry_token",
			},
		},
		{
			name:          "Username and password, identity token and registry token",
			serverAddress: "registry4.example.com",
			want: auth.Credential{
				Username:     "username",
				Password:     "password",
				RefreshToken: "identity_token",
				AccessToken:  "registry_token",
			},
		},
		{
			name:          "Empty credential",
			serverAddress: "registry5.example.com",
			want:          auth.EmptyCredential,
		},
		{
			name:          "Username and password, no auth",
			serverAddress: "registry6.example.com",
			want: auth.Credential{
				Username: "username",
				Password: "password",
			},
		},
		{
			name:          "Auth overriding Username and password",
			serverAddress: "registry7.example.com",
			want: auth.Credential{
				Username: "username",
				Password: "password",
			},
		},
		{
			name:          "Not in auths",
			serverAddress: "foo.example.com",
			want:          auth.EmptyCredential,
		},
		{
			name:          "No record",
			serverAddress: "registry999.example.com",
			want:          auth.EmptyCredential,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fs.Get(ctx, tt.serverAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileStore.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileStore.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStore_Get_invalidConfig(t *testing.T) {
	ctx := context.Background()
	fs, err := NewFileStore("testdata/invalid_config.json")
	if err != nil {
		t.Fatal("NewFileStore() error =", err)
	}

	tests := []struct {
		name          string
		serverAddress string
		want          auth.Credential
		wantErr       error
	}{
		{
			name:          "Invalid auth encode",
			serverAddress: "registry1.example.com",
			want:          auth.EmptyCredential,
			wantErr:       ErrInvalidConfigFormat,
		},
		{
			name:          "Invalid auths format",
			serverAddress: "registry2.example.com",
			want:          auth.EmptyCredential,
			wantErr:       ErrInvalidConfigFormat,
		},
		{
			name:          "Invalid type",
			serverAddress: "registry3.example.com",
			want:          auth.EmptyCredential,
			wantErr:       ErrInvalidConfigFormat,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fs.Get(ctx, tt.serverAddress)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FileStore.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileStore.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStore_Get_emptyConfig(t *testing.T) {
	ctx := context.Background()
	fs, err := NewFileStore("testdata/empty_config.json")
	if err != nil {
		t.Fatal("NewFileStore() error =", err)
	}

	tests := []struct {
		name          string
		serverAddress string
		want          auth.Credential
		wantErr       error
	}{
		{
			name:          "Not found",
			serverAddress: "registry.example.com",
			want:          auth.EmptyCredential,
			wantErr:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fs.Get(ctx, tt.serverAddress)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FileStore.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileStore.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStore_Get_notExistConfig(t *testing.T) {
	ctx := context.Background()
	fs, err := NewFileStore("whatever")
	if err != nil {
		t.Fatal("NewFileStore() error =", err)
	}

	tests := []struct {
		name          string
		serverAddress string
		want          auth.Credential
		wantErr       error
	}{
		{
			name:          "Not found",
			serverAddress: "registry.example.com",
			want:          auth.EmptyCredential,
			wantErr:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fs.Get(ctx, tt.serverAddress)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FileStore.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileStore.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStore_Put_notExistConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	ctx := context.Background()

	fs, err := NewFileStore(configPath)
	if err != nil {
		t.Fatal("NewFileStore() error =", err)
	}

	server := "test.example.com"
	cred := auth.Credential{
		Username:     "username",
		Password:     "password",
		RefreshToken: "refresh_token",
		AccessToken:  "access_token",
	}

	// test put
	if err := fs.Put(ctx, server, cred); err != nil {
		t.Fatalf("FileStore.Put() error = %v", err)
	}

	// verify config file
	configFile, err := os.Open(configPath)
	if err != nil {
		t.Fatalf("failed to open config file: %v", err)
	}
	defer configFile.Close()

	var cfg testConfig
	if err := json.NewDecoder(configFile).Decode(&cfg); err != nil {
		t.Fatalf("failed to decode config file: %v", err)
	}
	want := testConfig{
		AuthConfigs: map[string]testAuthConfig{
			server: {
				Auth:          "dXNlcm5hbWU6cGFzc3dvcmQ=",
				IdentityToken: "refresh_token",
				RegistryToken: "access_token",
			},
		},
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("Decoded config = %v, want %v", cfg, want)
	}

	// verify get
	got, err := fs.Get(ctx, server)
	if err != nil {
		t.Fatalf("FileStore.Get() error = %v", err)
	}
	if want := cred; !reflect.DeepEqual(got, want) {
		t.Errorf("FileStore.Get() = %v, want %v", got, want)
	}
}

func TestFileStore_Put_addNew(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	ctx := context.Background()

	// prepare test content
	server1 := "registry1.example.com"
	cred1 := auth.Credential{
		Username:     "username",
		Password:     "password",
		RefreshToken: "refresh_token",
		AccessToken:  "access_token",
	}

	cfg := testConfig{
		AuthConfigs: map[string]testAuthConfig{
			server1: {
				SomeAuthField: "whatever",
				Auth:          "dXNlcm5hbWU6cGFzc3dvcmQ=",
				IdentityToken: cred1.RefreshToken,
				RegistryToken: cred1.AccessToken,
			},
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

	// test put
	fs, err := NewFileStore(configPath)
	if err != nil {
		t.Fatal("NewFileStore() error =", err)
	}
	server2 := "registry2.example.com"
	cred2 := auth.Credential{
		Username:     "username_2",
		Password:     "password_2",
		RefreshToken: "refresh_token_2",
		AccessToken:  "access_token_2",
	}
	if err := fs.Put(ctx, server2, cred2); err != nil {
		t.Fatalf("FileStore.Put() error = %v", err)
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
			server1: {
				SomeAuthField: "whatever",
				Auth:          "dXNlcm5hbWU6cGFzc3dvcmQ=",
				IdentityToken: cred1.RefreshToken,
				RegistryToken: cred1.AccessToken,
			},
			server2: {
				Auth:          "dXNlcm5hbWVfMjpwYXNzd29yZF8y",
				IdentityToken: "refresh_token_2",
				RegistryToken: "access_token_2",
			},
		},
		SomeConfigField: cfg.SomeConfigField,
	}
	if !reflect.DeepEqual(gotCfg, wantCfg) {
		t.Errorf("Decoded config = %v, want %v", gotCfg, wantCfg)
	}

	// verify get
	got, err := fs.Get(ctx, server1)
	if err != nil {
		t.Fatalf("FileStore.Get() error = %v", err)
	}
	if want := cred1; !reflect.DeepEqual(got, want) {
		t.Errorf("FileStore.Get(%s) = %v, want %v", server1, got, want)
	}

	got, err = fs.Get(ctx, server2)
	if err != nil {
		t.Fatalf("FileStore.Get() error = %v", err)
	}
	if want := cred2; !reflect.DeepEqual(got, want) {
		t.Errorf("FileStore.Get(%s) = %v, want %v", server2, got, want)
	}
}

func TestFileStore_Put_updateOld(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	ctx := context.Background()

	// prepare test content
	server := "registry.example.com"
	cfg := testConfig{
		AuthConfigs: map[string]testAuthConfig{
			server: {
				SomeAuthField: "whatever",
				Username:      "foo",
				Password:      "bar",
				IdentityToken: "refresh_token",
			},
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

	// test put
	fs, err := NewFileStore(configPath)
	if err != nil {
		t.Fatal("NewFileStore() error =", err)
	}
	cred := auth.Credential{
		Username:    "username",
		Password:    "password",
		AccessToken: "access_token",
	}
	if err := fs.Put(ctx, server, cred); err != nil {
		t.Fatalf("FileStore.Put() error = %v", err)
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
			server: {
				Auth:          "dXNlcm5hbWU6cGFzc3dvcmQ=",
				RegistryToken: "access_token",
			},
		},
		SomeConfigField: cfg.SomeConfigField,
	}
	if !reflect.DeepEqual(gotCfg, wantCfg) {
		t.Errorf("Decoded config = %v, want %v", gotCfg, wantCfg)
	}

	// verify get
	got, err := fs.Get(ctx, server)
	if err != nil {
		t.Fatalf("FileStore.Get() error = %v", err)
	}
	if want := cred; !reflect.DeepEqual(got, want) {
		t.Errorf("FileStore.Get(%s) = %v, want %v", server, got, want)
	}
}

func TestStore_Put_disableSave(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	ctx := context.Background()

	fs, err := NewFileStore(configPath)
	if err != nil {
		t.Fatal("NewFileStore() error =", err)
	}
	fs.DisableSave = true

	server := "test.example.com"
	cred := auth.Credential{
		Username:     "username",
		Password:     "password",
		RefreshToken: "refresh_token",
		AccessToken:  "access_token",
	}
	err = fs.Put(ctx, server, cred)
	if wantErr := ErrPlaintextSaveDisabled; !errors.Is(err, wantErr) {
		t.Errorf("FileStore.Put() error = %v, wantErr %v", err, wantErr)
	}
}

func TestFileStore_Delete(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	ctx := context.Background()

	// prepare test content
	server1 := "registry1.example.com"
	cred1 := auth.Credential{
		Username:     "username",
		Password:     "password",
		RefreshToken: "refresh_token",
		AccessToken:  "access_token",
	}
	server2 := "registry2.example.com"
	cred2 := auth.Credential{
		Username:     "username_2",
		Password:     "password_2",
		RefreshToken: "refresh_token_2",
		AccessToken:  "access_token_2",
	}

	cfg := testConfig{
		AuthConfigs: map[string]testAuthConfig{
			server1: {
				Auth:          "dXNlcm5hbWU6cGFzc3dvcmQ=",
				IdentityToken: cred1.RefreshToken,
				RegistryToken: cred1.AccessToken,
			},
			server2: {
				Auth:          "dXNlcm5hbWVfMjpwYXNzd29yZF8y",
				IdentityToken: "refresh_token_2",
				RegistryToken: "access_token_2",
			},
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

	fs, err := NewFileStore(configPath)
	if err != nil {
		t.Fatal("NewFileStore() error =", err)
	}
	// test get
	got, err := fs.Get(ctx, server1)
	if err != nil {
		t.Fatalf("FileStore.Get() error = %v", err)
	}
	if want := cred1; !reflect.DeepEqual(got, want) {
		t.Errorf("FileStore.Get(%s) = %v, want %v", server1, got, want)
	}
	got, err = fs.Get(ctx, server2)
	if err != nil {
		t.Fatalf("FileStore.Get() error = %v", err)
	}
	if want := cred2; !reflect.DeepEqual(got, want) {
		t.Errorf("FileStore.Get(%s) = %v, want %v", server2, got, want)
	}

	// test delete
	if err := fs.Delete(ctx, server1); err != nil {
		t.Fatalf("FileStore.Delete() error = %v", err)
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
			server2: cfg.AuthConfigs[server2],
		},
		SomeConfigField: cfg.SomeConfigField,
	}
	if !reflect.DeepEqual(gotCfg, wantCfg) {
		t.Errorf("Decoded config = %v, want %v", gotCfg, wantCfg)
	}

	// test get again
	got, err = fs.Get(ctx, server1)
	if err != nil {
		t.Fatalf("FileStore.Get() error = %v", err)
	}
	if want := auth.EmptyCredential; !reflect.DeepEqual(got, want) {
		t.Errorf("FileStore.Get(%s) = %v, want %v", server1, got, want)
	}
	got, err = fs.Get(ctx, server2)
	if err != nil {
		t.Fatalf("FileStore.Get() error = %v", err)
	}
	if want := cred2; !reflect.DeepEqual(got, want) {
		t.Errorf("FileStore.Get(%s) = %v, want %v", server2, got, want)
	}
}

func TestFileStore_Delete_notExistRecord(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	ctx := context.Background()

	// prepare test content
	server := "registry1.example.com"
	cred := auth.Credential{
		Username:     "username",
		Password:     "password",
		RefreshToken: "refresh_token",
		AccessToken:  "access_token",
	}
	cfg := testConfig{
		AuthConfigs: map[string]testAuthConfig{
			server: {
				Auth:          "dXNlcm5hbWU6cGFzc3dvcmQ=",
				IdentityToken: cred.RefreshToken,
				RegistryToken: cred.AccessToken,
			},
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

	fs, err := NewFileStore(configPath)
	if err != nil {
		t.Fatal("NewFileStore() error =", err)
	}
	// test get
	got, err := fs.Get(ctx, server)
	if err != nil {
		t.Fatalf("FileStore.Get() error = %v", err)
	}
	if want := cred; !reflect.DeepEqual(got, want) {
		t.Errorf("FileStore.Get(%s) = %v, want %v", server, got, want)
	}

	// test delete
	if err := fs.Delete(ctx, "test.example.com"); err != nil {
		t.Fatalf("FileStore.Delete() error = %v", err)
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
			server: cfg.AuthConfigs[server],
		},
		SomeConfigField: cfg.SomeConfigField,
	}
	if !reflect.DeepEqual(gotCfg, wantCfg) {
		t.Errorf("Decoded config = %v, want %v", gotCfg, wantCfg)
	}

	// test get again
	got, err = fs.Get(ctx, server)
	if err != nil {
		t.Fatalf("FileStore.Get() error = %v", err)
	}
	if want := cred; !reflect.DeepEqual(got, want) {
		t.Errorf("FileStore.Get(%s) = %v, want %v", server, got, want)
	}
}

func TestFileStore_Delete_notExistConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	ctx := context.Background()

	fs, err := NewFileStore(configPath)
	if err != nil {
		t.Fatal("NewFileStore() error =", err)
	}

	server := "test.example.com"
	// test delete
	if err := fs.Delete(ctx, server); err != nil {
		t.Fatalf("FileStore.Delete() error = %v", err)
	}

	// verify config file is not created
	_, err = os.Stat(configPath)
	if wantErr := os.ErrNotExist; !errors.Is(err, wantErr) {
		t.Errorf("Stat(%s) error = %v, wantErr %v", configPath, err, wantErr)
	}
}

func Test_encodeAuth(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		want     string
	}{
		{
			name:     "Username and password",
			username: "username",
			password: "password",
			want:     "dXNlcm5hbWU6cGFzc3dvcmQ=",
		},
		{
			name:     "Empty username",
			username: "username",
			password: "",
			want:     "dXNlcm5hbWU6",
		},
		{
			name:     "Empty password",
			username: "",
			password: "password",
			want:     "OnBhc3N3b3Jk",
		},
		{
			name:     "Empty username and empty password",
			username: "",
			password: "",
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := encodeAuth(tt.username, tt.password); got != tt.want {
				t.Errorf("encodeAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decodeAuth(t *testing.T) {
	tests := []struct {
		name     string
		authStr  string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "Success",
			authStr:  "dXNlcm5hbWU6cGFzc3dvcmQ=", // username:password
			username: "username",
			password: "password",
		},
		{
			name:     "Invalid base64",
			authStr:  "whatever",
			username: "",
			password: "",
			wantErr:  true,
		},
		{
			name:     "Invalid username password format",
			authStr:  "d2hhdGV2ZXI=", // whatever
			username: "",
			password: "",
			wantErr:  true,
		},
		{
			name:     "Empty string",
			authStr:  "",
			username: "",
			password: "",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUsername, gotPassword, err := decodeAuth(tt.authStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUsername != tt.username {
				t.Errorf("decodeAuth() got = %v, want %v", gotUsername, tt.username)
			}
			if gotPassword != tt.password {
				t.Errorf("decodeAuth() got1 = %v, want %v", gotPassword, tt.password)
			}
		})
	}
}