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
	SomeField   string                    `json:"some_field"`
	AuthConfigs map[string]testAuthConfig `json:"auths"`
}

func TestFileStore_Get_ValidConfig(t *testing.T) {
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

func TestFileStore_Get_InvalidConfig(t *testing.T) {
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
			wantErr:       ErrInvalidFormat,
		},
		{
			name:          "Invalid auths format",
			serverAddress: "registry2.example.com",
			want:          auth.EmptyCredential,
			wantErr:       nil,
		},
		{
			name:          "Invalid type",
			serverAddress: "registry3.example.com",
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

func TestFileStore_Get_EmptyConfig(t *testing.T) {
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

func TestFileStore_Get_NotExistConfig(t *testing.T) {
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

func TestFileStore_Put_NotExistConfig(t *testing.T) {
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

func TestFileStore_Put_AddNew(t *testing.T) {
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
				SomeAuthField: "foobar",
				Auth:          "dXNlcm5hbWU6cGFzc3dvcmQ=",
				IdentityToken: cred1.RefreshToken,
				RegistryToken: cred1.AccessToken,
			},
		},
		SomeField: "Some value",
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
				SomeAuthField: "foobar",
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
		SomeField: cfg.SomeField,
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

func TestFileStore_Put_UpdateOld(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	ctx := context.Background()

	// prepare test content
	server := "registry.example.com"
	cfg := testConfig{
		AuthConfigs: map[string]testAuthConfig{
			server: {
				SomeAuthField: "foobar",
				Username:      "foo",
				Password:      "bar",
				IdentityToken: "refresh_token",
			},
		},
		SomeField: "Some value",
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
				SomeAuthField: "foobar",
				Auth:          "dXNlcm5hbWU6cGFzc3dvcmQ=",
				RegistryToken: "access_token",
			},
		},
		SomeField: cfg.SomeField,
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

func TestStore_Put_DisableSave(t *testing.T) {
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
	if wantErr := ErrPlainTextSaveDisabled; !errors.Is(err, wantErr) {
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
		SomeField: "Some value",
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
		SomeField: cfg.SomeField,
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

func TestFileStore_Delete_NotExistRecord(t *testing.T) {
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
		SomeField: "Some value",
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
		SomeField: cfg.SomeField,
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

// TODO: verify config does not exist
func TestFileStore_Delete_NotExistConfig(t *testing.T) {
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
