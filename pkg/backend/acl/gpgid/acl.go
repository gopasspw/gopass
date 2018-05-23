package gpgid

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/justwatchcom/gopass/pkg/backend"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/store"
)

// ACL is an gpg-id based store ACL
type ACL struct {
	crypto backend.Crypto
	rcs    backend.RCS
	fs     backend.Storage
	signed bool
	valid  bool
	recps  map[string]struct{}
	tokens TokenList
}

// Init initializes (new) store integrity / ACL structs
func Init(ctx context.Context, crypto backend.Crypto, rcs backend.RCS, fs backend.Storage) (*ACL, error) {
	a := &ACL{
		crypto: crypto,
		rcs:    rcs,
		fs:     fs,
		signed: false,
		valid:  false,
		recps:  make(map[string]struct{}),
		tokens: make(TokenList, 0),
	}
	if err := a.unmarshal(ctx); err != nil {
		return nil, err
	}
	if a.isSigned(ctx) {
		return nil, fmt.Errorf("already signed")
	}
	if err := a.init(ctx); err != nil {
		return a, err
	}
	return a, nil
}

// Load tries to load and verify existing ACL information
func Load(ctx context.Context, crypto backend.Crypto, rcs backend.RCS, fs backend.Storage) (*ACL, error) {
	a := &ACL{
		crypto: crypto,
		rcs:    rcs,
		fs:     fs,
		signed: false,
		valid:  false,
		recps:  make(map[string]struct{}),
		tokens: make(TokenList, 0),
	}
	if err := a.unmarshal(ctx); err != nil {
		return nil, err
	}
	if !a.isSigned(ctx) {
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
	return a.crypto.IDFile() + ".token"
}

func (a *ACL) isSigned(ctx context.Context) bool {
	return a.fs.Exists(ctx, a.tokenfile())
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
	if err := a.marshal(ctx); err != nil {
		return err
	}
	if err := a.rcs.Add(ctx, a.crypto.IDFile()); err != nil {
		return err
	}
	// sign gpg id
	idf, err := a.fs.Get(ctx, a.crypto.IDFile())
	if err != nil {
		return err
	}
	signed, err := a.crypto.Sign(ctx, idf)
	if err != nil {
		return err
	}
	sigfn := a.crypto.IDFile() + ".sig." + sigID
	if err := a.fs.Set(ctx, sigfn, signed); err != nil {
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
	if err := a.fs.Set(ctx, a.tokenfile(), ciphertext); err != nil {
		return err
	}

	return nil
}

func (a *ACL) unmarshalTokenFile(ctx context.Context) error {
	ciphertext, err := a.fs.Get(ctx, a.tokenfile())
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
	sigs, err := a.prefixFilter(ctx, a.crypto.IDFile()+".sig.")
	if err != nil {
		return err
	}
	if len(sigs) < 1 {
		return fmt.Errorf("no signatures found")
	}
	signByLatest := false
	for _, sigf := range sigs {
		out.Debug(ctx, "verify - Checking %s ...", sigf)
		sigBuf, err := a.fs.Get(ctx, sigf)
		if err != nil {
			return err
		}
		idfBuf, err := a.fs.Get(ctx, a.crypto.IDFile())
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

func (a *ACL) prefixFilter(ctx context.Context, prefix string) ([]string, error) {
	files, err := a.fs.List(ctx, prefix)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, len(files))
	for _, file := range files {
		if !strings.HasPrefix(file, prefix) {
			continue
		}
		res = append(res, file)
	}
	return res, nil
}
