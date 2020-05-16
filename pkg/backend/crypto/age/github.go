package age

import (
	"context"
	"fmt"
)

func (a *Age) getPublicKeysGithub(ctx context.Context, user string) ([]string, error) {
	// TODO: recheck SoT if cache is too old
	pk, err := a.ghCache.Get(user)
	if err != nil {
		return nil, err
	}
	if len(pk) > 0 {
		return pk, nil
	}

	kl, _, err := a.ghc.Users.ListKeys(ctx, user, nil)
	if err != nil {
		return nil, err
	}
	if len(kl) < 1 {
		return nil, fmt.Errorf("not found")
	}
	keys := make([]string, 0, len(kl))
	for _, k := range kl {
		keys = append(keys, k.GetKey())
	}
	a.ghCache.Set(user, keys)
	return keys, nil
}
