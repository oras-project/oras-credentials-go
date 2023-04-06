package credentials

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func Login(ctx context.Context, store Store, registry remote.Registry, cred auth.Credential) error {
	if err := registry.Ping(ctx); err != nil {
		return fmt.Errorf("unable to connect to the registry: %v", err)
	}
	if err := store.Put(ctx, registry.Reference.Registry, cred); err != nil {
		return fmt.Errorf("unable to store the credential: %v", err)
	}
	fmt.Println("Login Succeeded")
	return nil
}
