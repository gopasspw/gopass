package audit

import (
	"fmt"
	"sync"
	"time"

	"github.com/gopasspw/gopass/internal/hashsum"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/set"
)

type Finding struct {
	Severity string
	Message  string
}

type SecretReport struct {
	Name string
	// analyzer -> finding details
	Findings map[string]Finding
	Age      time.Duration
}

func (s *SecretReport) HasFindings() bool {
	for _, f := range s.Findings {
		if f.Severity != "none" {
			return true
		}
	}

	return false
}

func (s *SecretReport) HumanizeAge() string {
	if s.Age < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(s.Age.Hours()))
	}
	days := int(s.Age.Hours() / 24)
	if days < 30 {
		return fmt.Sprintf("%d days", days)
	}
	months := days / 30
	if months < 12 {
		return fmt.Sprintf("%d months", months)
	}
	years := months / 12

	return fmt.Sprintf("%d years", years)
}

type Report struct {
	// secret name -> report
	Secrets map[string]SecretReport

	// finding -> secrets
	Findings map[string]set.Set[string]

	Template string
	Duration time.Duration
}

type ReportBuilder struct {
	// protects all below
	sync.Mutex

	// secret name -> report
	secrets map[string]SecretReport
	// finding -> secrets
	findings map[string]set.Set[string]

	// SHA512(password) -> secret names
	duplicates map[string]set.Set[string]

	// HIBP
	// SHA1(password) -> secret names
	sha1sums map[string]set.Set[string]

	t0 time.Time
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

func (r *ReportBuilder) AddFinding(secret, finding, message, severity string) {
	if secret == "" || finding == "" || message == "" || severity == "" {
		return
	}

	r.Lock()
	defer r.Unlock()

	// record individual findings
	s := r.secrets[secret]
	s.Name = secret
	if s.Findings == nil {
		s.Findings = make(map[string]Finding, 4)
	}
	f := s.Findings[finding]
	f.Message = message
	f.Severity = severity
	s.Findings[finding] = f
	r.secrets[secret] = s

	debug.Log("Secret %q has finding %q: %s with severity %s", secret, finding, message, severity)
	if severity == "none" {
		return
	}

	// record secrets per finding, for the summary
	ss := r.findings[finding]
	ss.Add(secret)
	r.findings[finding] = ss
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

func newReport() *ReportBuilder {
	return &ReportBuilder{
		secrets:    make(map[string]SecretReport, 512),
		findings:   make(map[string]set.Set[string], 512),
		duplicates: make(map[string]set.Set[string], 512),
		sha1sums:   make(map[string]set.Set[string], 512),
		t0:         time.Now().UTC(),
	}
}

// Finalize computes the duplicates.
func (r *ReportBuilder) Finalize() *Report {
	for k, s := range r.secrets {
		for _, secs := range r.duplicates {
			if secs.Len() < 2 {
				continue
			}
			if !secs.Contains(k) {
				continue
			}
			if s.Findings == nil {
				s.Findings = make(map[string]Finding, 1)
			}
			s.Findings["duplicates"] = Finding{
				Severity: "warning",
				Message:  fmt.Sprintf("Duplicates detected. Shared with: %+v", secs.Difference(set.New(k))),
			}
		}
		r.secrets[k] = s
	}

	ret := &Report{
		Secrets:  make(map[string]SecretReport, len(r.secrets)),
		Findings: make(map[string]set.Set[string], len(r.findings)),
		Duration: time.Since(r.t0),
	}

	for k := range r.secrets {
		ret.Secrets[k] = r.secrets[k]
	}

	for k := range r.findings {
		ret.Findings[k] = r.findings[k]
	}

	debug.Log("Finalized report: %d secrets, %d findings", len(ret.Secrets), len(ret.Findings))

	return ret
}
