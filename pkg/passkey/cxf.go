package passkey

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"
)

// This file implements the import and export of credentials in the FIDO
// Credential Exchange Format (CXF) 1.0.
//
// See https://fidoalliance.org/specs/cx/cxf-v1.0-ps-errata-20260309.html

const (
	cxfVersionMajor = 1
	cxfVersionMinor = 0
	// cxfExporterRPID identifies gopass as the exporting provider.
	cxfExporterRPID = "gopass.pw"
	// cxfExporterName is the name of gopass shown by importing providers.
	cxfExporterName = "gopass"
	// cxfTypePasskey is the type of the passkey credential.
	cxfTypePasskey = "passkey"
)

// cxfHeader is the top-level structure of an exchange.
type cxfHeader struct {
	Version             cxfVersion   `json:"version"`
	ExporterRPID        string       `json:"exporterRpId"`
	ExporterDisplayName string       `json:"exporterDisplayName"`
	Timestamp           int64        `json:"timestamp"`
	Accounts            []cxfAccount `json:"accounts"`
}

type cxfVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
}

// cxfAccount is the account of the credential owner at the provider.
type cxfAccount struct {
	ID          Base64URL         `json:"id"`
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	Collections []json.RawMessage `json:"collections"`
	Items       []cxfItem         `json:"items"`
}

// cxfItem groups related credentials.
type cxfItem struct {
	ID          Base64URL         `json:"id"`
	CreationAt  int64             `json:"creationAt,omitempty"`
	Title       string            `json:"title"`
	Subtitle    string            `json:"subtitle,omitempty"`
	Credentials []json.RawMessage `json:"credentials"`
}

// cxfPasskey is the passkey credential.
type cxfPasskey struct {
	Type            string          `json:"type"`
	CredentialID    Base64URL       `json:"credentialId"`
	RPID            string          `json:"rpId"`
	Username        string          `json:"username"`
	UserDisplayName string          `json:"userDisplayName"`
	UserHandle      Base64URL       `json:"userHandle"`
	Key             Base64URL       `json:"key"`
	Extensions      json.RawMessage `json:"fido2Extensions,omitempty"`
}

