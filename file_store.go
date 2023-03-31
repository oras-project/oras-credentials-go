package credentials

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"oras.land/oras-go/v2/registry/remote/auth"
)

// FileStore implements a credentials store using the docker configuration file
// to keep the credentials in plain-text.
type FileStore struct {
	configPath  string
	data        map[string]interface{}
	DisableSave bool
}

const (
	ConfigFieldAuths         = "auths"
	ConfigFieldAuth          = "auth"
	ConfigFieldIdentityToken = "identitytoken"
	ConfigfieldRegistryToken = "registrytoken"
)

var (
	ErrCredentialNotFound = errors.New("credential not found")
	ErrInvalidFormat      = errors.New("invalid format")
)

// authConfig contains authorization information for connecting to a Registry
type authConfig struct {
	Auth string `json:"auth,omitempty"`
	// IdentityToken is used to authenticate the user and get
	// an access token for the registry.
	IdentityToken string `json:"identitytoken,omitempty"`
	// RegistryToken is a bearer token to be sent to a registry
	RegistryToken string `json:"registrytoken,omitempty"`
}

// NewFileStore creates a new file credentials store.
func NewFileStore(configPath string) (Store, error) {
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

// Put saves credentials into the store
// TODO: concurrency?
func (fs *FileStore) Put(_ context.Context, serverAddress string, cred auth.Credential) error {
	fs.updateAuths(serverAddress, cred)
	jsonData, err := json.MarshalIndent(fs.data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}
	if err = ioutil.WriteFile(fs.configPath, jsonData, 0666); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}
	return nil
}

func (fs *FileStore) updateAuths(serverAddress string, cred auth.Credential) {
	authsMap, ok := fs.data[ConfigFieldAuths].(map[string]interface{})
	if !ok {
		authsMap = make(map[string]interface{})
	}
	authConfigObj, ok := authsMap[serverAddress].(map[string]interface{})
	if !ok {
		authConfigObj = make(map[string]interface{})
	}
	// TODO: patch update or overwrite?
	if cred.Username != "" || cred.Password != "" {
		authConfigObj[ConfigFieldAuth] = encodeAuth(cred.Username, cred.Password)
	}
	if cred.RefreshToken != "" {
		authConfigObj[ConfigFieldIdentityToken] = cred.RefreshToken
	}
	if cred.AccessToken != "" {
		authConfigObj[ConfigfieldRegistryToken] = cred.AccessToken
	}

	// update data
	authsMap[serverAddress] = authConfigObj
	fs.data[ConfigFieldAuth] = authsMap
}

// Delete removes credentials from the store for the given server
func (fs *FileStore) Delete(ctx context.Context, serverAddress string) error {
	panic("not implemented") // TODO: Implement
}

// Get retrieves credentials from the store for the given server
func (fs *FileStore) Get(_ context.Context, serverAddress string) (auth.Credential, error) {
	cred := auth.Credential{}
	authConfig, err := fs.getAuthConfig(serverAddress)
	if err != nil {
		return auth.EmptyCredential, err
	}
	cred.Username, cred.Password, err = decodeAuth(authConfig.Auth)
	if err != nil {
		return auth.EmptyCredential, err
	}
	cred.RefreshToken = authConfig.IdentityToken
	cred.AccessToken = authConfig.RegistryToken
	return cred, nil
}

func (fs *FileStore) getAuthConfig(serverAddress string) (authConfig, error) {
	authsMap, ok := fs.data[ConfigFieldAuths].(map[string]interface{})
	if !ok {
		return authConfig{}, ErrCredentialNotFound
	}
	authConfigObj, ok := authsMap[serverAddress].(map[string]interface{})
	if !ok {
		return authConfig{}, ErrCredentialNotFound
	}

	var authConfig authConfig
	for k, v := range authConfigObj {
		switch k {
		case ConfigFieldAuth:
			authConfig.Auth = v.(string)
		case ConfigFieldIdentityToken:
			authConfig.IdentityToken = v.(string)
		case ConfigfieldRegistryToken:
			authConfig.RegistryToken = v.(string)
		}
	}
	return authConfig, nil
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
