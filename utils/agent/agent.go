package agent

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/justwatchcom/gopass/utils/agent/client"
	"github.com/justwatchcom/gopass/utils/pinentry"
)

type piner interface {
	Close()
	Confirm() bool
	Set(string, string) error
	GetPin() ([]byte, error)
}

// Agent is a gopass agent
type Agent struct {
	socket   string
	testing  bool
	server   *http.Server
	cache    *cache
	pinentry func() (piner, error)
}

// New creates a new agent
func New(dir string) *Agent {
	a := &Agent{
		socket: filepath.Join(dir, ".gopass-agent.sock"),
		cache: &cache{
			ttl:    time.Hour,
			maxTTL: 24 * time.Hour,
		},
		pinentry: func() (piner, error) {
			return pinentry.New()
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", a.servePing)
	mux.HandleFunc("/passphrase", a.servePassphrase)
	mux.HandleFunc("/cache/remove", a.serveRemove)
	mux.HandleFunc("/cache/purge", a.servePurge)
	a.server = &http.Server{
		Handler: mux,
	}
	return a
}

// NewForTesting creates a new agent for testing
func NewForTesting(dir, key, pass string) *Agent {
	a := New(dir)
	a.cache.set(key, pass)
	a.testing = true
	return a
}

// ListenAndServe starts listening and blocks
func (a *Agent) ListenAndServe() error {
	lis, err := net.Listen("unix", a.socket)
	if err != nil {
		if err := client.New(filepath.Dir(a.socket)).Ping(); err == nil {
			return fmt.Errorf("agent already running")
		}
		if err := os.Remove(a.socket); err != nil {
			return err
		}
		lis, err = net.Listen("unix", a.socket)
		if err != nil {
			return err
		}
	}
	return a.server.Serve(lis)
}

func (a *Agent) servePing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func (a *Agent) serveRemove(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if !a.testing {
		a.cache.remove(key)
	}
	fmt.Fprintf(w, "OK")
}

func (a *Agent) servePurge(w http.ResponseWriter, r *http.Request) {
	if !a.testing {
		a.cache.purge()
	}
	fmt.Fprintf(w, "OK")
}

func (a *Agent) servePassphrase(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	reason := r.URL.Query().Get("reason")

	if pass, found := a.cache.get(key); found || a.testing {
		fmt.Fprintf(w, pass)
		return
	}

	pi, err := a.pinentry()
	if err != nil {
		http.Error(w, fmt.Sprintf("Pinentry Error: %s", err), http.StatusInternalServerError)
		return
	}
	defer pi.Close()
	_ = pi.Set("title", "gopass Agent")
	_ = pi.Set("desc", "Need your passphrase "+reason)
	_ = pi.Set("prompt", "Please enter your passphrase:")
	_ = pi.Set("ok", "OK")
	pw, err := pi.GetPin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Pinentry Error: %s", err), http.StatusInternalServerError)
		return
	}
	a.cache.set(key, string(pw))
	fmt.Fprintf(w, string(pw))
}
