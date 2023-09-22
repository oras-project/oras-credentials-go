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

	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	orascreds "oras.land/oras-go/v2/registry/remote/credentials"
)

// ErrClientTypeUnsupported is thrown by Login() when the registry's client type
// is not supported.
//
// Deprecated: This type is deprecated.
// The same functionality is now provided by oras-go.
var ErrClientTypeUnsupported = orascreds.ErrClientTypeUnsupported

// Login provides the login functionality with the given credentials. The target
// registry's client should be nil or of type *auth.Client. Login uses
// a client local to the function and will not modify the original client of
// the registry.
//
// Deprecated: This function is deprecated.
// The same functionality is now provided by oras-go.
func Login(ctx context.Context, store Store, reg *remote.Registry, cred auth.Credential) error {
	return orascreds.Login(ctx, store, reg, cred)
}

// Logout provides the logout functionality given the registry name.
//
// Deprecated: This function is deprecated.
// The same functionality is now provided by oras-go.
func Logout(ctx context.Context, store Store, registryName string) error {
	return orascreds.Logout(ctx, store, registryName)
}

// Credential returns a Credential() function that can be used by auth.Client.
//
// Deprecated: This function is deprecated.
// The same functionality is now provided by oras-go.
func Credential(store Store) func(context.Context, string) (auth.Credential, error) {
	return orascreds.Credential(store)
}

// ServerAddressFromRegistry maps a registry to a server address, which is used as
// a key for credentials store. The Docker CLI expects that the credentials of
// the registry 'docker.io' will be added under the key "https://index.docker.io/v1/".
// See: https://github.com/moby/moby/blob/v24.0.2/registry/config.go#L25-L48
//
// Deprecated: This function is deprecated.
// The same functionality is now provided by oras-go.
func ServerAddressFromRegistry(registry string) string {
	return orascreds.ServerAddressFromRegistry(registry)
}

// ServerAddressFromHostname maps a hostname to a server address, which is used as
// a key for credentials store. It is expected that the traffic targetting the
// host "registry-1.docker.io" will be redirected to "https://index.docker.io/v1/".
// See: https://github.com/moby/moby/blob/v24.0.2/registry/config.go#L25-L48
//
// Deprecated: This function is deprecated.
// The same functionality is now provided by oras-go.
func ServerAddressFromHostname(hostname string) string {
	return orascreds.ServerAddressFromHostname(hostname)
}
