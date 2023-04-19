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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/oras-project/oras-credentials-go/internal/ioutil"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// Config represents a docker configuration file.
// Reference: https://github.com/docker/cli/blob/v24.0.0-beta.2/cli/config/configfile/file.go#L17-L44
type Config struct {
	// path is the path to the config file.
	path string
	// rwLock is a read-write-lock for the file store.
	rwLock sync.RWMutex
	// content is the content of the config file.
	// Reference: https://github.com/docker/cli/blob/v24.0.0-beta.2/cli/config/configfile/file.go#L17-L44
	content map[string]json.RawMessage
	// authsCache is a cache of the auths field of the config.
	// Reference: https://github.com/docker/cli/blob/v24.0.0-beta.2/cli/config/configfile/file.go#L19
	authsCache map[string]json.RawMessage
	// credentialsStoreCache is a cache of the credsStore field of the config.
	// Reference: https://github.com/docker/cli/blob/v24.0.0-beta.2/cli/config/configfile/file.go#L28
	credentialsStoreCache string
	// credentialHelpersCache is a cache of the credHelpers field of the config.
	// Reference: https://github.com/docker/cli/blob/v24.0.0-beta.2/cli/config/configfile/file.go#L29
	credentialHelpersCache map[string]string
}

// AuthConfig contains authorization information for connecting to a Registry.
// References:
//   - https://github.com/docker/cli/blob/v24.0.0-beta.2/cli/config/configfile/file.go#L17-L45
//   - https://github.com/docker/cli/blob/v24.0.0-beta.2/cli/config/types/authconfig.go#L3-L22
type AuthConfig struct {
	// Auth is a base64-encoded string of "{username}:{password}".
	Auth string `json:"auth,omitempty"`
	// IdentityToken is used to authenticate the user and get.
	// an access token for the registry.
	IdentityToken string `json:"identitytoken,omitempty"`
	// RegistryToken is a bearer token to be sent to a registry.
	RegistryToken string `json:"registrytoken,omitempty"`

	Username string `json:"username,omitempty"` // legacy field for compatibility
	Password string `json:"password,omitempty"` // legacy field for compatibility
}

const (
	// configFieldAuths is the "auths" field in the config file.
	// Reference: https://github.com/docker/cli/blob/v24.0.0-beta.2/cli/config/configfile/file.go#L19
	configFieldAuths = "auths"
	// configFieldCredentialsStore is the "credsStore" field in the config file.
	configFieldCredentialsStore = "credsStore"
	// configFieldCredentialHelpers is the "credHelpers" field in the config file.
	configFieldCredentialHelpers = "credHelpers"
)

// ErrInvalidConfigFormat is returned when the config format is invalid.
var ErrInvalidConfigFormat = errors.New("invalid config format")

// NewAuthConfig creates an authConfig based on cred.
func NewAuthConfig(cred auth.Credential) AuthConfig {
	return AuthConfig{
		Auth:          encodeAuth(cred.Username, cred.Password),
		IdentityToken: cred.RefreshToken,
		RegistryToken: cred.AccessToken,
	}
}

// Credential returns an auth.Credential based on ac.
func (ac AuthConfig) Credential() (auth.Credential, error) {
	cred := auth.Credential{
		Username:     ac.Username,
		Password:     ac.Password,
		RefreshToken: ac.IdentityToken,
		AccessToken:  ac.RegistryToken,
	}
	if ac.Auth != "" {
		var err error
		// override username and password
		cred.Username, cred.Password, err = decodeAuth(ac.Auth)
		if err != nil {
			return auth.EmptyCredential, fmt.Errorf("failed to decode auth field: %w: %v", ErrInvalidConfigFormat, err)
		}
	}
	return cred, nil
}

// LoadConfigFile loads Config from the given config path.
func LoadConfigFile(configPath string) (*Config, error) {
	cfg := &Config{path: configPath}
	configFile, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// init content and caches if the content file does not exist
			cfg.content = make(map[string]json.RawMessage)
			cfg.authsCache = make(map[string]json.RawMessage)
			cfg.credentialHelpersCache = make(map[string]string)
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to open config file at %s: %w", configPath, err)
	}
	defer configFile.Close()

	// decode config content if the config file exists
	if err := json.NewDecoder(configFile).Decode(&cfg.content); err != nil {
		return nil, fmt.Errorf("failed to decode config file at %s: %w: %v", configPath, ErrInvalidConfigFormat, err)
	}

	if credsStoreBytes, ok := cfg.content[configFieldCredentialsStore]; ok {
		if err := json.Unmarshal(credsStoreBytes, &cfg.credentialsStoreCache); err != nil {
			return nil, fmt.Errorf("failed to unmarshal creds store field: %w: %v", ErrInvalidConfigFormat, err)
		}
	}

	if credHelpersBytes, ok := cfg.content[configFieldCredentialHelpers]; ok {
		if err := json.Unmarshal(credHelpersBytes, &cfg.credentialHelpersCache); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cred helpers field: %w: %v", ErrInvalidConfigFormat, err)
		}
	} else {
		cfg.credentialHelpersCache = make(map[string]string)
	}

	if authsBytes, ok := cfg.content[configFieldAuths]; ok {
		if err := json.Unmarshal(authsBytes, &cfg.authsCache); err != nil {
			return nil, fmt.Errorf("failed to unmarshal auths field: %w: %v", ErrInvalidConfigFormat, err)
		}
	} else {
		cfg.authsCache = make(map[string]json.RawMessage)
	}

	return cfg, nil
}

