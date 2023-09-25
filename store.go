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

// Deprecated: This package is deprecated.
// The same functionality is now provided by [oras.land/oras-go/v2/registry/remote/credentials].
//
// [oras.land/oras-go/v2/registry/remote/credentials]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials
package credentials

import (
	"oras.land/oras-go/v2/registry/remote/credentials"
)

// Store is the interface that any credentials store must implement.
//
// Deprecated: This type is now simply [credentials.Store] of oras-go.
//
// [credentials.Store]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#Store
type Store = credentials.Store

// DynamicStore dynamically determines which store to use based on the settings
// in the config file.
//
// Deprecated: This type is now simply [credentials.DynamicStore] of oras-go.
//
// [credentials.DynamicStore]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#DynamicStore
type DynamicStore = credentials.DynamicStore

// StoreOptions provides options for NewStore.
//
// Deprecated: This type is now simply [credentials.StoreOptions] of oras-go.
//
// [credentials.StoreOptions]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#StoreOptions
type StoreOptions = credentials.StoreOptions

// NewStore returns a Store based on the given configuration file.
//
// For Get(), Put() and Delete(), the returned Store will dynamically determine
// which underlying credentials store to use for the given server address.
// The underlying credentials store is determined in the following order:
//  1. Native server-specific credential helper
//  2. Native credentials store
//  3. The plain-text config file itself
//
// References:
//   - https://docs.docker.com/engine/reference/commandline/login/#credentials-store
//   - https://docs.docker.com/engine/reference/commandline/cli/#docker-cli-configuration-file-configjson-properties
//
// Deprecated: This funciton now simply calls [credentials.NewStore] of oras-go.
//
// [credentials.NewStore]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#NewStore
func NewStore(configPath string, opts StoreOptions) (*DynamicStore, error) {
	return credentials.NewStore(configPath, opts)
}

// NewStoreFromDocker returns a Store based on the default docker config file.
//   - If the $DOCKER_CONFIG environment variable is set,
//     $DOCKER_CONFIG/config.json will be used.
//   - Otherwise, the default location $HOME/.docker/config.json will be used.
//
// NewStoreFromDocker internally calls [NewStore].
//
// References:
//   - https://docs.docker.com/engine/reference/commandline/cli/#configuration-files
//   - https://docs.docker.com/engine/reference/commandline/cli/#change-the-docker-directory
//
// Deprecated: This funciton now simply calls [credentials.NewStoreFromDocker] of oras-go.
//
// [credentials.NewStoreFromDocker]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#NewStoreFromDocker
func NewStoreFromDocker(opts StoreOptions) (*DynamicStore, error) {
	return credentials.NewStoreFromDocker(opts)
}

// NewStoreWithFallbacks returns a new store based on the given stores.
//   - Get() searches the primary and the fallback stores
//     for the credentials and returns when it finds the
//     credentials in any of the stores.
//   - Put() saves the credentials into the primary store.
//   - Delete() deletes the credentials from the primary store.
//
// Deprecated: This funciton now simply calls [credentials.NewStoreWithFallbacks] of oras-go.
//
// [credentials.NewStoreWithFallbacks]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#NewStoreWithFallbacks
func NewStoreWithFallbacks(primary Store, fallbacks ...Store) Store {
	return credentials.NewStoreWithFallbacks(primary, fallbacks...)
}
