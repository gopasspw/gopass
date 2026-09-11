//go:build linux

package tests

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/gopasspw/gopass/internal/secretservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startPrivateBus starts a private D-Bus session bus and returns its address
// plus a cleanup function. The test is skipped if dbus-daemon is unavailable.
func startPrivateBus(t *testing.T) (string, func()) {
	t.Helper()

	if _, err := exec.LookPath("dbus-daemon"); err != nil {
		t.Skipf("dbus-daemon not available: %v", err)
	}

	cmd := exec.Command("dbus-daemon", "--session", "--nofork", "--print-address")
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	require.NoError(t, cmd.Start())

	stop := func() { _ = cmd.Process.Kill() }

	addrCh := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "unix:") {
				addrCh <- line

				return
			}
		}
		close(addrCh)
	}()

	select {
	case addr, ok := <-addrCh:
		if !ok || addr == "" {
			stop()

			t.Skipf("failed to read private bus address: %s", stderr.String())
		}

		return addr, stop
	case <-time.After(10 * time.Second):
		stop()

		t.Skipf("timed out waiting for private bus: %s", stderr.String())

		return "", stop
	}
}

// connectBus connects to an explicit D-Bus address.
func connectBus(t *testing.T, address string) *dbus.Conn {
	t.Helper()

	conn, err := dbus.Dial(address)
	require.NoError(t, err)
	require.NoError(t, conn.Auth(nil))
	require.NoError(t, conn.Hello())

	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

// startSecretServiceDaemon starts `gopass secret-service serve` as a
// subprocess and returns its log buffer. Subprocesses inherit the test
// environment, so GOPASS_HOMEDIR, GNUPGHOME and DBUS_SESSION_BUS_ADDRESS apply.
func startSecretServiceDaemon(t *testing.T, ts *tester) *bytes.Buffer {
	t.Helper()

	logs := &bytes.Buffer{}

	daemon := exec.Command(ts.Binary, "secret-service", "serve", "--prefix", "secret-service")
	daemon.Stdout = logs
	daemon.Stderr = logs
	require.NoError(t, daemon.Start())

	t.Cleanup(func() {
		if daemon.Process != nil {
			_ = daemon.Process.Kill()
		}
		_ = daemon.Wait()
	})

	return logs
}

// waitForSecretService waits until some peer owns org.freedesktop.secrets.
func waitForSecretService(t *testing.T, conn *dbus.Conn, logs *bytes.Buffer) {
	t.Helper()

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		var hasOwner bool
		if err := conn.BusObject().Call("org.freedesktop.DBus.NameHasOwner", 0, secretservice.ServiceName).Store(&hasOwner); err == nil && hasOwner {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}

	t.Fatalf("secret-service never acquired %s; daemon output:\n%s", secretservice.ServiceName, logs.String())
}

// openSession opens a plain session and registers its cleanup.
func openSession(t *testing.T, conn *dbus.Conn) dbus.ObjectPath {
	t.Helper()

	var (
		output  dbus.Variant
		session dbus.ObjectPath
	)

	svc := conn.Object(secretservice.ServiceName, secretservice.ServicePath)
	err := svc.Call(secretservice.ServiceIface+".OpenSession", 0, secretservice.AlgorithmPlain, dbus.MakeVariant([]byte{})).Store(&output, &session)
	require.NoError(t, err, "OpenSession failed")
	require.NotEqual(t, secretservice.NullPath, session)

	t.Cleanup(func() {
		_ = conn.Object(secretservice.ServiceName, session).Call(secretservice.SessionIface+".Close", 0).Err
	})

	return session
}

// TestSecretServiceRoundTrip starts the daemon on a private bus and exercises
// the full Secret Service API with the "plain" transport algorithm, then
// verifies the secrets are visible through the regular gopass CLI.
func TestSecretServiceRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	address, stopBus := startPrivateBus(t)
	defer stopBus()

	t.Setenv("DBUS_SESSION_BUS_ADDRESS", address)

	logs := startSecretServiceDaemon(t, ts)

	conn := connectBus(t, address)
	waitForSecretService(t, conn, logs)

	svc := conn.Object(secretservice.ServiceName, secretservice.ServicePath)
	session := openSession(t, conn)

	// The default collection alias must resolve.
	var defaultPath dbus.ObjectPath
	require.NoError(t, svc.Call(secretservice.ServiceIface+".ReadAlias", 0, "default").Store(&defaultPath))
	require.NotEqual(t, secretservice.NullPath, defaultPath, "default alias not set")

	// The session collection alias must resolve too.
	var sessionCollPath dbus.ObjectPath
	require.NoError(t, svc.Call(secretservice.ServiceIface+".ReadAlias", 0, secretservice.SessionCollectionName).Store(&sessionCollPath))
	assert.Equal(t, secretservice.CollectionDBusPath(secretservice.SessionCollectionName), sessionCollPath)

	coll := conn.Object(secretservice.ServiceName, defaultPath)

	// Create an item through the D-Bus API.
	props := map[string]dbus.Variant{
		secretservice.ItemIface + ".Label":      dbus.MakeVariant("GitHub Token"),
		secretservice.ItemIface + ".Attributes": dbus.MakeVariant(map[string]string{"service": "github.com", "username": "octocat"}),
	}
	secret := secretservice.Secret{
		Session:     session,
		Value:       []byte("s3cr3t-value"),
		ContentType: "text/plain",
	}

	var (
		itemPath   dbus.ObjectPath
		promptPath dbus.ObjectPath
	)
	require.NoError(t, coll.Call(secretservice.CollectionIface+".CreateItem", 0, props, secret, false).Store(&itemPath, &promptPath))
	assert.Equal(t, secretservice.NullPath, promptPath, "CreateItem should complete without a prompt")
	require.NotEqual(t, secretservice.NullPath, itemPath)

	// Read it back through the D-Bus API.
	var got secretservice.Secret
	require.NoError(t, conn.Object(secretservice.ServiceName, itemPath).Call(secretservice.ItemIface+".GetSecret", 0, session).Store(&got))
	assert.Equal(t, "s3cr3t-value", string(got.Value))
	assert.Equal(t, "text/plain", got.ContentType)

	// The item must be searchable by attribute.
	var unlocked, locked []dbus.ObjectPath
	require.NoError(t, svc.Call(secretservice.ServiceIface+".SearchItems", 0, map[string]string{"service": "github.com"}).Store(&unlocked, &locked))
	assert.Contains(t, unlocked, itemPath)
	assert.Empty(t, locked)

	// The item must be visible through the gopass CLI.
	id := string(itemPath[strings.LastIndex(string(itemPath), "/")+1:])
	out, err := ts.run("show secret-service/default/" + id)
	require.NoError(t, err, "gopass show failed:\n%s", out)
	assert.Contains(t, out, "s3cr3t-value")

	// A secret written by the CLI must be readable through D-Bus.
	out, err = ts.runCmd([]string{ts.Binary, "insert", "--multiline", "secret-service/default/cli-inserted"}, []byte("cli-value\n---\nsource: cli\n"))
	require.NoError(t, err, "gopass insert failed:\n%s", out)

	require.NoError(t, svc.Call(secretservice.ServiceIface+".SearchItems", 0, map[string]string{"source": "cli"}).Store(&unlocked, &locked))
	require.Len(t, unlocked, 1, "CLI-inserted secret not found via SearchItems")

	var cliSecret secretservice.Secret
	require.NoError(t, conn.Object(secretservice.ServiceName, unlocked[0]).Call(secretservice.ItemIface+".GetSecret", 0, session).Store(&cliSecret))
	assert.Equal(t, "cli-value", string(cliSecret.Value))
}