// GetAuthConfig returns an AuthConfig for serverAddress.
func (cfg *Config) GetAuthConfig(serverAddress string) (AuthConfig, error) {
	cfg.rwLock.RLock()
	defer cfg.rwLock.RUnlock()

	authCfgBytes, ok := cfg.authsCache[serverAddress]
	if !ok {
		return AuthConfig{}, nil
	}
	var authCfg AuthConfig
	if err := json.Unmarshal(authCfgBytes, &authCfg); err != nil {
		return AuthConfig{}, fmt.Errorf("failed to unmarshal auth field: %w: %v", ErrInvalidConfigFormat, err)
	}
	return authCfg, nil
}

// PutAuthConfig puts authCfg for serverAddress.
func (cfg *Config) PutAuthConfig(serverAddress string, authCfg AuthConfig) error {
	cfg.rwLock.Lock()
	defer cfg.rwLock.Unlock()

	authCfgBytes, err := json.Marshal(authCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal auth field: %w", err)
	}
	cfg.authsCache[serverAddress] = authCfgBytes
	return cfg.saveFile()
}

// DeleteAuthConfig deletes the corresponding AuthConfig for serverAddress.
func (cfg *Config) DeleteAuthConfig(serverAddress string) error {
	cfg.rwLock.Lock()
	defer cfg.rwLock.Unlock()

	if _, ok := cfg.authsCache[serverAddress]; !ok {
		// no ops
		return nil
	}
	delete(cfg.authsCache, serverAddress)
	return cfg.saveFile()
}

// GetCredentialHelpers returns the credential helpers for serverAddress.
func (cfg *Config) GetCredentialHelpers(serverAddress string) string {
	return cfg.credentialHelpersCache[serverAddress]
}

// GetCredentialHelpers returns the configured credentials store.
func (cfg *Config) GetCredentialsStore() string {
	cfg.rwLock.RLock()
	defer cfg.rwLock.RUnlock()

	return cfg.credentialsStoreCache
}

// PutCredentialsStore puts the configured credentials store.
func (cfg *Config) PutCredentialsStore(credsStore string) error {
	cfg.rwLock.Lock()
	defer cfg.rwLock.Unlock()

	cfg.credentialsStoreCache = credsStore
	return cfg.saveFile()
}

// saveFile saves Config into the file.
func (cfg *Config) saveFile() (returnErr error) {
	// marshal content
	if cfg.credentialsStoreCache != "" {
		credsStoreBytes, err := json.Marshal(cfg.credentialsStoreCache)
		if err != nil {
			return fmt.Errorf("failed to marshal creds store: %w", err)
		}
		cfg.content[configFieldCredentialsStore] = credsStoreBytes
	} else {
		// omit empty
		delete(cfg.content, configFieldCredentialsStore)
	}

	if len(cfg.credentialHelpersCache) > 0 {
		credHelpersBytes, err := json.Marshal(cfg.credentialHelpersCache)
		if err != nil {
			return fmt.Errorf("failed to marshal cred helpers: %w", err)
		}
		cfg.content[configFieldCredentialHelpers] = credHelpersBytes
	} else {
		// omit empty
		delete(cfg.content, configFieldCredentialHelpers)
	}

	authsBytes, err := json.Marshal(cfg.authsCache)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}
	cfg.content[configFieldAuths] = authsBytes
	jsonBytes, err := json.MarshalIndent(cfg.content, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// write the content to a ingest file for atomicity
	configDir := filepath.Dir(cfg.path)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to make directory %s: %w", configDir, err)
	}
	ingest, err := ioutil.Ingest(configDir, bytes.NewReader(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to save config file: %w", err)
	}
	defer func() {
		if returnErr != nil {
			// clean up the ingest file in case of error
			os.Remove(ingest)
		}
	}()

	// overwrite the config file
	if err := os.Rename(ingest, cfg.path); err != nil {
		return fmt.Errorf("failed to save config file: %w", err)
	}
	return nil
}

// IsAuthConfigured returns whether there is authentication configured in this
// config file or not.
func (cfg *Config) IsAuthConfigured() bool {
	return cfg.credentialsStoreCache != "" ||
		len(cfg.credentialHelpersCache) > 0 ||
		len(cfg.authsCache) > 0
}

// encodeAuth base64-encodes username and password into base64(username:password).
func encodeAuth(username, password string) string {
	if username == "" && password == "" {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

// decodeAuth decodes a base64 encoded string and returns username and password.
func decodeAuth(authStr string) (username string, password string, err error) {
	if authStr == "" {
		return "", "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(authStr)
	if err != nil {
		return "", "", err
	}
	decodedStr := string(decoded)
	username, password, ok := strings.Cut(decodedStr, ":")
	if !ok {
		return "", "", fmt.Errorf("auth '%s' does not conform the base64(username:password) format", decodedStr)
	}
	return username, password, nil
}
