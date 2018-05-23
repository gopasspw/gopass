package gpgid

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
)

// Recipients returns a sorted list of all existing recipents
func (a *ACL) Recipients() []string {
	keys := make([]string, 0, len(a.recps))
	for k := range a.recps {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Add adds an new recipient to the ACL
func (a *ACL) Add(ctx context.Context, recp string) error {
	if recp == "" {
		return fmt.Errorf("invalid recipient")
	}
	if a.recps == nil {
		a.recps = make(map[string]struct{}, 1)
	}
	// recipient already in the list, nothing to do
	if _, found := a.recps[recp]; found {
		return nil
	}
	// add recipient
	a.recps[recp] = struct{}{}
	// rotate root token
	a.tokens = append(a.tokens, NewToken())
	// encrypt, sign and save
	return a.save(ctx)
}

// Remove removes an existing recipient from the ACL
func (a *ACL) Remove(ctx context.Context, recp string) error {
	if recp == "" {
		return fmt.Errorf("invalid recipient")
	}
	if a.recps == nil {
		return nil
	}
	// recipient not in the list, nothing to do
	if _, found := a.recps[recp]; !found {
		return nil
	}
	// remove recipient
	delete(a.recps, recp)
	// rotate root token
	a.tokens = append(a.tokens, NewToken())
	// encrypt, sign and save
	return a.save(ctx)
}

func (a *ACL) unmarshal(ctx context.Context) error {
	if a.recps == nil {
		a.recps = make(map[string]struct{}, 1)
	}

	idfBuf, err := a.fs.Get(ctx, a.crypto.IDFile())
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(idfBuf))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			// deduplicate
			a.recps[line] = struct{}{}
		}
	}
	return nil
}

// marshal all in memory Recipients line by line to []byte.
func (a *ACL) marshal(ctx context.Context) error {
	// sort
	keys := make([]string, 0, len(a.recps))
	for k := range a.recps {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := &bytes.Buffer{}
	for _, k := range keys {
		_, _ = buf.WriteString(k)
		_, _ = buf.WriteString("\n")
	}
	if err := a.fs.Set(ctx, a.crypto.IDFile(), buf.Bytes()); err != nil {
		return err
	}

	return nil
}
