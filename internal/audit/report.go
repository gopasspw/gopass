package audit

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gopasspw/gopass/internal/hashsum"
	"github.com/gopasspw/gopass/internal/set"
)

type SecretReport struct {
	Name     string
	Errors   []error
	Warnings []string
	Age      time.Duration
}

func (s SecretReport) Record() []string {
	return []string{
		s.Name,
		s.Age.String(),
		strings.Join(errors(s.Errors), ";"),
		strings.Join(s.Warnings, ";"),
	}
}

func errors(e []error) []string {
	s := make([]string, 0, len(e))
	for _, es := range e {
		s = append(s, es.Error())
	}

	return s
}

type Report struct {
	Secrets map[string]SecretReport
}

type ReportBuilder struct {
	sync.Mutex

	secrets map[string]SecretReport
	// SHA512(password) -> secret names
	duplicates map[string]set.Set[string]

	// HIBP
	// SHA1(password) -> secret names
	sha1sums map[string]set.Set[string]
}

func (r *ReportBuilder) AddPassword(name, pw string) {
	if name == "" || pw == "" {
		return
	}

	r.Lock()
	defer r.Unlock()

	s256 := hashsum.SHA256Hex(pw)
	d := r.duplicates[s256]
	d.Add(name)
	r.duplicates[s256] = d

	s1 := hashsum.SHA1Hex(pw)
	s := r.sha1sums[s1]
	s.Add(name)
	r.sha1sums[s1] = s
}

func (r *ReportBuilder) AddError(name string, e error) {
	if name == "" || e == nil {
		return
	}

	r.Lock()
	defer r.Unlock()

	s := r.secrets[name]
	s.Name = name
	s.Errors = append(s.Errors, e)
	r.secrets[name] = s
}

func (r *ReportBuilder) SetAge(name string, age time.Duration) {
	if name == "" {
		return
	}

	r.Lock()
	defer r.Unlock()

	s := r.secrets[name]
	s.Name = name
	s.Age = age
	r.secrets[name] = s
}

func (r *ReportBuilder) AddWarning(name, msg string) {
	if name == "" || msg == "" {
		return
	}

	r.Lock()
	defer r.Unlock()

	s := r.secrets[name]
	s.Name = name
	if s.Warnings == nil {
		s.Warnings = make([]string, 0, 1)
	}
	s.Warnings = append(s.Warnings, msg)
	r.secrets[name] = s
}

func newReport() *ReportBuilder {
	return &ReportBuilder{
		secrets:    make(map[string]SecretReport, 512),
		duplicates: make(map[string]set.Set[string], 512),
		sha1sums:   make(map[string]set.Set[string], 512),
	}
}

// Finalize computes the duplicates.
func (r *ReportBuilder) Finalize() *Report {
	for k, s := range r.secrets {
		for _, secs := range r.duplicates {
			if secs.Contains(k) {
				s.Warnings = append(s.Warnings, fmt.Sprintf("Duplicates detected. Shared with: %+v", secs.Difference(set.New(k))))
			}
		}
	}

	ret := &Report{
		Secrets: make(map[string]SecretReport, len(r.secrets)),
	}

	for k := range r.secrets {
		ret.Secrets[k] = r.secrets[k]
	}

	return ret
}
