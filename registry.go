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
	"errors"
	"fmt"

	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// Login provides the login functionality with the given credentials. The target
// registry's client should be nil or of type *auth.Client. Login uses
// a client local to the function and will not modify the original client of
// the registry.
func Login(ctx context.Context, store Store, reg *remote.Registry, cred auth.Credential) error {
	// create a clone of the original registry for login purpose
	regClone := *reg
	// we use the original client if applicable, otherwise use a default client
	var authClient auth.Client
	if reg.Client == nil {
		authClient = *auth.DefaultClient
		authClient.Cache = nil // no cache
	} else if client, ok := reg.Client.(*auth.Client); ok {
		authClient = *client
	} else {
		return errors.New("client type not supported")
	}
	regClone.Client = &authClient
	// update credentials with the client
	authClient.Credential = auth.StaticCredential(reg.Reference.Registry, cred)
	// login and store credential
	if err := regClone.Ping(ctx); err != nil {
		return fmt.Errorf("unable to login to the registry %s: %w", regClone.Reference.Registry, err)
	}
	hostname := mapHostname(regClone.Reference.Registry)
	if err := store.Put(ctx, hostname, cred); err != nil {
		return fmt.Errorf("unable to store the credential for %s: %w", hostname, err)
	}
	return nil
}

// Logout provides the logout functionality given the registry name.
func Logout(ctx context.Context, store Store, registryName string) error {
	registryName = mapHostname(registryName)
	if err := store.Delete(ctx, registryName); err != nil {
		return fmt.Errorf("unable to logout: %v", err)
	}
	return nil
}

func mapHostname(hostname string) string {
	// The Docker CLI expects that the 'docker.io' credential
	// will be added under the key "https://index.docker.io/v1/"
	if hostname == "docker.io" {
		return "https://index.docker.io/v1/"
	}
	return hostname
}
