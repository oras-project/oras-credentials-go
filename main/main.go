package main

import (
	"context"
	"fmt"

	credentials "github.com/oras-project/oras-credentials-go"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func main() {
	ns := credentials.NewNativeStore("osxkeychain")
	ns.Put(context.Background(), "localhost:2333", auth.Credential{Username: "goodbye", Password: "jupyter"})
	c, _ := ns.Get(context.Background(), "localhost:2333")
	fmt.Println(c)
	fmt.Println("Let me test")
}
