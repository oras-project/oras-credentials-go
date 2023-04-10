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
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// testStore implements the Store interface, used for testing purpose.
type testStore struct {
	storage map[string]auth.Credential
}

func (t *testStore) Get(ctx context.Context, serverAddress string) (auth.Credential, error) {
	return t.storage[serverAddress], nil
}

func (t *testStore) Put(ctx context.Context, serverAddress string, cred auth.Credential) error {
	if len(t.storage) == 0 {
		t.storage = make(map[string]auth.Credential)
	}
	t.storage[serverAddress] = cred
	return nil
}

func (t *testStore) Delete(ctx context.Context, serverAddress string) error {
	return nil
}

func TestLogin(t *testing.T) {
	// create a test registry
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()
	uri, _ := url.Parse(ts.URL)
	successReg, err := remote.NewRegistry(uri.Host)
	if err != nil {
		t.Fatalf("cannot create test registry: %v", err)

	}
	successReg.PlainHTTP = true
	failureReg, err := remote.NewRegistry("test.io")
	if err != nil {
		t.Fatalf("cannot create test registry: %v", err)
	}

	// create a test store
	ns := &testStore{}
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
			registry: *successReg,
			cred:     auth.Credential{Username: testUsername, Password: testPassword},
			wantErr:  false,
		},
		{
			name:     "login fails (nil context makes remote.Ping fails)",
			ctx:      nil,
			registry: *failureReg,
			cred:     auth.Credential{Username: testUsername, Password: testPassword},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Login(tt.ctx, ns, tt.registry, tt.cred)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got := ns.storage[tt.registry.Reference.Registry]; !reflect.DeepEqual(got, tt.cred) {
				t.Errorf("Stored credential = %v, want %v", got, tt.cred)
			}
		})
	}
}

func Test_mapHostname(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{
			"map docker.io to https://index.docker.io/v1/",
			"docker.io",
			"https://index.docker.io/v1/",
		},
		{
			"do not map other host names",
			"localhost:2333",
			"localhost:2333",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapHostname(tt.host); got != tt.want {
				t.Errorf("mapHostname() = %v, want %v", got, tt.want)
			}
		})
	}
}
