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
	"context"
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

// FileStore implements a credentials store using the docker configuration file
// to keep the credentials in plain-text.
type FileStore struct {
	// DisableSave disables saving credentials in plaintext.
	// If DisableSave is set to true, Put() will return ErrPlaintextSaveDisabled.
	DisableSave bool

	// configPath is the path to the config file.
	configPath string
	// content is the content of the config file.
	content map[string]json.RawMessage
	// authsCache is a cache of the auths field of the config field.
	authsCache map[string]json.RawMessage
	// rwLock is a read-write-lock for the file store.
	rwLock sync.RWMutex
}

const configFieldAuths = "auths"

var (
	// ErrInvalidConfigFormat is returned when the config format is invalid.
	ErrInvalidConfigFormat = errors.New("invalid config format")
	// ErrPlaintextSaveDisabled is returned by Put() when DisableSave is set
	// to true.
	ErrPlaintextSaveDisabled = errors.New("plaintext save is disabled")
)

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

// newAuthConfig creates an authConfig based on cred.
func newAuthConfig(cred auth.Credential) authConfig {
	return authConfig{
		Auth:          encodeAuth(cred.Username, cred.Password),
		IdentityToken: cred.RefreshToken,
		RegistryToken: cred.AccessToken,
	}
}

// Credential returns an auth.Credential based on ac.
func (ac authConfig) Credential() (auth.Credential, error) {
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

// NewFileStore creates a new file credentials store.
func NewFileStore(configPath string) (*FileStore, error) {
	fs := &FileStore{configPath: configPath}
	configFile, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// init content map and auths cache if the content file does not exist
			fs.content = make(map[string]json.RawMessage)
			fs.authsCache = make(map[string]json.RawMessage)
			return fs, nil
		}
		return nil, fmt.Errorf("failed to open config file at %s: %w", configPath, err)
	}
	defer configFile.Close()

	// decode config content if the config file exists
	if err := json.NewDecoder(configFile).Decode(&fs.content); err != nil {
		return nil, fmt.Errorf("failed to decode config file at %s: %w: %v", configPath, ErrInvalidConfigFormat, err)
	}
	authsBytes, ok := fs.content[configFieldAuths]
	if !ok {
		// init auths cache
		fs.authsCache = make(map[string]json.RawMessage)
		return fs, nil
	}
	if err := json.Unmarshal(authsBytes, &fs.authsCache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auths field: %w: %v", ErrInvalidConfigFormat, err)
	}
	return fs, nil
}

// Get retrieves credentials from the store for the given server address.
func (fs *FileStore) Get(_ context.Context, serverAddress string) (auth.Credential, error) {
	fs.rwLock.RLock()
	defer fs.rwLock.RUnlock()

	authCfgBytes, ok := fs.authsCache[serverAddress]
	if !ok {
		return auth.EmptyCredential, nil
	}
	var authCfg authConfig
	if err := json.Unmarshal(authCfgBytes, &authCfg); err != nil {
		return auth.EmptyCredential, fmt.Errorf("failed to unmarshal auth field: %w: %v", ErrInvalidConfigFormat, err)
	}
	return authCfg.Credential()
}

// Put saves credentials into the store for the given server address.
// Returns ErrPlaintextSaveDisabled if fs.DisableSave is set to true.
func (fs *FileStore) Put(_ context.Context, serverAddress string, cred auth.Credential) error {
	if fs.DisableSave {
		return ErrPlaintextSaveDisabled
	}

	fs.rwLock.Lock()
	defer fs.rwLock.Unlock()

	authCfg := newAuthConfig(cred)
	authCfgBytes, err := json.Marshal(authCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal auth field: %w", err)
	}
	fs.authsCache[serverAddress] = authCfgBytes
	return fs.saveFile()
}

// Delete removes credentials from the store for the given server address.
func (fs *FileStore) Delete(ctx context.Context, serverAddress string) error {
	fs.rwLock.Lock()
	defer fs.rwLock.Unlock()

	if _, ok := fs.authsCache[serverAddress]; !ok {
		// no ops
		return nil
	}
	delete(fs.authsCache, serverAddress)
	return fs.saveFile()
}

// saveFile saves fs.content into fs.configPath.
func (fs *FileStore) saveFile() error {
	if len(fs.authsCache) > 0 {
		authsBytes, err := json.Marshal(fs.authsCache)
		if err != nil {
			return fmt.Errorf("failed to marshal credentials: %w", err)
		}
		fs.content[configFieldAuths] = authsBytes
	} else {
		// omit empty
		delete(fs.content, configFieldAuths)
	}
	jsonBytes, err := json.MarshalIndent(fs.content, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// write the content to a temp file first for atomicity
	configDir := filepath.Dir(fs.configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to make directory %s: %w", configDir, err)
	}
	ingest, err := ioutil.Ingest(configDir, bytes.NewReader(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to save config file: %w", err)
	}

	// handle symlink
	targetPath := fs.configPath
	if link, err := os.Readlink(fs.configPath); err == nil {
		targetPath = link
	}
	// copy file with original ownership and permissions
	ioutil.CopyFilePermissions(targetPath, ingest)
	if err := os.Rename(ingest, targetPath); err != nil {
		// clean up the ingest file
		os.Remove(ingest)
		return fmt.Errorf("failed to save config file: %w", err)
	}
	return nil
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
