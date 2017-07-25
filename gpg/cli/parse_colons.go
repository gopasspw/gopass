package gpg

import (
	"bufio"
	"io"
	"strings"

	"github.com/justwatchcom/gopass/gpg"
)

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

// parseColons parses the `--with-colons` output format of GPG
func (g *GPG) parseColons(reader io.Reader) gpg.KeyList {
	kl := make(gpg.KeyList, 0, 100)

	scanner := bufio.NewScanner(reader)

	var cur gpg.Key

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
			cur = gpg.Key{
				KeyType:        fields[0],
				Validity:       validity,
				KeyLength:      parseInt(fields[2]),
				CreationDate:   parseTS(fields[5]),
				ExpirationDate: parseTS(fields[6]),
				Ownertrust:     fields[8],
				Identities:     make(map[string]gpg.Identity, 1),
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
			ni := gpg.Identity{}
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
