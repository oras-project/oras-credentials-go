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
	"sync"

	"oras.land/oras-go/v2/registry/remote/auth"
)

// InMemoryStore is a store that keeps credentials in memory.
type InMemoryStore struct {
	rwLock sync.RWMutex
	store  map[string]auth.Credential
}

// NewInMemoryStore creates a new in-memory credentials store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{store: map[string]auth.Credential{}}
}

// Get retrieves credentials from the store for the given server address.
func (is *InMemoryStore) Get(_ context.Context, serverAddress string) (auth.Credential, error) {
	is.rwLock.RLock()
	defer is.rwLock.RUnlock()

	cred, found := is.store[serverAddress]
	if !found {
		return auth.EmptyCredential, nil
	}
	return cred, nil
}

// Put saves credentials into the store for the given server address.
func (is *InMemoryStore) Put(_ context.Context, serverAddress string, cred auth.Credential) error {
	is.rwLock.Lock()
	defer is.rwLock.Unlock()

	is.store[serverAddress] = cred
	return nil
}

// Delete removes credentials from the store for the given server address.
func (is *InMemoryStore) Delete(_ context.Context, serverAddress string) error {
	is.rwLock.Lock()
	defer is.rwLock.Unlock()

	delete(is.store, serverAddress)
	return nil
}
