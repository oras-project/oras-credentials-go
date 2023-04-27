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
	"fmt"
	"net/http"

	credentials "github.com/oras-project/oras-credentials-go"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func ExampleLogin() {
	store, err := credentials.NewStore("example/path/config.json", credentials.StoreOptions{AllowPlaintextPut: true})
	if err != nil {
		panic(err)
	}
	registry, err := remote.NewRegistry("example.registry.com")
	if err != nil {
		panic(err)
	}
	cred := auth.Credential{Username: "username-example", Password: "password-example"}
	err = credentials.Login(context.Background(), store, registry, cred)
	if err != nil {
		panic(err)
	}
	fmt.Println("Login succeeded")
}

func ExampleLogout() {
	store, err := credentials.NewStore("example/path/config.json", credentials.StoreOptions{AllowPlaintextPut: true})
	if err != nil {
		panic(err)
	}
	err = credentials.Logout(context.Background(), store, "example.registry.com")
	if err != nil {
		panic(err)
	}
	fmt.Println("Logout succeeded")
}

func ExampleCredential() {
	store, err := credentials.NewStore("example/path/config.json", credentials.StoreOptions{AllowPlaintextPut: true})
	if err != nil {
		panic(err)
	}

	client := auth.Client{}
	client.Credential = credentials.Credential(store)

	request, err := http.NewRequest(http.MethodGet, "localhost:8080", nil)
	if err != nil {
		panic(err)
	}

	client.Do(request)
}
