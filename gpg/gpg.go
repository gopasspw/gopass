package gpg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	fileMode = 0600
	dirPerm  = 0700
)

func init() {
	// ensure created files don't have group or world perms set
	// this setting should be inherited by sub-processes
	syscall.Umask(077)
}

var (
	reUIDComment = regexp.MustCompile(`([^(<]+)\s+(\([^)]+\))\s+<([^>]+)>`)
	reUID        = regexp.MustCompile(`([^(<]+)\s+<([^>]+)>`)
	// GPGArgs contains the default GPG args for non-interactive use. Note: Do not use '--batch'
	// as this will disable (necessary) passphrase questions!
	GPGArgs = []string{"--quiet", "--yes", "--compress-algo=none", "--no-encrypt-to", "--no-auto-check-trustdb"}
	// Debug prints all the commands executed
	Debug = false
)

// KeyList is a searchable slice of Keys
type KeyList []Key

// UseableKeys returns the list of useable (valid keys)
func (kl KeyList) UseableKeys() KeyList {
	nkl := make(KeyList, 0, len(kl))
	for _, k := range kl {
		if !k.IsUseable() {
			continue
		}
		nkl = append(nkl, k)
	}
	return nkl
}

// FindKey will try to find the requested key
func (kl KeyList) FindKey(id string) (Key, error) {
	id = strings.TrimPrefix(id, "0x")
	for _, k := range kl {
		if k.Fingerprint == id {
			return k, nil
		}
		if strings.HasSuffix(k.Fingerprint, id) {
			return k, nil
		}
		for _, ident := range k.Identities {
			if ident.String() == id {
				return k, nil
			}
			if ident.Email == id {
				return k, nil
			}
		}
		for sk := range k.SubKeys {
			if strings.HasSuffix(sk, id) {
				return k, nil
			}
		}
	}
	return Key{}, fmt.Errorf("No matching key found")
}

// ParseColons parses the `--with-colons` output format of GPG
func ParseColons(reader io.Reader) KeyList {
	kl := make(KeyList, 0, 100)

	scanner := bufio.NewScanner(reader)

	// http://git.gnupg.org/cgi-bin/gitweb.cgi?p=gnupg.git;a=blob_plain;f=doc/DETAILS
	// Fields:
	// 0 - Type of record
	//     Types:
	//     pub - Public Key
	//     crt - X.509 cert
	//     crs - X.509 cert and private key
	//     sub - Subkey (Secondary Key)
	//     sec - Secret / Private Key
	//     ssb - Secret Subkey
	//     uid - User ID
	//     uat - User attribute
	//     sig - Signature
	//     rev - Revocation Signature
	//     fpr - Fingerprint (field 9)
	//     pkd - Public Key Data
	//     grp - Keygrip
	//     rvk - Revocation KEy
	//     tfs - TOFU stats
	//     tru - Trust database info
	//     spk - Signature subpacket
	//     cfg - Configuration data
	// 1 - Validity
	// 2 - Key length
	// 3 - Public Key Algo
	// 4 - KeyID
	// 5 - Creation Date (UTC)
	// 6 - Expiration Date
	// 7 - Cert S/N
	// 8 - Ownertrust
	// 9 - User-ID
	// 10 - Sign. Class
	// 11 - Key Caps.
	// 12 - Issuer cert fp
	// 13 - Flag
	// 14 - S/N of a token
	// 15 - Hash algo (2 - SHA-1, 8 - SHA-256)
	// 16 - Curve Name

	var cur Key

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Split(line, ":")

		switch fields[0] {
		case "pub":
			fallthrough
		case "sec":
			if cur.Fingerprint != "" && cur.KeyLength > 0 {
				kl = append(kl, cur)
			}
			validity := fields[1]
			if validity == "" && fields[0] == "sec" {
				validity = "u"
			}
			cur = Key{
				KeyType:        fields[0],
				Validity:       validity,
				KeyLength:      parseInt(fields[2]),
				CreationDate:   parseTS(fields[5]),
				ExpirationDate: parseTS(fields[6]),
				Ownertrust:     fields[8],
				Identities:     make(map[string]Identity, 1),
				SubKeys:        make(map[string]struct{}, 1),
			}
		case "sub":
			fallthrough
		case "ssb":
			cur.SubKeys[fields[4]] = struct{}{}
		case "fpr":
			if cur.Fingerprint == "" {
				cur.Fingerprint = fields[9]
			}
		case "uid":
			sn := fields[7]
			id := fields[9]
			ni := Identity{}
			if reUIDComment.MatchString(id) {
				if m := reUIDComment.FindStringSubmatch(id); len(m) > 3 {
					ni.Name = m[1]
					ni.Comment = strings.Trim(m[2], "()")
					ni.Email = m[3]
				}
			} else if reUID.MatchString(id) {
				if m := reUID.FindStringSubmatch(id); len(m) > 2 {
					ni.Name = m[1]
					ni.Email = m[2]
				}
			}
			cur.Identities[sn] = ni
		}
	}

	if cur.Fingerprint != "" && cur.KeyLength > 0 {
		kl = append(kl, cur)
	}

	return kl
}

// parseTS parses the passed string as an Epoch int and returns
// the time struct or the zero time struct
func parseTS(str string) time.Time {
	t := time.Time{}

	if sec, err := strconv.ParseInt(str, 10, 64); err == nil {
		t = time.Unix(sec, 0)
	}

	return t
}

// parseInt parses the passed string as an int and returns it
// or 0 on errors
func parseInt(str string) int {
	i := 0

	if iv, err := strconv.ParseInt(str, 10, 32); err == nil {
		i = int(iv)
	}

	return i
}

