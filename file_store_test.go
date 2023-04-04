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
			name:          "Not found",
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

func TestFileStore_Get_ConfigNotExist(t *testing.T) {
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

func TestFileStore_Put(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	ctx := context.Background()

	fs, err := NewFileStore(configPath)
	if err != nil {
		panic(err)
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

	type config struct {
		AuthConfigs map[string]authConfig `json:"auths"`
	}
	var cfg config
	if err := json.NewDecoder(configFile).Decode(&cfg); err != nil {
		t.Fatalf("failed to decode config file: %v", err)
	}
	want := config{
		AuthConfigs: map[string]authConfig{
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

func TestFileStore_Put_ExistingConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	ctx := context.Background()

	server1 := "registry1.example.com"
	type config struct {
		AuthConfigs map[string]authConfig `json:"auths"`
		SomeField   string                `json:"some_field"`
	}
	cred1 := auth.Credential{
		Username:     "username",
		Password:     "password",
		RefreshToken: "refresh_token",
		AccessToken:  "access_token",
	}
	cfg := config{
		AuthConfigs: map[string]authConfig{
			server1: {
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
		panic(err)
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
	var gotCfg config
	if err := json.NewDecoder(configFile).Decode(&gotCfg); err != nil {
		t.Fatalf("failed to decode config file: %v", err)
	}
	wantCfg := config{
		AuthConfigs: map[string]authConfig{
			server1: cfg.AuthConfigs[server1],
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

func TestStore_Put_DisablePlainText(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	ctx := context.Background()

	fs, err := NewFileStore(configPath)
	if err != nil {
		panic(err)
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

// func TestFileStore_Delete(t *testing.T) {
// 	tempDir := t.TempDir()
// 	configPath := filepath.Join(tempDir, "config.test.json")

// 	fs, err := NewFileStore(configPath)
// 	if err != nil {
// 		panic(err)
// 	}

// 	reg1 := "test1.test.com"
// 	cred1 := auth.Credential{
// 		Username:     "username",
// 		Password:     "password",
// 		RefreshToken: "refresh_token",
// 		AccessToken:  "access_token",
// 	}

// 	ctx := context.Background()
// 	if err := fs.Put(ctx, reg1, cred1); err != nil {
// 		panic(err)
// 	}

// 	reg2 := "test2.test.com"
// 	cred2 := auth.Credential{
// 		Username:     "username2",
// 		Password:     "password2",
// 		RefreshToken: "refresh_token2",
// 		AccessToken:  "access_token2",
// 	}
// 	if err := fs.Put(ctx, reg2, cred2); err != nil {
// 		panic(err)
// 	}

// 	if err := fs.Delete(ctx, reg1); err != nil {
// 		panic(err)
// 	}
// }
