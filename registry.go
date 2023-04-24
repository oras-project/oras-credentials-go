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

// errClientTypeUnsupported is thrown by Login() when the registry's client type
// is not supported.
var errClientTypeUnsupported = errors.New("client type not supported")

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
		return errClientTypeUnsupported
	}
	regClone.Client = &authClient
	// update credentials with the client
	authClient.Credential = auth.StaticCredential(reg.Reference.Registry, cred)
	// login and store credential
	if err := regClone.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping the registry %s: %w", regClone.Reference.Registry, err)
	}
	hostname := mapStoreRegistryName(regClone.Reference.Registry)
	if err := store.Put(ctx, hostname, cred); err != nil {
		return fmt.Errorf("unable to store the credential for %s: %w", hostname, err)
	}
	return nil
}

// Logout provides the logout functionality given the registry name.
func Logout(ctx context.Context, store Store, registryName string) error {
	registryName = mapStoreRegistryName(registryName)
	if err := store.Delete(ctx, registryName); err != nil {
		return fmt.Errorf("unable to delete the credential for %s: %w", registryName, err)
	}
	return nil
}

// Credential returns a Credential() function that can be used by auth.Client.
func Credential(store Store) func(context.Context, string) (auth.Credential, error) {
	return func(ctx context.Context, reg string) (auth.Credential, error) {
		reg = mapAuthenticationRegistryName(reg)
		if reg == "" {
			return auth.EmptyCredential, nil
		}
		return store.Get(ctx, reg)
	}
}

func mapStoreRegistryName(registry string) string {
	// The Docker CLI expects that the 'docker.io' credential
	// will be added under the key "https://index.docker.io/v1/"
	// See: https://github.com/moby/moby/blob/v24.0.0-beta.2/registry/config.go#L25-L48
	if registry == "docker.io" {
		return "https://index.docker.io/v1/"
	}
	return registry
}

func mapAuthenticationRegistryName(hostname string) string {
	// It is expected that the traffic targetting "registry-1.docker.io"
	// will be redirected to "https://index.docker.io/v1/"
	// See: https://github.com/moby/moby/blob/v24.0.0-beta.2/registry/config.go#L25-L48
	if hostname == "registry-1.docker.io" {
		return "https://index.docker.io/v1/"
	}
	return hostname
}
