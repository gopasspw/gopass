// Package identityfile parses bounded age identity files.
package identityfile

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"filippo.io/age"
)

const privateKeySizeLimit = 1 << 24 // 16 MiB

// Parser converts one non-empty, non-comment line into an age identity.
type Parser func(string) (age.Identity, error)

// Parse reads identities and delegates identity-specific decoding to parser.
func Parse(reader io.Reader, parser Parser) ([]age.Identity, error) {
	var identities []age.Identity
	scanner := bufio.NewScanner(io.LimitReader(reader, privateKeySizeLimit))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		identity, err := parser(line)
		if err != nil {
			return nil, fmt.Errorf("error at line %d: %w", lineNumber, err)
		}
		identities = append(identities, identity)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read secret keys file: %w", err)
	}
	if len(identities) == 0 {
		return nil, fmt.Errorf("no secret keys found")
	}

	return identities, nil
}
