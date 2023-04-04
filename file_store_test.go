package credentials

import (
	"context"
	"errors"
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
		wantErr       error
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
			wantErr:       ErrCredentialNotFound,
		},
		{
			name:          "No record",
			serverAddress: "registry999.example.com",
			want:          auth.EmptyCredential,
			wantErr:       ErrCredentialNotFound,
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
			wantErr:       ErrCredentialNotFound,
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
			wantErr:       ErrCredentialNotFound,
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
			wantErr:       ErrCredentialNotFound,
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

// TODO: how to test?
func TestFileStore_Put(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.test.json")

	fs, err := NewFileStore(configPath)
	if err != nil {
		panic(err)
	}

	reg := "test0331.test.com"
	cred := auth.Credential{
		Username:     "username",
		Password:     "password",
		RefreshToken: "refresh_token",
		AccessToken:  "access_token",
	}

	ctx := context.Background()
	if err := fs.Put(ctx, reg, cred); err != nil {
		panic(err)
	}
}

func TestFileStore_Delete(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.test.json")

	fs, err := NewFileStore(configPath)
	if err != nil {
		panic(err)
	}

	reg1 := "test1.test.com"
	cred1 := auth.Credential{
		Username:     "username",
		Password:     "password",
		RefreshToken: "refresh_token",
		AccessToken:  "access_token",
	}

	ctx := context.Background()
	if err := fs.Put(ctx, reg1, cred1); err != nil {
		panic(err)
	}

	reg2 := "test2.test.com"
	cred2 := auth.Credential{
		Username:     "username2",
		Password:     "password2",
		RefreshToken: "refresh_token2",
		AccessToken:  "access_token2",
	}
	if err := fs.Put(ctx, reg2, cred2); err != nil {
		panic(err)
	}

	if err := fs.Delete(ctx, reg1); err != nil {
		panic(err)
	}
}
