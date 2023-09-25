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
	credentials "oras.land/oras-go/v2/registry/remote/credentials"
)

// ErrClientTypeUnsupported is thrown by Login() when the registry's client type
// is not supported.
//
// Deprecated: This type is now simply [credentials.ErrClientTypeUnsupported] of oras-go.
//
// [credentials.ErrClientTypeUnsupported]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#ErrClientTypeUnsupported
var ErrClientTypeUnsupported = credentials.ErrClientTypeUnsupported

// Login provides the login functionality with the given credentials. The target
// registry's client should be nil or of type *auth.Client. Login uses
// a client local to the function and will not modify the original client of
// the registry.
//
// Deprecated: This funciton now simply calls [credentials.Login] of oras-go.
//
// [credentials.Login]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#Login
func Login(ctx context.Context, store Store, reg *remote.Registry, cred auth.Credential) error {
	return credentials.Login(ctx, store, reg, cred)
}

// Logout provides the logout functionality given the registry name.
//
// Deprecated: This funciton now simply calls [credentials.Logout] of oras-go.
//
// [credentials.Logout]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#Logout
func Logout(ctx context.Context, store Store, registryName string) error {
	return credentials.Logout(ctx, store, registryName)
}

// Credential returns a Credential() function that can be used by auth.Client.
//
// Deprecated: This funciton now simply calls [credentials.Credential] of oras-go.
//
// [credentials.Credential]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#Credential
func Credential(store Store) func(context.Context, string) (auth.Credential, error) {
	return credentials.Credential(store)
}

// ServerAddressFromRegistry maps a registry to a server address, which is used as
// a key for credentials store. The Docker CLI expects that the credentials of
// the registry 'docker.io' will be added under the key "https://index.docker.io/v1/".
// See: https://github.com/moby/moby/blob/v24.0.2/registry/config.go#L25-L48
//
// Deprecated: This funciton now simply calls [credentials.ServerAddressFromRegistry] of oras-go.
//
// [credentials.ServerAddressFromRegistry]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#ServerAddressFromRegistry
func ServerAddressFromRegistry(registry string) string {
	return credentials.ServerAddressFromRegistry(registry)
}

// ServerAddressFromHostname maps a hostname to a server address, which is used as
// a key for credentials store. It is expected that the traffic targetting the
// host "registry-1.docker.io" will be redirected to "https://index.docker.io/v1/".
// See: https://github.com/moby/moby/blob/v24.0.2/registry/config.go#L25-L48
//
// Deprecated: This funciton now simply calls [credentials.ServerAddressFromHostname] of oras-go.
//
// [credentials.ServerAddressFromHostname]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#ServerAddressFromHostname
func ServerAddressFromHostname(hostname string) string {
	return credentials.ServerAddressFromHostname(hostname)
}
