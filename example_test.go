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

package credentials_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	credentials "github.com/oras-project/oras-credentials-go"
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
	delete(t.storage, serverAddress)
	return nil
}

var exampleRegistry *remote.Registry

func TestMain(m *testing.M) {
	// create a test registry
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantedAuthHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte("test_username"+":"+"test_password"))
		authHeader := r.Header.Get("Authorization")
		if authHeader != wantedAuthHeader {
			w.Header().Set("Www-Authenticate", `Basic realm="Test Server"`)
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer ts.Close()
	uri, _ := url.Parse(ts.URL)
	exampleRegistry, err := remote.NewRegistry(uri.Host)
	if err != nil {
		panic(err)
	}
	exampleRegistry.PlainHTTP = true
	os.Exit(m.Run())
}

func ExampleLogin_withBasicAuth() {
	// create a test store
	s := &testStore{}
	basicAuthCredentials := auth.Credential{Username: "test_username", Password: "test_password"}
	err := credentials.Login(context.Background(), s, exampleRegistry, basicAuthCredentials)
	if err != nil {
		panic(err) // Handle error
	}
	fmt.Println("Login succeeded")
	// Output:
	// Login succeeded
}
