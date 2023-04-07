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
	"fmt"

	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// Login provides the login functionality with the given credentials.
func Login(ctx context.Context, store Store, registry remote.Registry, cred auth.Credential) error {
	if err := registry.Ping(ctx); err != nil {
		return fmt.Errorf("unable to login to the registry %s: %w", registry.Reference.Registry, err)
	}
	if err := store.Put(ctx, registry.Reference.Registry, cred); err != nil {
		return fmt.Errorf("unable to store the credential: %v", err)
	}
	return nil
}
