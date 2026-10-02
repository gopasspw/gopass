package agent

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gopasspw/gopass/internal/buildinfo"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Client is a client for the age agent.
type Client struct {
	socketPath string
}

// NewClient creates a new client.
func NewClient() *Client {
	return &Client{
		socketPath: filepath.Join(appdir.UserRuntime(), socketName),
	}
}

// SocketPath returns the path of the agent socket this client dials. The
// path depends on the platform's runtime-dir convention (GOPASS_HOMEDIR,
// XDG_RUNTIME_DIR or LOCALAPPDATA), so tests that simulate an agent at the
// exact location the client expects should listen here rather than
// reconstructing the path.
func (c *Client) SocketPath() string {
	return c.socketPath
}

func (c *Client) connect() (net.Conn, error) {
	if err := c.checkSocketSecurity(); err != nil {
		return nil, err
	}

	debug.Log("connecting to agent at %s", c.socketPath)
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent: %w", err)
	}

	debug.Log("connected to agent at %s", c.socketPath)

	return conn, nil
}

// exchange writes one command and reads exactly one response line over an
// established connection. The response read is bounded by maxLineSize (the
// agent's own line limit), so a misbehaving agent cannot grow client
// memory without bound. ERR responses are converted to errors. The agent
// is strictly request-response, so using a fresh scanner per exchange
// cannot swallow a later response.
func (c *Client) exchange(conn net.Conn, cmd string) (string, error) {
	if _, err := fmt.Fprintln(conn, cmd); err != nil {
		return "", fmt.Errorf("failed to send command to agent: %w", err)
	}

	sc := bufio.NewScanner(conn)
	sc.Buffer(nil, maxLineSize)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return "", fmt.Errorf("failed to read response from agent: %w", err)
		}

		return "", fmt.Errorf("failed to read response from agent: unexpected EOF")
	}

	resp := strings.TrimSpace(sc.Text())
	if strings.HasPrefix(resp, "ERR") {
		return "", fmt.Errorf("agent error: %s", strings.TrimPrefix(resp, "ERR "))
	}

	return strings.TrimPrefix(resp, "OK "), nil
}

func (c *Client) send(cmd string) (string, error) {
	conn, err := c.connect()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.Close()
	}()

	return c.exchange(conn, cmd)
}

// Capabilities is the parsed payload of the agent's hello response.
// Tokens are either bare (presence-only, e.g. "decrypt") or key=value
// (e.g. "maxline=16777216"). Unknown tokens are ignored by design so newer
// agents can advertise new capabilities without breaking older clients.
type Capabilities struct {
	raw    string
	tokens map[string]string
}

// Has reports whether the agent advertised the given capability token.
func (c *Capabilities) Has(token string) bool {
	if c == nil {
		return false
	}
	_, ok := c.tokens[token]

	return ok
}

// Value returns the value of a key=value token. For bare tokens it returns
// ("", true).
func (c *Capabilities) Value(key string) (string, bool) {
	if c == nil {
		return "", false
	}
	v, ok := c.tokens[key]

	return v, ok
}

// Raw returns the raw response payload, preserving the agent's token order
// (e.g. for display).
func (c *Capabilities) Raw() string {
	if c == nil {
		return ""
	}

	return c.raw
}

// parseCapabilities parses a hello response payload. Unknown tokens are
// ignored and malformed tokens (e.g. an empty key) are skipped. When a
// token appears more than once the first occurrence wins: redefinitions
// (including a bare token repeating a key=value one) are ignored, so a
// contradictory payload cannot silently downgrade a capability.
func parseCapabilities(resp string) *Capabilities {
	caps := &Capabilities{
		raw:    resp,
		tokens: make(map[string]string),
	}
	for _, tok := range strings.Fields(resp) {
		k, v, _ := strings.Cut(tok, "=")
		if k == "" {
			continue
		}
		if _, exists := caps.tokens[k]; exists {
			continue
		}
		caps.tokens[k] = v
	}

	return caps
}

// Capabilities performs the hello capability handshake. A nil error means
// the agent speaks hello; ANY error - a connection failure or an ERR
// response (a pre-hello agent answers "ERR unknown command") - means
// negotiation is unavailable and callers must keep the pre-hello
// behaviour. Callers must judge by the error's existence, never by its
// text.
func (c *Client) Capabilities() (*Capabilities, error) {
	cmd := "hello"
	if v, ok := buildinfo.ModuleVersion(); ok {
		cmd += " gopass/" + v
	}
	resp, err := c.send(cmd)
	if err != nil {
		return nil, err
	}

	return parseCapabilities(resp), nil
}

// AgentInfo is the combined status and capability view of a running agent,
// gathered over a single connection.
type AgentInfo struct {
	// Status is the payload of the status command: "" or "locked".
	Status string

	// Capabilities is nil when the agent did not answer hello (a pre-hello
	// legacy agent, or a failure mid-handshake).
	Capabilities *Capabilities
}

// Info queries the agent status and its capabilities over a single
// connection (the client is otherwise one connection per command). A
// transport failure on the status exchange fails the call; a failure or
// ERR on hello only degrades Capabilities to nil, matching the
// negotiation contract: judge by the error's existence, never its text,
// and keep the pre-hello behaviour.
func (c *Client) Info() (*AgentInfo, error) {
	conn, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	status, err := c.exchange(conn, "status")
	if err != nil {
		return nil, err
	}

	helloCmd := "hello"
	if v, ok := buildinfo.ModuleVersion(); ok {
		helloCmd += " gopass/" + v
	}
	resp, err := c.exchange(conn, helloCmd)
	if err != nil {
		return &AgentInfo{Status: status}, nil
	}

	return &AgentInfo{Status: status, Capabilities: parseCapabilities(resp)}, nil
}

// Ping pings the agent.
func (c *Client) Ping() error {
	_, err := c.send("ping")

	return err
}

// Status returns the agent's status.
func (c *Client) Status() (string, error) {
	return c.send("status")
}

// SendIdentities sends the identities to the agent.
func (c *Client) SendIdentities(ids string) error {
	_, err := c.send("identities " + ids)

	return err
}

// Decrypt decrypts the given ciphertext.
func (c *Client) Decrypt(ciphertext []byte) ([]byte, error) {
	resp, err := c.send("decrypt " + base64.StdEncoding.EncodeToString(ciphertext))
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(resp)
}

// Lock locks the agent.
func (c *Client) Lock() error {
	_, err := c.send("lock")

	return err
}

// Unlock unlocks the agent.
func (c *Client) Unlock() error {
	_, err := c.send("unlock")

	return err
}

// SetTimeout sets the agent's timeout.
func (c *Client) SetTimeout(timeout int) error {
	_, err := c.send("set-timeout " + strconv.Itoa(timeout))

	return err
}

// Quit quits the agent.
func (c *Client) Quit() error {
	_, err := c.send("quit")

	return err
}
