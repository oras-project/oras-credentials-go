package credentials

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	ConfigFieldAuths = "auths"
)

var (
	ErrCredentialNotFound = errors.New("credential not found")
)

// AuthConfig contains authorization information for connecting to a Registry
type AuthConfig struct {
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
func NewFileStore(configPath string) (Store, error) {
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	fs := &FileStore{configPath: configPath}
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
func (fs *FileStore) Put(ctx context.Context, serverAddress string, cred auth.Credential) error {
	panic("not implemented") // TODO: Implement
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

func (fs *FileStore) getAuthConfig(serverAddress string) (AuthConfig, error) {
	authsMap, ok := fs.data[ConfigFieldAuths].(map[string]interface{})
	if !ok {
		return AuthConfig{}, ErrCredentialNotFound
	}
	authConfigMap, ok := authsMap[serverAddress].(map[string]interface{})
	if !ok {
		return AuthConfig{}, ErrCredentialNotFound
	}

	var authConfig AuthConfig
	for k, v := range authConfigMap {
		switch k {
		case "auth":
			authConfig.Auth = v.(string)
		case "identitytoken":
			authConfig.IdentityToken = v.(string)
		case "registrytoken":
			authConfig.RegistryToken = v.(string)
		}
	}

	return authConfig, nil
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
