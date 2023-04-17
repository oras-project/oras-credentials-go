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

	"oras.land/oras-go/v2/registry/remote/auth"
)

// StoreWithFallbacks is a store that has multiple fallback stores.
type StoreWithFallbacks struct {
	primaryStore   Store
	fallbackStores []Store
}

// NewStoreWithFallbacks returns a new store based on the given stores.
// The second and the subsequent stores will be used as fallbacks for the first store.
func NewStoreWithFallbacks(store Store, fallbacks ...Store) Store {
	return &StoreWithFallbacks{
		primaryStore:   store,
		fallbackStores: fallbacks,
	}
}

// Get retrieves credentials from the store for the given server.
func (sf *StoreWithFallbacks) Get(ctx context.Context, serverAddress string) (auth.Credential, error) {
	cred, err := sf.primaryStore.Get(ctx, serverAddress)
	if err != nil {
		return auth.EmptyCredential, err
	}
	if cred != auth.EmptyCredential {
		return cred, nil
	}
	for _, s := range sf.fallbackStores {
		cred, err = s.Get(ctx, serverAddress)
		if err != nil {
			return auth.EmptyCredential, err
		}
		if cred != auth.EmptyCredential {
			return cred, nil
		}
	}
	return auth.EmptyCredential, nil
}

// Put saves credentials into the store.
func (sf *StoreWithFallbacks) Put(ctx context.Context, serverAddress string, cred auth.Credential) error {
	return sf.primaryStore.Put(ctx, serverAddress, cred)
}

// Delete removes credentials from the store for the given server.
func (sf *StoreWithFallbacks) Delete(ctx context.Context, serverAddress string) error {
	return sf.primaryStore.Delete(ctx, serverAddress)
}
