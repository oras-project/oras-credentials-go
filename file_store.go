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
	DisableSave bool
	configPath  string
	data        map[string]interface{}
	dataLock    sync.RWMutex
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

// TODO: parse username and password
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
			fs.data = make(map[string]interface{})
			return fs, nil
		}
		return nil, err
	}
	defer configFile.Close()

	jsonObj, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(jsonObj, &fs.data); err != nil {
		return nil, err
	}

	return fs, nil
}

// Get retrieves credentials from the store for the given server
func (fs *FileStore) Get(_ context.Context, serverAddress string) (auth.Credential, error) {
	fs.dataLock.RLock()
	defer fs.dataLock.RUnlock()

	authCfg := fs.getAuthConfig(serverAddress)
	cred := auth.Credential{
		Username:     authCfg.Username,
		Password:     authCfg.Password,
		RefreshToken: authCfg.IdentityToken,
		AccessToken:  authCfg.RegistryToken,
	}
	if authCfg.Auth != "" {
		var err error
		cred.Username, cred.Password, err = decodeAuth(authCfg.Auth)
		if err != nil {
			return auth.EmptyCredential, fmt.Errorf("failed to decode username and password: %w: %v", ErrInvalidFormat, err)
		}
	}
	return cred, nil
}

// Put saves credentials into the store
func (fs *FileStore) Put(_ context.Context, serverAddress string, cred auth.Credential) error {
	if fs.DisableSave {
		return ErrPlainTextSaveDisabled
	}

	fs.dataLock.Lock()
	defer fs.dataLock.Unlock()

	fs.updateAuths(serverAddress, cred)
	return fs.saveFile()
}

// Delete removes credentials from the store for the given server
func (fs *FileStore) Delete(ctx context.Context, serverAddress string) error {
	fs.dataLock.Lock()
	defer fs.dataLock.Unlock()

	authsMap, ok := fs.data[ConfigFieldAuthConfigs].(map[string]interface{})
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
	fs.data[ConfigFieldAuthConfigs] = authsMap
	// TODO: create config or not if not exist?
	return fs.saveFile()
}

func (fs *FileStore) updateAuths(serverAddress string, cred auth.Credential) {
	authsMap, ok := fs.data[ConfigFieldAuthConfigs].(map[string]interface{})
	if !ok {
		authsMap = make(map[string]interface{})
	}
	authConfigObj, ok := authsMap[serverAddress].(map[string]interface{})
	if !ok {
		authConfigObj = make(map[string]interface{})
	}
	// TODO: patch update or overwrite?
	if cred.Username != "" || cred.Password != "" {
		authConfigObj[ConfigFieldBasicAuth] = encodeAuth(cred.Username, cred.Password)
	}
	if cred.RefreshToken != "" {
		authConfigObj[ConfigFieldIdentityToken] = cred.RefreshToken
	}
	if cred.AccessToken != "" {
		authConfigObj[ConfigfieldRegistryToken] = cred.AccessToken
	}

	// update data
	authsMap[serverAddress] = authConfigObj
	fs.data[ConfigFieldAuthConfigs] = authsMap
}

func (fs *FileStore) getAuthConfig(serverAddress string) authConfig {
	authsMap, ok := fs.data[ConfigFieldAuthConfigs].(map[string]interface{})
	if !ok {
		return authConfig{}
	}
	authConfigObj, ok := authsMap[serverAddress].(map[string]interface{})
	if !ok {
		return authConfig{}
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
	return authCfg
}

func (fs *FileStore) saveFile() error {
	// TODO: save to temp and copy
	// TODO: handle symlink
	// TODO: handle permissions
	jsonData, err := json.MarshalIndent(fs.data, "", "\t")
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

// encodeAuth creates a base64 encoded string to containing authorization information
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

// decodeAuth decodes a base64 encoded string and returns username and password
func decodeAuth(authStr string) (string, string, error) {
	if authStr == "" {
		return "", "", nil
	}

	decLen := base64.StdEncoding.DecodedLen(len(authStr))
	decoded := make([]byte, decLen)
	authByte := []byte(authStr)
	n, err := base64.StdEncoding.Decode(decoded, authByte)
	if err != nil {
		return "", "", err
	}
	if n > decLen {
		return "", "", errors.New("something went wrong decoding auth config")
	}
	arr := strings.SplitN(string(decoded), ":", 2)
	if len(arr) != 2 {
		return "", "", errors.New("invalid auth configuration file")
	}
	password := strings.Trim(arr[1], "\x00")
	return arr[0], password, nil
}
