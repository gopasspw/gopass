package api_test

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/gopass/api"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
)

func Example() {
	ctx := context.Background()

	gp, err := api.New(ctx)
	if err != nil {
		panic(err)
	}

	// Listing secrets
	ls, err := gp.List(ctx)
	if err != nil {
		panic(err)
	}

	for _, s := range ls {
		fmt.Printf("Secret: %s", s)
	}

	// Writing secrets
	sec := secrets.New()
	sec.SetPassword("foobar")
	if err := gp.Set(ctx, "my/new/secret", sec); err != nil {
		panic(err)
	}

	// Reading secrets
	sec, err = gp.Get(ctx, "my/new/secret", "latest")
	if err != nil {
		panic(err)
	}
	fmt.Printf("content of %s: %s\n", "my/new/secret", string(sec.Bytes()))

	// Removing a secret
	if err := gp.Remove(ctx, "my/new/secret"); err != nil {
		panic(err)
	}

	// Cleaning up
	if err := gp.Close(ctx); err != nil {
		panic(err)
	}
}