// TestSecretServiceSessionCollection verifies that the volatile session
// collection is served from memory and never touches the gopass store.
func TestSecretServiceSessionCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	address, stopBus := startPrivateBus(t)
	defer stopBus()

	t.Setenv("DBUS_SESSION_BUS_ADDRESS", address)

	logs := startSecretServiceDaemon(t, ts)

	conn := connectBus(t, address)
	waitForSecretService(t, conn, logs)

	svc := conn.Object(secretservice.ServiceName, secretservice.ServicePath)
	session := openSession(t, conn)

	// The session collection is not part of Service.Collections.
	var collections []dbus.ObjectPath
	require.NoError(t, svc.StoreProperty(secretservice.ServiceIface+".Collections", &collections))
	assert.NotContains(t, collections, secretservice.CollectionDBusPath(secretservice.SessionCollectionName))

	// libsecret resolves aliases to /org/freedesktop/secrets/aliases/<alias>,
	// so the Collection interface must be served there as well.
	sessionColl := conn.Object(secretservice.ServiceName, secretservice.AliasDBusPath(secretservice.SessionCollectionName))

	// Store a transient secret.
	props := map[string]dbus.Variant{
		secretservice.ItemIface + ".Label": dbus.MakeVariant("Transient"),
	}
	secret := secretservice.Secret{
		Session: session,
		Value:   []byte("volatile-value"),
	}

	var itemPath, promptPath dbus.ObjectPath
	require.NoError(t, sessionColl.Call(secretservice.CollectionIface+".CreateItem", 0, props, secret, false).Store(&itemPath, &promptPath))
	require.NotEqual(t, secretservice.NullPath, itemPath)

	var got secretservice.Secret
	require.NoError(t, conn.Object(secretservice.ServiceName, itemPath).Call(secretservice.ItemIface+".GetSecret", 0, session).Store(&got))
	assert.Equal(t, "volatile-value", string(got.Value))

	// Nothing may have been written to the gopass store.
	out, _ := ts.run("ls secret-service")
	assert.NotContains(t, out, secretservice.SessionCollectionName, "session collection leaked into the gopass store:\n%s", out)

	// The session collection cannot be deleted.
	var delPrompt dbus.ObjectPath
	err := sessionColl.Call(secretservice.CollectionIface+".Delete", 0).Store(&delPrompt)
	assert.Error(t, err, "deleting the session collection must fail")
}

// TestSecretServiceStatus checks the status subcommand against a private bus.
func TestSecretServiceStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	address, stopBus := startPrivateBus(t)
	defer stopBus()

	t.Setenv("DBUS_SESSION_BUS_ADDRESS", address)

	out, err := ts.run("secret-service status")
	require.NoError(t, err)
	assert.Contains(t, out, "no provider owns")

	logs := startSecretServiceDaemon(t, ts)

	conn := connectBus(t, address)
	waitForSecretService(t, conn, logs)

	out, err = ts.run("secret-service status")
	require.NoError(t, err)
	assert.Contains(t, out, "is owned")
}
