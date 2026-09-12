//go:build linux

package secretservice

import (
	"fmt"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/gopasspw/gopass/internal/secretservice/crypto"
)

// session is one client's transport-encryption session.
type session struct {
	path   dbus.ObjectPath
	crypto crypto.Session
	conn   *dbus.Conn
	closed bool
	onGone func()

	mu sync.RWMutex
}

// sessionManager tracks the active sessions.
type sessionManager struct {
	conn     *dbus.Conn
	sessions map[string]*session

	mu sync.Mutex
}

func newSessionManager(conn *dbus.Conn) *sessionManager {
	return &sessionManager{
		conn:     conn,
		sessions: make(map[string]*session),
	}
}

// open negotiates a new session and exports it on the bus.
func (m *sessionManager) open(algorithm string, input []byte) (*session, []byte, error) {
	cryptoSession, output, err := crypto.New(algorithm, input)
	if err != nil {
		return nil, nil, err
	}

	id, err := newID('s')
	if err != nil {
		return nil, nil, err
	}

	s := &session{
		path:   SessionDBusPath(id),
		crypto: cryptoSession,
		conn:   m.conn,
	}
	s.onGone = func() {
		m.mu.Lock()
		delete(m.sessions, id)
		m.mu.Unlock()
	}

	m.mu.Lock()
	m.sessions[id] = s
	m.mu.Unlock()

	if err := m.conn.Export(s, s.path, SessionIface); err != nil {
		m.remove(id)

		return nil, nil, fmt.Errorf("export session: %w", err)
	}

	if err := m.conn.Export(introspectable(sessionIntrospection), s.path, IntrospectableIface); err != nil {
		m.remove(id)

		return nil, nil, fmt.Errorf("export session introspection: %w", err)
	}

	return s, output, nil
}

// get returns a session by its object path.
func (m *sessionManager) get(p dbus.ObjectPath) (*session, error) {
	id, err := parseSessionPath(p)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	s, ok := m.sessions[id]
	m.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("session not found: %s", p)
	}

	return s, nil
}

func (m *sessionManager) remove(id string) {
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()
}

// closeAll closes every open session, e.g. at daemon shutdown.
func (m *sessionManager) closeAll() {
	m.mu.Lock()
	sessions := make([]*session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	m.sessions = make(map[string]*session)
	m.mu.Unlock()

	for _, s := range sessions {
		s.teardown(false)
	}
}

// Close implements org.freedesktop.Secret.Session.Close.
func (s *session) Close() *dbus.Error {
	s.teardown(true)

	return nil
}

func (s *session) teardown(notify bool) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()

		return
	}
	s.closed = true
	s.mu.Unlock()

	_ = s.conn.Export(nil, s.path, SessionIface)
	_ = s.conn.Export(nil, s.path, IntrospectableIface)
	_ = s.crypto.Close()
	if notify && s.onGone != nil {
		s.onGone()
	}
}

// encrypt encrypts a payload using the session's algorithm.
func (s *session) encrypt(plaintext []byte) ([]byte, []byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, nil, fmt.Errorf("session is closed")
	}

	return s.crypto.Encrypt(plaintext)
}

// decrypt decrypts a payload using the session's algorithm.
func (s *session) decrypt(params, ciphertext []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, fmt.Errorf("session is closed")
	}

	return s.crypto.Decrypt(params, ciphertext)
}