// ExportCXF writes the credentials of the entries to w as an unencrypted
// Credential Exchange Format document. The document contains the private
// keys and must be protected accordingly.
//
// The format has no notion of a signature counter. Credentials that maintain
// one must not be exported. They are left out and returned.
func ExportCXF(w io.Writer, entries []Entry) ([]Entry, error) {
	accountID := make([]byte, 16)
	if _, err := rand.Read(accountID); err != nil {
		return nil, fmt.Errorf("failed to generate account ID: %w", err)
	}

	items := make([]cxfItem, 0, len(entries))
	skipped := make([]Entry, 0)

	seen := make([][]byte, 0, len(entries))

	for _, entry := range entries {
		cred := entry.Credential
		if err := cred.validate(); err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Name, err)
		}

		// A credential that is stored more than once is exported once: the
		// IDs of the items of an account have to be unique.
		if slices.ContainsFunc(seen, func(id []byte) bool { return bytes.Equal(id, cred.ID) }) {
			continue
		}
		seen = append(seen, cred.ID)

		if cred.Counting() {
			skipped = append(skipped, entry)

			continue
		}

		key, err := marshalPrivateKey(cred.Key)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Name, err)
		}

		pk, err := json.Marshal(cxfPasskey{
			Type:            cxfTypePasskey,
			CredentialID:    cred.ID,
			RPID:            cred.RPID,
			Username:        cred.UserName,
			UserDisplayName: cred.UserDisplayName,
			UserHandle:      cred.UserHandle,
			Key:             key,
			Extensions:      cred.Extensions,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: failed to encode passkey: %w", entry.Name, err)
		}

		// Item IDs are opaque and limited to 64 bytes. Credential IDs can be
		// longer, so the ID is derived from it.
		id := sha256.Sum256(cred.ID)

		item := cxfItem{
			ID:          id[:],
			Title:       cred.RPID,
			Subtitle:    cred.Label(),
			Credentials: []json.RawMessage{pk},
		}
		if !cred.Created.IsZero() {
			item.CreationAt = cred.Created.Unix()
		}

		items = append(items, item)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	err := enc.Encode(cxfHeader{
		Version:             cxfVersion{Major: cxfVersionMajor, Minor: cxfVersionMinor},
		ExporterRPID:        cxfExporterRPID,
		ExporterDisplayName: cxfExporterName,
		Timestamp:           time.Now().Unix(),
		Accounts: []cxfAccount{{
			ID:          accountID,
			Collections: []json.RawMessage{},
			Items:       items,
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to write export: %w", err)
	}

	return skipped, nil
}

// ImportCXF reads the passkeys of an unencrypted Credential Exchange Format
// document. Credentials of any other type are ignored. Passkeys that can not
// be used, e.g. because their algorithm is not supported, are skipped and
// reported in the returned error together with the credentials that could
// be read.
//
// Imported credentials never maintain a signature counter.
func ImportCXF(r io.Reader) ([]*Credential, error) {
	dec := json.NewDecoder(r)

	var header cxfHeader
	if err := dec.Decode(&header); err != nil {
		return nil, fmt.Errorf("failed to parse Credential Exchange Format document: %w", err)
	}

	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("failed to parse Credential Exchange Format document: unexpected data after the document")
	}

	if header.Version.Major != cxfVersionMajor {
		return nil, fmt.Errorf("unsupported Credential Exchange Format version %d.%d", header.Version.Major, header.Version.Minor)
	}

	var errs []error

	creds := make([]*Credential, 0)

	for _, account := range header.Accounts {
		for _, item := range account.Items {
			for _, raw := range item.Credentials {
				// Credentials of other types may not even be objects.
				var typ struct {
					Type string `json:"type"`
				}
				if err := json.Unmarshal(raw, &typ); err != nil || typ.Type != cxfTypePasskey {
					continue
				}

				cred, err := importPasskey(raw, item)
				if err != nil {
					errs = append(errs, fmt.Errorf("skipping passkey of item %q: %w", item.Title, err))

					continue
				}

				creds = append(creds, cred)
			}
		}
	}

	return creds, errors.Join(errs...)
}

// importPasskey converts a passkey credential of an exchange document.
func importPasskey(raw json.RawMessage, item cxfItem) (*Credential, error) {
	var pk cxfPasskey
	if err := json.Unmarshal(raw, &pk); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCredential, err)
	}

	key, err := parsePrivateKey(pk.Key)
	if err != nil {
		return nil, fmt.Errorf("%q at %q: %w", pk.Username, pk.RPID, err)
	}

	// The extension data is kept as it is, as long as it is a dictionary.
	switch ext := bytes.TrimSpace(pk.Extensions); {
	case len(ext) == 0, bytes.Equal(ext, []byte("null")):
		pk.Extensions = nil
	case ext[0] != '{':
		return nil, fmt.Errorf("%q at %q: %w: fido2Extensions is not an object", pk.Username, pk.RPID, ErrInvalidCredential)
	}

	cred := &Credential{
		ID:              pk.CredentialID,
		RPID:            pk.RPID,
		UserHandle:      pk.UserHandle,
		UserName:        pk.Username,
		UserDisplayName: pk.UserDisplayName,
		Key:             key,
		// Importers must use a zero counter and must never increment it.
		SignCount: 0,
		// Only credentials that can be synced can be exchanged at all.
		BackupEligible: true,
		BackupState:    true,
		Extensions:     pk.Extensions,
	}

	if item.CreationAt > 0 {
		cred.Created = time.Unix(item.CreationAt, 0).UTC()
	}

	// The RP ID is part of what the Relying Party verifies. A credential
	// with an RP ID in a form that is not supported is rejected rather than
	// rewritten into one that would never be accepted.
	if err := cred.validate(); err != nil {
		return nil, fmt.Errorf("%q at %q: %w", pk.Username, pk.RPID, err)
	}

	return cred, nil
}
