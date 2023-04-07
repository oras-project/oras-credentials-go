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
	"testing"

	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// testStore implements the Store interface, used for testing purpose.
type testStore struct{}

func (t *testStore) Get(ctx context.Context, serverAddress string) (auth.Credential, error) {
	return auth.Credential{}, nil
}

func (t *testStore) Put(ctx context.Context, serverAddress string, cred auth.Credential) error {
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
	reg, err := remote.NewRegistry(uri.Host)
	if err != nil {
		panic("cannot create test registry")
	}
	reg.PlainHTTP = true

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
			store:    ns,
			registry: *reg,
			cred:     auth.Credential{Username: testUsername, Password: testPassword},
			wantErr:  false,
		},
		{
			name:     "login fails (nil context makes remote.Ping fails)",
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
