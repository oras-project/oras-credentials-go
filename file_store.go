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

// FileStore implements a credentials store using the docker configuration file
// to keep the credentials in plain-text.
//
// Reference: https://docs.docker.com/engine/reference/commandline/cli/#docker-cli-configuration-file-configjson-properties
//
// Deprecated: This type is now simply [credentials.FileStore] of oras-go.
//
// [credentials.FileStore]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#FileStore
type FileStore = orascreds.FileStore

var (
	// ErrPlaintextPutDisabled is returned by Put() when DisablePut is set
	// to true.
	//
	// Deprecated: This type is now simply [credentials.ErrPlaintextPutDisabled] of oras-go.
	//
	// [credentials.ErrPlaintextPutDisabled]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#ErrPlaintextPutDisabled
	ErrPlaintextPutDisabled = orascreds.ErrPlaintextPutDisabled
	// ErrBadCredentialFormat is returned by Put() when the credential format
	// is bad.
	//
	// Deprecated: This type is now simply [credentials.ErrBadCredentialFormat] of oras-go.
	//
	// [credentials.ErrBadCredentialFormat]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#ErrBadCredentialFormat
	ErrBadCredentialFormat = orascreds.ErrBadCredentialFormat
)

// NewFileStore creates a new file credentials store.
//
// Reference: https://docs.docker.com/engine/reference/commandline/cli/#docker-cli-configuration-file-configjson-properties
//
// Deprecated: This funciton now simply calls [credentials.NewFileStore] of oras-go.
//
// [credentials.NewFileStore]: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/credentials#NewFileStore
func NewFileStore(configPath string) (*FileStore, error) {
	return orascreds.NewFileStore(configPath)
}
