//go:build linux

package secretservice

import (
	"fmt"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/gopasspw/gopass/internal/secretservice/crypto"
	"github.com/gopasspw/gopass/pkg/debug"
)

// session is one client's transport-encryption session.
type session struct {
	path   dbus.ObjectPath
	crypto crypto.Session
	conn   *dbus.Conn
	// owner is the unique bus name of the client that opened the session. It
	// is used to reject use of the session by other clients and to tear the
	// session down when the owner disconnects.
	owner  string
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
	m := &sessionManager{
		conn:     conn,
		sessions: make(map[string]*session),
	}
	m.watchNameOwnerChanges()

	return m
}

// watchNameOwnerChanges removes a client's sessions when its unique bus name
// disappears. Without this a crashed or malicious client could leave unbounded
// exported session objects and AES keys alive for the daemon's lifetime.
func (m *sessionManager) watchNameOwnerChanges() {
	if err := m.conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus"),
		dbus.WithMatchMember("NameOwnerChanged"),
		dbus.WithMatchObjectPath("/org/freedesktop/DBus"),
	); err != nil {
		debug.Log("secret-service: cannot watch NameOwnerChanged: %s", err)

		return
	}

	ch := make(chan *dbus.Signal, 16)
	m.conn.Signal(ch)

	go func() {
		for sig := range ch {
			if sig.Name != "org.freedesktop.DBus.NameOwnerChanged" || len(sig.Body) != 3 {
				continue
			}

			name, _ := sig.Body[0].(string)
			newOwner, _ := sig.Body[2].(string)
			// Only a name that lost its owner (newOwner == "") matters.
			if name == "" || newOwner != "" {
				continue
			}

			m.removeOwner(name)
		}
	}()
}

// open negotiates a new session and exports it on the bus. owner is the unique
// bus name of the client that requested the session.
func (m *sessionManager) open(algorithm string, input []byte, owner string) (*session, []byte, error) {
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
		owner:  owner,
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

// get returns a session by its object path. When owner is non-empty the
// session must belong to that client, so one client cannot use another's
// session.
func (m *sessionManager) get(p dbus.ObjectPath, owner string) (*session, error) {
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
	if s.owner != owner {
		return nil, fmt.Errorf("session %s belongs to another client", p)
	}

	return s, nil
}

func (m *sessionManager) remove(id string) {
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()
}

// removeOwner tears down every session opened by the given unique bus name.
func (m *sessionManager) removeOwner(owner string) {
	m.mu.Lock()
	var gone []*session

	for id, s := range m.sessions {
		if s.owner == owner {
			gone = append(gone, s)
			delete(m.sessions, id)
		}
	}
	m.mu.Unlock()

	for _, s := range gone {
		s.teardown(false)
	}
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
func (s *session) Close(sender dbus.Sender) *dbus.Error {
	if s.owner != string(sender) {
		return dbusError(errNoSession, fmt.Errorf("session belongs to another client"))
	}

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
