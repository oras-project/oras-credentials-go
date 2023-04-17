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
	"reflect"
	"testing"

	"oras.land/oras-go/v2/registry/remote/auth"
)

func TestStoreWithFallbacks(t *testing.T) {
	ns := &NativeStore{
		programFunc: testCommandFn,
	}
	fs, err := NewFileStore("testdata/valid_auths_config.json")
	if err != nil {
		t.Fatal("NewFileStore() error =", err)
	}
	sf := NewStoreWithFallbacks(ns, fs)

	// Put
	err = sf.Put(context.Background(), "localhost:2333", auth.Credential{Username: testUsername, Password: testPassword})
	if err != nil {
		t.Fatal("sf.Put() error =", err)
	}
	// Get an entry stored in the primary store
	cred, err := sf.Get(context.Background(), "localhost:2333")
	if err != nil {
		t.Fatal("sf.Get() error =", err)
	}
	if !reflect.DeepEqual(cred, auth.Credential{Username: testUsername, Password: testPassword}) {
		t.Fatal("incorrect credential from the primary store")
	}
	// Get an entry stored in the fallback store
	cred, err = sf.Get(context.Background(), "registry2.example.com")
	if err != nil {
		t.Fatal("sf.Get() error =", err)
	}
	if !reflect.DeepEqual(cred, auth.Credential{RefreshToken: "identity_token"}) {
		t.Fatal("incorrect credential from the backup store")
	}
	// Delete
	err = sf.Delete(context.Background(), "localhost:2333")
	if err != nil {
		t.Fatal("sf.Delete() error =", err)
	}
}
