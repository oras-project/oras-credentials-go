package credentials

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker-credential-helpers/credentials"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

var loginTestHostURL string

// loginTestCommand implements the Program interface for testing purpose.
// It simulates interactions between the docker client and a remote
// credentials helper.
type loginTestCommand struct {
	arg   string
	input io.Reader
}

// Output returns responses from the remote credentials helper.
// It mocks those responses based in the input in the mock.
func (m *loginTestCommand) Output() ([]byte, error) {
	in, err := io.ReadAll(m.input)
	if err != nil {
		return nil, err
	}
	inS := string(in)
	switch m.arg {
	case "store":
		var c credentials.Credentials
		err := json.NewDecoder(strings.NewReader(inS)).Decode(&c)
		if err != nil {
			return []byte("program failed"), errCommandExited
		}
		switch c.ServerURL {
		case loginTestHostURL:
			return nil, nil
		default:
			return []byte("program failed"), errCommandExited
		}
	}
	return []byte(fmt.Sprintf("unknown argument %q with %q", m.arg, inS)), errCommandExited
}

// Input sets the input to send to a remote credentials helper.
func (m *loginTestCommand) Input(in io.Reader) {
	m.input = in
}

func loginTestCommandFn(args ...string) client.Program {
	return &loginTestCommand{
		arg: args[0],
	}
}

func TestLogin(t *testing.T) {
	// create a test registry
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()
	host := ts.URL
	uri, _ := url.Parse(host)
	loginTestHostURL = uri.Host
	reg, err := remote.NewRegistry(loginTestHostURL)
	if err != nil {
		panic("cannot create test registry")
	}
	reg.PlainHTTP = true
	// create a test native store
	ns := &NativeStore{
		programFunc: loginTestCommandFn,
	}
	tests := []struct {
		name     string
		ctx      context.Context
		store    Store
		registry remote.Registry
		cred     auth.Credential
		wantErr  bool
	}{
		{
			name:     "login succeeds",
			ctx:      context.Background(),
			store:    ns,
			registry: *reg,
			cred:     auth.Credential{Username: testUsername, Password: testPassword},
			wantErr:  false,
		},
		{
			name:     "login fails (nil context makes Ping fails)",
			ctx:      nil,
			store:    ns,
			registry: *reg,
			cred:     auth.Credential{Username: testUsername, Password: testPassword},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Login(tt.ctx, tt.store, tt.registry, tt.cred); (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
