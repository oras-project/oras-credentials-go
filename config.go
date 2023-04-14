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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/oras-project/oras-credentials-go/internal/ioutil"
)

type config struct {
	// Path is the path to the config file.
	Path string
	// Content is the Content of the config file.
	// Reference: https://github.com/docker/cli/blob/v24.0.0-beta.1/cli/config/configfile/file.go#L17-L45
	Content map[string]json.RawMessage
	// AuthsCache is a cache of the auths field of the config field.
	// Reference: https://github.com/docker/cli/blob/v24.0.0-beta.1/cli/config/configfile/file.go#L19
	AuthsCache        map[string]json.RawMessage
	CredentialsStore  string            `json:"credsStore,omitempty"`
	CredentialHelpers map[string]string `json:"credHelpers,omitempty"`
	// rwLock is a read-write-lock for the file store.
	rwLock sync.RWMutex
}

// authConfig contains authorization information for connecting to a Registry.
// References:
//   - https://github.com/docker/cli/blob/v24.0.0-beta.1/cli/config/configfile/file.go#L17-L45
//   - https://github.com/docker/cli/blob/v24.0.0-beta.1/cli/config/types/authconfig.go#L3-L22
type authConfig struct {
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

func loadConfigFile(configPath string) (*config, error) {
	cfg := &config{Path: configPath}
	configFile, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// init content map and auths cache if the content file does not exist
			cfg.Content = make(map[string]json.RawMessage)
			cfg.AuthsCache = make(map[string]json.RawMessage)
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to open config file at %s: %w", configPath, err)
	}
	defer configFile.Close()

	// decode config content if the config file exists
	if err := json.NewDecoder(configFile).Decode(&cfg.Content); err != nil {
		return nil, fmt.Errorf("failed to decode config file at %s: %w: %v", configPath, ErrInvalidConfigFormat, err)
	}
	authsBytes, ok := cfg.Content[configFieldAuths]
	if !ok {
		// init auths cache
		cfg.AuthsCache = make(map[string]json.RawMessage)
		return cfg, nil
	}
	if err := json.Unmarshal(authsBytes, &cfg.AuthsCache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auths field: %w: %v", ErrInvalidConfigFormat, err)
	}
	return cfg, nil
}

func (cfg config) getAuthConfig(serverAddress string) (authConfig, error) {
	cfg.rwLock.RLock()
	defer cfg.rwLock.RUnlock()

	authCfgBytes, ok := cfg.AuthsCache[serverAddress]
	if !ok {
		return authConfig{}, nil
	}
	var authCfg authConfig
	if err := json.Unmarshal(authCfgBytes, &authCfg); err != nil {
		return authConfig{}, fmt.Errorf("failed to unmarshal auth field: %w: %v", ErrInvalidConfigFormat, err)
	}
	return authCfg, nil
}

func (cfg config) putAuthConfig(serverAddress string, authCfg authConfig) error {
	cfg.rwLock.Lock()
	defer cfg.rwLock.Unlock()

	authCfgBytes, err := json.Marshal(authCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal auth field: %w", err)
	}
	cfg.AuthsCache[serverAddress] = authCfgBytes
	return cfg.saveFile()
}

func (cfg *config) deleteAuthConfig(serverAddress string) error {
	cfg.rwLock.Lock()
	defer cfg.rwLock.Unlock()

	if _, ok := cfg.AuthsCache[serverAddress]; !ok {
		// no ops
		return nil
	}
	delete(cfg.AuthsCache, serverAddress)
	return cfg.saveFile()
}

func (cfg *config) isAuthConfigured() bool {
	return cfg.CredentialsStore != "" ||
		len(cfg.CredentialHelpers) > 0 ||
		len(cfg.AuthsCache) > 0
}

func (cfg *config) saveFile() (returnErr error) {
	// marshal content
	authsBytes, err := json.Marshal(cfg.AuthsCache)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}
	cfg.Content[configFieldAuths] = authsBytes
	jsonBytes, err := json.MarshalIndent(cfg.Content, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// write the content to a ingest file for atomicity
	configDir := filepath.Dir(cfg.Path)
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
	if err := os.Rename(ingest, cfg.Path); err != nil {
		return fmt.Errorf("failed to save config file: %w", err)
	}
	return nil
}
