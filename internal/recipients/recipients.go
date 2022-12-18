package recipients

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/set"
	"golang.org/x/exp/maps"
)

// Recipients is a list of Key IDs. It will try to retain the file as much as possible while manipulating the recipients.
type Recipients struct {
	r   map[string]bool
	raw strings.Builder
}

// New creates a new list of Key IDs.
func New() *Recipients {
	return &Recipients{
		r:   make(map[string]bool, 4),
		raw: strings.Builder{},
	}
}

// IDs returns the key IDs.
func (r *Recipients) IDs() []string {
	res := maps.Keys(r.r)
	sort.Strings(res)

	return res
}

// Add adds a new recipients. It returns true if the recipient was added.
func (r *Recipients) Add(key string) bool {
	key = strings.TrimSpace(key)
	if _, found := r.r[key]; found {
		return false
	}

	r.r[key] = true

	return true
}

// Remove deletes an existing recipient. It returns true if the recipients
// was present and got removed.
func (r *Recipients) Remove(key string) bool {
	key = strings.TrimSpace(key)
	if _, found := r.r[key]; !found {
		return false
	}

	delete(r.r, key)

	return true
}

// Has returns true if the recipient is found.
func (r *Recipients) Has(key string) bool {
	key = strings.TrimSpace(key)
	_, found := r.r[key]

	return found
}

// Marshal all in memory Recipients line by line to []byte.
func (r *Recipients) Marshal() []byte {
	if len(r.r) == 0 {
		return []byte("\n")
	}

	seen := make(map[string]bool, len(r.r))

	out := bytes.Buffer{}
	s := bufio.NewScanner(strings.NewReader(r.raw.String()))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())

		// pass through comments
		if strings.HasPrefix(line, "#") || line == "" {
			out.WriteString(line)
			out.WriteString("\n")

			continue
		}

		key := line
		// trim any trailing comments
		if idx := strings.Index(line, "#"); idx != -1 {
			key = strings.TrimSpace(line[:idx])
		}

		// skip deleted IDs
		if _, found := r.r[key]; !found {
			continue
		}

		out.WriteString(line)
		out.WriteString("\n")

		seen[key] = true
	}

	// add new keys
	for _, k := range set.SortedKeys(r.r) {
		// added before
		if _, found := seen[k]; found {
			continue
		}

		out.WriteString(k)
		out.WriteString("\n")

		seen[k] = true
	}

	return out.Bytes()
}

// Hash returns the hex encoded SHA256 sum of the recipients.
func (r *Recipients) Hash() string {
	h := sha256.New()
	_, _ = h.Write(r.Marshal())

	return fmt.Sprintf("%x", h.Sum(nil))
}

// Unmarshal Recipients line by line from a io.Reader. Handles Unix, Windows and Mac line endings.
func Unmarshal(buf []byte) *Recipients {
	in := strings.ReplaceAll(string(buf), "\r", "\n")

	r := New()
	s := bufio.NewScanner(strings.NewReader(in))
	for s.Scan() {
		line := s.Text()

		r.raw.WriteString(line)
		r.raw.WriteString("\n")

		line = strings.TrimSpace(line)

		// skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// trim trailing comments
		key := line
		if idx := strings.Index(line, "#"); idx != -1 {
			key = strings.TrimSpace(line[:idx])
		}

		if len(key) < 1 {
			continue
		}

		r.r[key] = true
	}

	return r
}
