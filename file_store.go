package credentials

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"oras.land/oras-go/v2/registry/remote/auth"
)

// FileStore implements a credentials store using the docker configuration file
// to keep the credentials in plain-text.
type FileStore struct {
	DisableSave   bool
	configPath    string
	content       map[string]interface{}
	contentRWLock sync.RWMutex
}

const (
	ConfigFieldAuthConfigs   = "auths"
	ConfigFieldUsername      = "username"
	ConfigFieldPassword      = "password"
	ConfigFieldBasicAuth     = "auth"
	ConfigFieldIdentityToken = "identitytoken"
	ConfigfieldRegistryToken = "registrytoken"
)

var (
	ErrInvalidFormat         = errors.New("invalid format")
	ErrPlainTextSaveDisabled = errors.New("plain text save is disabled")
)

// authConfig contains authorization information for connecting to a Registry
type authConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Auth     string `json:"auth,omitempty"`

	// IdentityToken is used to authenticate the user and get
	// an access token for the registry.
	IdentityToken string `json:"identitytoken,omitempty"`
	// RegistryToken is a bearer token to be sent to a registry
	RegistryToken string `json:"registrytoken,omitempty"`
}

// NewFileStore creates a new file credentials store.
func NewFileStore(configPath string) (*FileStore, error) {
	fs := &FileStore{configPath: configPath}
	configFile, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// init content map if the content file does not exist
			fs.content = make(map[string]interface{})
			return fs, nil
		}
		return nil, fmt.Errorf("failed to open config file at %s: %w", configPath, err)
	}
	defer configFile.Close()

	// decode config content if the config file exists
	if err := json.NewDecoder(configFile).Decode(&fs.content); err != nil {
		return nil, fmt.Errorf("failed to decode config file at %s: %w", configPath, err)
	}
	return fs, nil
}

// Get retrieves credentials from the store for the given server address.
func (fs *FileStore) Get(_ context.Context, serverAddress string) (auth.Credential, error) {
	fs.contentRWLock.RLock()
	defer fs.contentRWLock.RUnlock()

	authCfg, ok := fs.getAuthConfig(serverAddress)
	if !ok {
		return auth.EmptyCredential, nil
	}

	cred := auth.Credential{
		Username:     authCfg.Username,
		Password:     authCfg.Password,
		RefreshToken: authCfg.IdentityToken,
		AccessToken:  authCfg.RegistryToken,
	}
	if authCfg.Auth != "" {
		var err error
		// override username and password
		cred.Username, cred.Password, err = decodeAuth(authCfg.Auth)
		if err != nil {
			return auth.EmptyCredential, fmt.Errorf("failed to decode username and password: %w: %v", ErrInvalidFormat, err)
		}
	}
	return cred, nil
}

// Put saves credentials into the store for the given server address.
// Returns ErrPlainTextSaveDisabled if s.DisableSave is set to true.
func (fs *FileStore) Put(_ context.Context, serverAddress string, cred auth.Credential) error {
	if fs.DisableSave {
		return ErrPlainTextSaveDisabled
	}

	fs.contentRWLock.Lock()
	defer fs.contentRWLock.Unlock()

	fs.updateAuths(serverAddress, cred)
	return fs.saveFile()
}

// Delete removes credentials from the store for the given server address.
func (fs *FileStore) Delete(ctx context.Context, serverAddress string) error {
	fs.contentRWLock.Lock()
	defer fs.contentRWLock.Unlock()

	if _, err := os.Stat(fs.configPath); errors.Is(err, os.ErrNotExist) {
		// no ops if the config file does not exist
		return nil
	}
	authsMap, ok := fs.content[ConfigFieldAuthConfigs].(map[string]interface{})
	if !ok {
		// no ops
		return nil
	}
	if _, ok = authsMap[serverAddress]; !ok {
		// no ops
		return nil
	}

	// update data
	delete(authsMap, serverAddress)
	fs.content[ConfigFieldAuthConfigs] = authsMap
	return fs.saveFile()
}

// getAuthConfig reads the config and returns authConfig for serverAddress.
func (fs *FileStore) getAuthConfig(serverAddress string) (authConfig, bool) {
	authsMap, ok := fs.content[ConfigFieldAuthConfigs].(map[string]interface{})
	if !ok {
		return authConfig{}, false
	}
	authConfigObj, ok := authsMap[serverAddress].(map[string]interface{})
	if !ok {
		return authConfig{}, false
	}

	var authCfg authConfig
	for k, v := range authConfigObj {
		switch k {
		case ConfigFieldUsername:
			authCfg.Username, _ = v.(string)
		case ConfigFieldPassword:
			authCfg.Password, _ = v.(string)
		case ConfigFieldBasicAuth:
			authCfg.Auth, _ = v.(string)
		case ConfigFieldIdentityToken:
			authCfg.IdentityToken, _ = v.(string)
		case ConfigfieldRegistryToken:
			authCfg.RegistryToken, _ = v.(string)
		}
	}
	return authCfg, true
}

// updateAuths updates the Auths field of fs.content based on cred.
func (fs *FileStore) updateAuths(serverAddress string, cred auth.Credential) {
	authsMap, ok := fs.content[ConfigFieldAuthConfigs].(map[string]interface{})
	if !ok {
		authsMap = make(map[string]interface{})
	}
	authCfg, ok := authsMap[serverAddress].(map[string]interface{})
	if !ok {
		authCfg = make(map[string]interface{})
	}
	authCfg[ConfigFieldBasicAuth] = encodeAuth(cred.Username, cred.Password)
	authCfg[ConfigFieldUsername] = ""
	authCfg[ConfigFieldPassword] = ""
	authCfg[ConfigFieldIdentityToken] = cred.RefreshToken
	authCfg[ConfigfieldRegistryToken] = cred.AccessToken

	// update data
	authsMap[serverAddress] = authCfg
	fs.content[ConfigFieldAuthConfigs] = authsMap
}

func (fs *FileStore) saveFile() error {
	// TODO: save to temp and copy
	// TODO: handle symlink
	// TODO: handle permissions
	jsonData, err := json.MarshalIndent(fs.content, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}
	dir := filepath.Dir(fs.configPath)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("failed to make directory: %w", err)
	}
	if err = ioutil.WriteFile(fs.configPath, jsonData, 0666); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}
	return nil
}

// encodeAuth base64-encodes username and password into base64(username:password).
func encodeAuth(username, password string) string {
	if username == "" && password == "" {
		return ""
	}

	authStr := username + ":" + password
	msg := []byte(authStr)
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(msg)))
	base64.StdEncoding.Encode(encoded, msg)
	return string(encoded)
}

// decodeAuth decodes a base64 encoded string and returns username and password.
func decodeAuth(authStr string) (username string, password string, err error) {
	if authStr == "" {
		return "", "", nil
	}

	decodedLen := base64.StdEncoding.DecodedLen(len(authStr))
	decoded := make([]byte, decodedLen)
	authByte := []byte(authStr)
	n, err := base64.StdEncoding.Decode(decoded, authByte)
	if err != nil {
		return "", "", err
	}
	if n > decodedLen {
		return "", "", errors.New("size mismatch")
	}
	arr := strings.SplitN(string(decoded), ":", 2)
	if len(arr) != 2 {
		return "", "", errors.New("auth does not conform username:password format")
	}
	password = strings.Trim(arr[1], "\x00")
	return arr[0], password, nil
}
