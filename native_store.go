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
	orascreds "oras.land/oras-go/v2/registry/remote/credentials"
)

// NewNativeStore creates a new native store that uses a remote helper program to
// manage credentials.
//
// The argument of NewNativeStore can be the native keychains
// ("wincred" for Windows, "pass" for linux and "osxkeychain" for macOS),
// or any program that follows the docker-credentials-helper protocol.
//
// Reference:
//   - https://docs.docker.com/engine/reference/commandline/login#credentials-store
//
// Deprecated: This function is deprecated.
// The same functionality is now provided by oras-go.
func NewNativeStore(helperSuffix string) Store {
	return orascreds.NewNativeStore(helperSuffix)
}

// NewDefaultNativeStore returns a native store based on the platform-default
// docker credentials helper and a bool indicating if the native store is
// available.
//   - Windows: "wincred"
//   - Linux: "pass" or "secretservice"
//   - macOS: "osxkeychain"
//
// Reference:
//   - https://docs.docker.com/engine/reference/commandline/login/#credentials-store
//
// Deprecated: This function is deprecated.
// The same functionality is now provided by oras-go.
func NewDefaultNativeStore() (Store, bool) {
	return orascreds.NewDefaultNativeStore()
}
