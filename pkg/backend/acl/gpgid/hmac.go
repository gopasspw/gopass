package gpgid

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/justwatchcom/gopass/pkg/out"
)

func (a *ACL) hmacNameFromSig(fn string) string {
	// store/.gpg-id.sig.0xDEADBEEF
	// store/.gpg-id.hmac.0xDEADBEEF
	return a.crypto.IDFile() + ".hmac." + strings.TrimPrefix(fn, a.crypto.IDFile()+".sig.")

}

func (a *ACL) computeHMAC(ctx context.Context, tok Token, hmf string, srcf string) error {
	if len(a.tokens) < 1 {
		return fmt.Errorf("need at least one token")
	}
	// always use the latest token to sign
	mac := hmac.New(sha256.New, a.tokens.Current())
	fh, err := os.Open(srcf)
	if err != nil {
		return err
	}
	defer func() {
		_ = fh.Close()
	}()

	if _, err := io.Copy(mac, fh); err != nil {
		return err
	}
	return ioutil.WriteFile(hmf, mac.Sum(nil), 0600)
}

func (a *ACL) verifyHMAC(ctx context.Context, hmf string, srcf string) (bool, error) {
	hmb, err := ioutil.ReadFile(hmf)
	if err != nil {
		return false, err
	}

	out.Debug(ctx, "ACL.verifyHMAC(%s,%s): hmb: %X", hmf, srcf, hmb)

	buf, err := ioutil.ReadFile(srcf)
	if err != nil {
		return false, err
	}
	latest := true
	// try all tokens to verify
	for i := len(a.tokens) - 1; i >= 0; i-- {
		tok := a.tokens[i]
		out.Debug(ctx, "ACL.verifyHMAC - trying token %d/%s ...", i, tok)

		mac := hmac.New(sha256.New, []byte(tok))
		_, _ = mac.Write(buf)
		expectedMAC := mac.Sum(nil)
		if hmac.Equal(expectedMAC, hmb) {
			out.Debug(ctx, "ACL.verifyHMAC - VALID")
			return latest, nil
		}
		latest = false
		out.Debug(ctx, "ACL.verifyHMAC - INVALID")
	}
	return false, fmt.Errorf("invalid HMAC")
}
