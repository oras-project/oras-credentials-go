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
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	basicAuthHost     = "localhost:2333"
	bearerAuthHost    = "localhost:6666"
	exeErrorHost      = "localhost:500/exeError"
	jsonErrorHost     = "localhost:500/jsonError"
	noCredentialsHost = "localhost:404"
	testUsername      = "test_username"
	testPassword      = "test_password"
	testRefreshToken  = "test_token"
)

var (
	errCommandExited       = fmt.Errorf("exited with error")
	errExecute             = fmt.Errorf("Execute failed")
	errCredentialsNotFound = fmt.Errorf(errCredentialsNotFoundMessage)
)

// testExecuter implements the Executer interface for testing purpose.
// It simulates interactions between the docker client and a remote
// credentials helper.
type testExecuter struct{}

// Execute mocks the behavior of a credential helper binary. It returns responses
// and errors based on the input.
func (e *testExecuter) Execute(ctx context.Context, input io.Reader, action string) ([]byte, error) {
	in, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}
	inS := string(in)
	switch action {
	case "get":
		switch inS {
		case basicAuthHost:
			return []byte(`{"Username": "test_username", "Secret": "test_password"}`), nil
		case bearerAuthHost:
			return []byte(`{"Username": "<token>", "Secret": "test_token"}`), nil
		case exeErrorHost:
			return []byte("Execute failed"), errExecute
		case jsonErrorHost:
			return []byte("json.Unmarshal failed"), nil
		case noCredentialsHost:
			return []byte("credentials not found"), errCredentialsNotFound
		default:
			return []byte("program failed"), errCommandExited
		}
	case "store":
		var c dockerCredentials
		err := json.NewDecoder(strings.NewReader(inS)).Decode(&c)
		if err != nil {
			return []byte("program failed"), errCommandExited
		}
		switch c.ServerURL {
		case basicAuthHost, bearerAuthHost, exeErrorHost:
			return nil, nil
		default:
			return []byte("program failed"), errCommandExited
		}
	case "erase":
		switch inS {
		case basicAuthHost, bearerAuthHost:
			return nil, nil
		default:
			return []byte("program failed"), errCommandExited
		}
	}
	return []byte(fmt.Sprintf("unknown argument %q with %q", action, inS)), errCommandExited
}

func TestNativeStore_interface(t *testing.T) {
	var ns interface{} = &nativeStore{}
	if _, ok := ns.(Store); !ok {
		t.Error("&NativeStore{} does not conform Store")
	}
}

func TestNativeStore_basicAuth(t *testing.T) {
	ns := &nativeStore{
		&testExecuter{},
	}
	// Put
	err := ns.Put(context.Background(), basicAuthHost, auth.Credential{Username: testUsername, Password: testPassword})
	if err != nil {
		t.Fatalf("basic auth test ns.Put fails: %v", err)
	}
	// Get
	cred, err := ns.Get(context.Background(), basicAuthHost)
	if err != nil {
		t.Fatalf("basic auth test ns.Get fails: %v", err)
	}
	if cred.Username != testUsername {
		t.Fatal("incorrect username")
	}
	if cred.Password != testPassword {
		t.Fatal("incorrect password")
	}
	// Delete
	err = ns.Delete(context.Background(), basicAuthHost)
	if err != nil {
		t.Fatalf("basic auth test ns.Delete fails: %v", err)
	}
}

func TestNativeStore_refreshToken(t *testing.T) {
	ns := &nativeStore{
		&testExecuter{},
	}
	// Put
	err := ns.Put(context.Background(), bearerAuthHost, auth.Credential{RefreshToken: testRefreshToken})
	if err != nil {
		t.Fatalf("refresh token test ns.Put fails: %v", err)
	}
	// Get
	cred, err := ns.Get(context.Background(), bearerAuthHost)
	if err != nil {
		t.Fatalf("refresh token test ns.Get fails: %v", err)
	}
	if cred.Username != "" {
		t.Fatalf("expect username to be empty, got %s", cred.Username)
	}
	if cred.RefreshToken != testRefreshToken {
		t.Fatal("incorrect refresh token")
	}
	// Delete
	err = ns.Delete(context.Background(), basicAuthHost)
	if err != nil {
		t.Fatalf("refresh token test ns.Delete fails: %v", err)
	}
}

func TestNativeStore_errorHandling(t *testing.T) {
	ns := &nativeStore{
		&testExecuter{},
	}
	// Get Error: Execute error
	_, err := ns.Get(context.Background(), exeErrorHost)
	if err != errExecute {
		t.Fatalf("got error: %v, should get exeErr", err)
	}
	// Get Error: json.Unmarshal
	_, err = ns.Get(context.Background(), jsonErrorHost)
	if err == nil {
		t.Fatalf("should get error from json.Unmarshal")
	}
	// Get: Should not return error when credentials are not found
	_, err = ns.Get(context.Background(), noCredentialsHost)
	if err != nil {
		t.Fatalf("should not get error when no credentials are found")
	}
}

func TestNewDefaultNativeStore(t *testing.T) {
	defaultHelper := getDefaultHelperSuffix()
	wantOK := (defaultHelper != "")

	if _, ok := NewDefaultNativeStore(); ok != wantOK {
		t.Errorf("NewDefaultNativeStore() = %v, want %v", ok, wantOK)
	}
}