// Key is a GPG key (public or secret)
type Key struct {
	KeyType        string
	KeyLength      int
	Validity       string
	CreationDate   time.Time
	ExpirationDate time.Time
	Ownertrust     string
	Fingerprint    string
	Identities     map[string]Identity
	SubKeys        map[string]struct{}
}

// IsUseable returns true if GPG would assume this key is useable for encryption
func (k Key) IsUseable() bool {
	if k.ExpirationDate.Before(time.Now()) {
		return false
	}
	switch k.Validity {
	case "m":
		return true
	case "f":
		return true
	case "u":
		return true
	}
	return false
}

// String implement fmt.Stringer. This method produces output that is close to, but
// not exactly the same, as the output form GPG itself
func (k Key) String() string {
	fp := ""
	if len(k.Fingerprint) > 24 {
		fp = k.Fingerprint[24:]
	}
	out := fmt.Sprintf("%s   %dD/0x%s %s", k.KeyType, k.KeyLength, fp, k.CreationDate.Format("2006-01-02"))
	if !k.ExpirationDate.IsZero() {
		out += fmt.Sprintf(" [expires: %s]", k.ExpirationDate.Format("2006-01-02"))
	}
	out += "\n      Key fingerprint = " + k.Fingerprint
	for _, id := range k.Identities {
		out += fmt.Sprintf("\n" + id.String())
	}
	return out
}

// OneLine prints a terse representation of this key on one line (includes only
// the first identity!)
func (k Key) OneLine() string {
	id := Identity{}
	for _, i := range k.Identities {
		id = i
		break
	}
	return fmt.Sprintf("0x%s - %s", k.Fingerprint[24:], id.ID())
}

// Identity is a GPG identity, one key can have many IDs
type Identity struct {
	Name    string
	Comment string
	Email   string
}

// ID returns the GPG ID format
func (i Identity) ID() string {
	out := i.Name
	if i.Comment != "" {
		out += " (" + i.Comment + ")"
	}
	out += " <" + i.Email + ">"
	return out
}

// String implement fmt.Stringer. This method resembels the output gpg uses
// for user-ids
func (i Identity) String() string {
	return "uid                            " + i.ID()
}

// listKey lists all keys of the given type and matching the search strings
func listKeys(typ string, search ...string) (KeyList, error) {
	args := []string{"--with-colons", "--with-fingerprint", "--fixed-list-mode", "--list-" + typ + "-keys"}
	args = append(args, search...)
	cmd := exec.Command("gpg", args...)
	if Debug {
		fmt.Printf("gpg.listKeys: %s %+v\n", cmd.Path, cmd.Args)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		if bytes.Contains(out, []byte("secret key not available")) {
			return KeyList{}, nil
		}
		return KeyList{}, err
	}

	return ParseColons(bytes.NewBuffer(out)), nil
}

// ListPublicKeys returns a parsed list of GPG public keys
func ListPublicKeys(search ...string) (KeyList, error) {
	return listKeys("public", search...)
}

// ListPrivateKeys returns a parsed list of GPG secret keys
func ListPrivateKeys(search ...string) (KeyList, error) {
	return listKeys("secret", search...)
}

// GetRecipients returns a list of recipient IDs for a given file
func GetRecipients(file string) ([]string, error) {
	_ = os.Setenv("LANGUAGE", "C")
	recp := make([]string, 0, 5)

	args := []string{"--batch", "--list-only", "--no-default-keyring", "--secret-keyring", "/dev/null", file}
	cmd := exec.Command("gpg", args...)
	if Debug {
		fmt.Printf("gpg.GetRecipients: %s %+v\n", cmd.Path, cmd.Args)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, err
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "gpg:") {
			continue
		}
		p := strings.Split(line, ",")
		if len(p) < 2 {
			continue
		}
		p = strings.Split(strings.TrimSpace(p[1]), " ")
		if len(p) < 2 {
			continue
		}
		recp = append(recp, p[1])
	}

	return recp, nil
}

// Encrypt will encrypt the given content for the recipients. If alwaysTrust is true
// the trust-model will be set to always as to avoid (annoying) "unuseable public key"
// errors when encrypting.
func Encrypt(path string, content []byte, recipients []string, alwaysTrust bool) error {
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return err
	}

	args := append(GPGArgs, "--encrypt", "--output", path)
	if alwaysTrust {
		// changing the trustmodel is possibly dangerous. A user should always
		// explicitly opt-in to do this
		args = append(args, "--trust-model=always")
	}
	for _, r := range recipients {
		args = append(args, "--recipient", r)
	}

	cmd := exec.Command("gpg", args...)
	if Debug {
		fmt.Printf("gpg.Encrypt: %s %+v\n", cmd.Path, cmd.Args)
	}
	cmd.Stdin = bytes.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// Decrypt will try to decrypt the given file
func Decrypt(path string) ([]byte, error) {
	args := append(GPGArgs, "--decrypt", path)
	cmd := exec.Command("gpg", args...)
	if Debug {
		fmt.Printf("gpg.Decrypt: %s %+v\n", cmd.Path, cmd.Args)
	}
	return cmd.Output()
}

// ExportPublicKey will export the named public key to the location given
func ExportPublicKey(id, filename string) error {
	args := append(GPGArgs, "--armor", "--export", id)
	cmd := exec.Command("gpg", args...)
	if Debug {
		fmt.Printf("gpg.ExportPublicKey: %s %+v\n", cmd.Path, cmd.Args)
	}
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, out, fileMode)
}

// ImportPublicKey will import a key from the given location
func ImportPublicKey(filename string) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	args := append(GPGArgs, "--import")
	cmd := exec.Command("gpg", args...)
	if Debug {
		fmt.Printf("gpg.ImportPublicKey: %s %+v\n", cmd.Path, cmd.Args)
	}
	cmd.Stdin = bytes.NewReader(buf)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
