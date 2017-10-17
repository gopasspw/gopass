package gpgid

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/justwatchcom/gopass/pkg/backend"
	"github.com/justwatchcom/gopass/pkg/fsutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/store"
)

// ACL is an gpg-id based store ACL
type ACL struct {
	idf    string
	crypto backend.Crypto
	rcs    backend.RCS
	signed bool
	valid  bool
	recps  map[string]struct{}
	tokens TokenList
}

// Init initializes (new) store integrity / ACL structs
func Init(ctx context.Context, crypto backend.Crypto, rcs backend.RCS, idfile string) (*ACL, error) {
	a := &ACL{
		idf:    idfile,
		crypto: crypto,
		rcs:    rcs,
		signed: false,
		valid:  false,
		recps:  make(map[string]struct{}),
		tokens: make(TokenList, 0),
	}
	if err := a.unmarshal(); err != nil {
		return nil, err
	}
	if a.isSigned() {
		return nil, fmt.Errorf("already signed")
	}
	if err := a.init(ctx); err != nil {
		return a, err
	}
	return a, nil
}

// Load tries to load and verify existing ACL information
func Load(ctx context.Context, crypto backend.Crypto, rcs backend.RCS, idfile string) (*ACL, error) {
	a := &ACL{
		idf:    idfile,
		crypto: crypto,
		rcs:    rcs,
		signed: false,
		valid:  false,
		recps:  make(map[string]struct{}),
		tokens: make(TokenList, 0),
	}
	if err := a.unmarshal(); err != nil {
		return nil, err
	}
	if !a.isSigned() {
		return a, nil
	}
	if err := a.verify(ctx); err != nil {
		return a, err
	}
	return a, nil
}

// Save tries to save and sign the current ACL
func (a *ACL) Save(ctx context.Context) error {
	return a.save(ctx)
}

// SigningKeyID returns the key for signing
func (a *ACL) SigningKeyID(ctx context.Context) string {
	for k := range a.recps {
		kl, err := a.crypto.FindPrivateKeys(ctx, k)
		if err != nil || len(kl) < 1 {
			continue
		}
		return kl[0]
	}
	return ""
}

func (a *ACL) tokenfile() string {
	return a.idf + ".token"
}

func (a *ACL) isSigned() bool {
	return fsutil.IsFile(a.tokenfile())
}

func (a *ACL) save(ctx context.Context) error {
	sigID := a.SigningKeyID(ctx)
	if sigID == "" {
		return fmt.Errorf("no signing key id for %+v", a.Recipients())
	}
	if a.tokens == nil {
		return fmt.Errorf("tokens not initialized")
	}
	if a.recps == nil {
		return fmt.Errorf("recipients not initialized")
	}
	// save token file
	if err := a.marshalTokenFile(ctx); err != nil {
		return err
	}
	if err := a.rcs.Add(ctx, a.tokenfile()); err != nil {
		return err
	}
	// save recipients
	if err := a.marshal(); err != nil {
		return err
	}
	if err := a.rcs.Add(ctx, a.idf); err != nil {
		return err
	}
	// sign gpg id
	idf, err := ioutil.ReadFile(a.idf)
	if err != nil {
		return err
	}
	signed, err := a.crypto.Sign(ctx, idf)
	if err != nil {
		return err
	}
	sigfn := a.idf + ".sig." + sigID
	if err := ioutil.WriteFile(sigfn, signed, 0600); err != nil {
		return err
	}
	if err := a.rcs.Add(ctx, sigfn); err != nil {
		return err
	}
	// hmac
	hmacfn := a.hmacNameFromSig(sigfn)
	if err := a.computeHMAC(ctx, a.tokens.Current(), hmacfn, sigfn); err != nil {
		return err
	}
	if err := a.rcs.Add(ctx, hmacfn); err != nil {
		return err
	}
	commitMsg := "saved gpgid ACL"
	if cm := getCommitMsg(ctx); cm != "" {
		commitMsg = cm
	}
	return a.rcs.Commit(ctx, commitMsg)
}

func (a *ACL) init(ctx context.Context) error {
	// create new token file
	a.tokens = TokenList{NewToken()}
	return a.save(withCommitMsg(ctx, "initialized gpgid ACL"))
}

func (a *ACL) marshalTokenFile(ctx context.Context) error {
	buf, err := json.Marshal(a.tokens)
	if err != nil {
		return err
	}

	ciphertext, err := a.crypto.Encrypt(ctx, buf, a.Recipients())
	if err != nil {
		return store.ErrEncrypt
	}
	if err := ioutil.WriteFile(a.tokenfile(), ciphertext, 0600); err != nil {
		return err
	}

	return nil
}

func (a *ACL) unmarshalTokenFile(ctx context.Context) error {
	ciphertext, err := ioutil.ReadFile(a.tokenfile())
	if err != nil {
		return err
	}

	buf, err := a.crypto.Decrypt(ctx, ciphertext)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, &a.tokens)
}

func (a *ACL) verify(ctx context.Context) error {
	if err := a.unmarshalTokenFile(ctx); err != nil {
		return err
	}

	// store/.gpg-id.sig.0xDEADBEEF
	// store/.gpg-id.hmac.0xDEADBEEF
	sigs, err := filepath.Glob(a.idf + ".sig.*")
	if err != nil {
		return err
	}
	if len(sigs) < 1 {
		return fmt.Errorf("no signatures found")
	}
	signByLatest := false
	for _, sigf := range sigs {
		out.Debug(ctx, "verify - Checking %s ...", sigf)
		sigBuf, err := ioutil.ReadFile(sigf)
		if err != nil {
			return err
		}
		idfBuf, err := ioutil.ReadFile(a.idf)
		if err != nil {
			return err
		}
		// verify signature
		if _, err := a.crypto.Verify(ctx, idfBuf, sigBuf); err != nil {
			return err
		}
		out.Debug(ctx, "verify - GPG valid")
		// verify HMAC
		latest, err := a.verifyHMAC(ctx, a.hmacNameFromSig(sigf), sigf)
		if err != nil {
			return err
		}
		out.Debug(ctx, "verify - HMAC valid")
		signByLatest = signByLatest || latest
	}

	if !signByLatest {
		return fmt.Errorf("no signature by latest token. possibly replay attack")
	}
	return nil
}
