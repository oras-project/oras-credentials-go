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

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/oras-project/oras-credentials-go/internal/config/configtest"
)

func TestConfig_IsAuthConfigured(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name             string
		fileName         string
		shouldCreateFile bool
		cfg              configtest.Config
		want             bool
	}{
		{
			name:             "not existing file",
			fileName:         "config.json",
			shouldCreateFile: false,
			cfg:              configtest.Config{},
			want:             false,
		},
		{
			name:             "no auth",
			fileName:         "config.json",
			shouldCreateFile: true,
			cfg: configtest.Config{
				SomeConfigField: 123,
			},
			want: false,
		},
		{
			name:             "empty auths exist",
			fileName:         "empty_auths.json",
			shouldCreateFile: true,
			cfg: configtest.Config{
				AuthConfigs: map[string]configtest.AuthConfig{},
			},
			want: false,
		},
		{
			name:             "auths exist, but no credential",
			fileName:         "no_cred_auths.json",
			shouldCreateFile: true,
			cfg: configtest.Config{
				AuthConfigs: map[string]configtest.AuthConfig{
					"test.example.com": {},
				},
			},
			want: true,
		},
		{
			name:             "auths exist",
			fileName:         "auths.json",
			shouldCreateFile: true,
			cfg: configtest.Config{
				AuthConfigs: map[string]configtest.AuthConfig{
					"test.example.com": {
						Auth: "dXNlcm5hbWU6cGFzc3dvcmQ=",
					},
				},
			},
			want: true,
		},
		{
			name:             "credsStore exists",
			fileName:         "credsStore.json",
			shouldCreateFile: true,
			cfg: configtest.Config{
				CredentialsStore: "teststore",
			},
			want: true,
		},
		{
			name:             "empty credHelpers exist",
			fileName:         "empty_credsStore.json",
			shouldCreateFile: true,
			cfg: configtest.Config{
				CredentialHelpers: map[string]string{},
			},
			want: false,
		},
		{
			name:             "credHelpers exist",
			fileName:         "credsStore.json",
			shouldCreateFile: true,
			cfg: configtest.Config{
				CredentialHelpers: map[string]string{
					"test.example.com": "testhelper",
				},
			},
			want: true,
		},
		{
			name:             "all exist",
			fileName:         "credsStore.json",
			shouldCreateFile: true,
			cfg: configtest.Config{
				SomeConfigField: 123,
				AuthConfigs: map[string]configtest.AuthConfig{
					"test.example.com": {},
				},
				CredentialsStore: "teststore",
				CredentialHelpers: map[string]string{
					"test.example.com": "testhelper",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare test content
			configPath := filepath.Join(tempDir, tt.fileName)
			if tt.shouldCreateFile {
				jsonStr, err := json.Marshal(tt.cfg)
				if err != nil {
					t.Fatalf("failed to marshal config: %v", err)
				}
				if err := os.WriteFile(configPath, jsonStr, 0666); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
			}

			cfg, err := Load(configPath)
			if err != nil {
				t.Fatal("LoadConfigFile() error =", err)
			}
			if got := cfg.IsAuthConfigured(); got != tt.want {
				t.Errorf("IsAuthConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_saveFile(t *testing.T) {
	tempDir := t.TempDir()
	tests := []struct {
		name             string
		fileName         string
		shouldCreateFile bool
		oldCfg           configtest.Config
		newCfg           configtest.Config
		wantCfg          configtest.Config
	}{
		{
			name:     "set credsStore in a non-existing file",
			fileName: "config.json",
			oldCfg:   configtest.Config{},
			newCfg: configtest.Config{
				CredentialsStore: "teststore",
			},
			wantCfg: configtest.Config{
				AuthConfigs:      make(map[string]configtest.AuthConfig),
				CredentialsStore: "teststore",
			},
			shouldCreateFile: false,
		},
		{
			name:     "set credsStore in empty file",
			fileName: "empty.json",
			oldCfg:   configtest.Config{},
			newCfg: configtest.Config{
				CredentialsStore: "teststore",
			},
			wantCfg: configtest.Config{
				AuthConfigs:      make(map[string]configtest.AuthConfig),
				CredentialsStore: "teststore",
			},
			shouldCreateFile: true,
		},
		{
			name:     "set credsStore in a no-auth-configured file",
			fileName: "empty.json",
			oldCfg: configtest.Config{
				SomeConfigField: 123,
			},
			newCfg: configtest.Config{
				CredentialsStore: "teststore",
			},
			wantCfg: configtest.Config{
				SomeConfigField:  123,
				AuthConfigs:      make(map[string]configtest.AuthConfig),
				CredentialsStore: "teststore",
			},
			shouldCreateFile: true,
		},
		{
			name:     "Set credsStore and credHelpers in an auth-configured file",
			fileName: "auth_configured.json",
			oldCfg: configtest.Config{
				SomeConfigField: 123,
				AuthConfigs: map[string]configtest.AuthConfig{
					"registry1.example.com": {
						SomeAuthField: "something",
						Auth:          "dXNlcm5hbWU6cGFzc3dvcmQ=",
					},
				},
				CredentialsStore: "oldstore",
				CredentialHelpers: map[string]string{
					"registry2.example.com": "testhelper",
				},
			},
			newCfg: configtest.Config{
				AuthConfigs:      make(map[string]configtest.AuthConfig),
				SomeConfigField:  123,
				CredentialsStore: "newstore",
				CredentialHelpers: map[string]string{
					"xxx": "yyy",
				},
			},
			wantCfg: configtest.Config{
				SomeConfigField: 123,
				AuthConfigs: map[string]configtest.AuthConfig{
					"registry1.example.com": {
						SomeAuthField: "something",
						Auth:          "dXNlcm5hbWU6cGFzc3dvcmQ=",
					},
				},
				CredentialsStore: "newstore",
				CredentialHelpers: map[string]string{
					"registry2.example.com": "testhelper", // cred helpers will not be updated
				},
			},
			shouldCreateFile: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare test content
			configPath := filepath.Join(tempDir, tt.fileName)
			if tt.shouldCreateFile {
				jsonStr, err := json.Marshal(tt.oldCfg)
				if err != nil {
					t.Fatalf("failed to marshal config: %v", err)
				}
				if err := os.WriteFile(configPath, jsonStr, 0666); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
			}

			cfg, err := Load(configPath)
			if err != nil {
				t.Fatal("LoadConfigFile() error =", err)
			}
			cfg.credentialsStore = tt.newCfg.CredentialsStore
			cfg.credentialHelpers = tt.newCfg.CredentialHelpers
			if err := cfg.saveFile(); err != nil {
				t.Fatal("saveFile() error =", err)
			}

			// verify config file
			configFile, err := os.Open(configPath)
			if err != nil {
				t.Fatalf("failed to open config file: %v", err)
			}
			defer configFile.Close()
			var gotCfg configtest.Config
			if err := json.NewDecoder(configFile).Decode(&gotCfg); err != nil {
				t.Fatalf("failed to decode config file: %v", err)
			}
			if !reflect.DeepEqual(gotCfg, tt.wantCfg) {
				t.Errorf("Decoded config = %v, want %v", gotCfg, tt.wantCfg)
			}
		})
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
			name:     "Username only",
			username: "username",
			password: "",
			want:     "dXNlcm5hbWU6",
		},
		{
			name:     "Password only",
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
			name:     "Valid base64",
			authStr:  "dXNlcm5hbWU6cGFzc3dvcmQ=", // username:password
			username: "username",
			password: "password",
		},
		{
			name:     "Valid base64, username only",
			authStr:  "dXNlcm5hbWU6", // username:
			username: "username",
		},
		{
			name:     "Valid base64, password only",
			authStr:  "OnBhc3N3b3Jk", // :password
			password: "password",
		},
		{
			name:     "Valid base64, bad format",
			authStr:  "d2hhdGV2ZXI=", // whatever
			username: "",
			password: "",
			wantErr:  true,
		},
		{
			name:     "Invalid base64",
			authStr:  "whatever",
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
