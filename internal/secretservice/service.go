//go:build linux

package secretservice

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
	"github.com/gopasspw/gopass/internal/notify"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Config configures the Secret Service daemon.
type Config struct {
	// Prefix is the gopass path prefix under which secrets are stored.
	Prefix string
	// Replace evicts an existing org.freedesktop.secrets owner.
	Replace bool
	// NotifyOnAccess sends a desktop notification whenever a secret is read.
	NotifyOnAccess bool
	// BusAddress overrides the D-Bus address. Empty means the session bus.
	// It exists mainly for tests.
	BusAddress string
}

// Service implements org.freedesktop.Secret.Service.
type Service struct {
	conn  *dbus.Conn
	store Store
	cfg   Config
	m     mapper

	// sessionStore backs the volatile session collection. Its payloads live
	// in the kernel keyring, never in the gopass store.
	sessionStore Store

	sessions *sessionManager

	mu sync.Mutex
	// collections maps D-Bus object paths to exported Collection objects. A
	// collection has one canonical object plus one object per alias, so it can
	// appear more than once.
	collections map[string]*Collection
	items       map[string]*Item

	serviceProps *prop.Properties
}

// New connects to D-Bus and prepares the service. It does not acquire the bus
// name yet; call Start or Serve for that.
func New(ctx context.Context, cfg Config) (*Service, error) {
	if cfg.Prefix == "" {
		cfg.Prefix = "secret-service"
	}

	conn, err := dial(cfg.BusAddress)
	if err != nil {
		return nil, err
	}

	store, err := NewGopassStore(ctx, cfg.Prefix)
	if err != nil {
		_ = conn.Close()

		return nil, err
	}

	return &Service{
		conn:         conn,
		store:        store,
		sessionStore: newKeyringStore(),
		cfg:          cfg,
		m:            mapper{prefix: cfg.Prefix},
		sessions:     newSessionManager(conn),
		collections:  make(map[string]*Collection),
		items:        make(map[string]*Item),
	}, nil
}

// storeFor returns the backend that serves a collection. The well-known
// session collection is served by the volatile store; every other collection
// lives in the gopass store.
func (s *Service) storeFor(collection string) Store {
	if collection == SessionCollectionName {
		return s.sessionStore
	}

	return s.store
}

// collectionAt returns an already-exported collection object by object path.
func (s *Service) collectionAt(p dbus.ObjectPath) (*Collection, bool) {
	s.mu.Lock()
	c, ok := s.collections[string(p)]
	s.mu.Unlock()

	return c, ok
}

// collectionObjects returns every exported object for a collection name: the
// canonical object plus one object per alias.
func (s *Service) collectionObjects(name string) []*Collection {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]*Collection, 0, 1)
	for _, c := range s.collections {
		if c.name == name {
			out = append(out, c)
		}
	}

	return out
}

// collectionPaths returns every object path a collection is exported at.
func (s *Service) collectionPaths(name string) []dbus.ObjectPath {
	objs := s.collectionObjects(name)
	out := make([]dbus.ObjectPath, 0, len(objs))
	for _, c := range objs {
		out = append(out, c.path)
	}

	return out
}

// emitCollectionSignal emits sig from every object path a collection is
// exported at, so clients that resolved the collection via an alias see the
// signal too.
func (s *Service) emitCollectionSignal(name, sig string, args ...any) {
	for _, p := range s.collectionPaths(name) {
		if err := s.conn.Emit(p, sig, args...); err != nil {
			debug.Log("secret-service: emitting %s on %s: %s", sig, p, err)
		}
	}
}

// emitServiceSignal emits a signal from the service object.
func (s *Service) emitServiceSignal(sig string, args ...any) {
	if err := s.conn.Emit(ServicePath, sig, args...); err != nil {
		debug.Log("secret-service: emitting %s: %s", sig, err)
	}
}

// serviceChanged announces that a collection changed.
func (s *Service) serviceChanged(collection string) {
	s.emitServiceSignal(ServiceIface+".CollectionChanged", CollectionDBusPath(collection))
}

// dial connects to the session bus or to an explicit address.
func dial(address string) (*dbus.Conn, error) {
	if address == "" {
		conn, err := dbus.SessionBus()
		if err != nil {
			return nil, fmt.Errorf("connect to session bus: %w", err)
		}

		return conn, nil
	}

	conn, err := dbus.Dial(address)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", address, err)
	}
	if err := conn.Auth(nil); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("authenticate on %s: %w", address, err)
	}
	if err := conn.Hello(); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("hello on %s: %w", address, err)
	}

	return conn, nil
}

// Start exports the service object, acquires the bus name and ensures the
// default collection exists.
func (s *Service) Start(ctx context.Context) error {
	if err := s.conn.Export(s, ServicePath, ServiceIface); err != nil {
		return fmt.Errorf("export service: %w", err)
	}

	if err := s.conn.Export(introspectable(serviceIntrospection), ServicePath, IntrospectableIface); err != nil {
		return fmt.Errorf("export introspection: %w", err)
	}

	props, err := prop.Export(s.conn, ServicePath, prop.Map{
		ServiceIface: {
			"Collections": {Value: []dbus.ObjectPath{}, Writable: false, Emit: prop.EmitTrue},
		},
	})
	if err != nil {
		return fmt.Errorf("export service properties: %w", err)
	}
	s.serviceProps = props

	flags := dbus.NameFlagDoNotQueue
	if s.cfg.Replace {
		flags |= dbus.NameFlagReplaceExisting
	}

	reply, err := s.conn.RequestName(ServiceName, flags)
	if err != nil {
		return fmt.Errorf("request name %s: %w", ServiceName, err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return fmt.Errorf("name %s is already taken (use --replace to take it over)", ServiceName)
	}

	debug.Log("secret-service: acquired %s", ServiceName)

	if err := s.ensureDefaultCollection(ctx); err != nil {
		return err
	}

	// The volatile session collection is always present and is deliberately
	// not part of the Collections property.
	if _, err := s.ensureCollectionFor(ctx, SessionCollectionName, s.sessionStore); err != nil {
		return fmt.Errorf("export session collection: %w", err)
	}

	if err := s.syncAliases(ctx); err != nil {
		return err
	}

	return s.refreshCollectionsProperty(ctx)
}

// syncAliases exports a Collection object at /org/freedesktop/secrets/aliases/
// for every configured alias and unexports objects whose alias disappeared.
// libsecret resolves collection aliases to that path directly, so this mapping
// is required for compatibility.
func (s *Service) syncAliases(ctx context.Context) error {
	aliases, err := s.store.Aliases(ctx)
	if err != nil {
		aliases = map[string]string{"default": "default"}
	}

	aliases[SessionCollectionName] = SessionCollectionName

	wanted := make(map[string]bool, len(aliases))
	for alias, name := range aliases {
		if name == "" {
			continue
		}

		p := AliasDBusPath(alias)
		wanted[string(p)] = true

		if _, ok := s.collectionAt(p); ok {
			continue
		}

		store := s.storeFor(name)
		if _, err := store.GetCollection(ctx, name); err != nil {
			// The alias points at a collection that no longer exists.
			debug.Log("secret-service: alias %q points at missing collection %q", alias, name)

			continue
		}

		if _, err := s.ensureCollectionForPath(ctx, name, store, p); err != nil {
			debug.Log("secret-service: cannot export alias %q: %s", alias, err)
		}
	}

	// Unexport alias objects that are no longer wanted.
	s.mu.Lock()
	var stale []*Collection

	for path, c := range s.collections {
		if !strings.HasPrefix(path, aliasPathPrefix) {
			continue
		}
		if wanted[path] {
			continue
		}
		stale = append(stale, c)
		delete(s.collections, path)
	}
	s.mu.Unlock()

	for _, c := range stale {
		_ = s.conn.Export(nil, c.path, CollectionIface)
		_ = s.conn.Export(nil, c.path, PropertiesIface)
		_ = s.conn.Export(nil, c.path, IntrospectableIface)
	}

	return nil
}

// Serve starts the service and blocks until ctx is cancelled.
func (s *Service) Serve(ctx context.Context) error {
	if err := s.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()

	return s.Stop()
}

// Stop releases the bus name and closes the store.
func (s *Service) Stop() error {
	s.sessions.closeAll()

	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.sessionStore.Close(closeCtx); err != nil {
		debug.Log("secret-service: closing session store: %s", err)
	}
	if err := s.store.Close(closeCtx); err != nil {
		debug.Log("secret-service: closing store: %s", err)
	}

	if _, err := s.conn.ReleaseName(ServiceName); err != nil {
		debug.Log("secret-service: releasing name: %s", err)
	}

	return s.conn.Close()
}

// OpenSession implements org.freedesktop.Secret.Service.OpenSession.
func (s *Service) OpenSession(algorithm string, input dbus.Variant) (dbus.Variant, dbus.ObjectPath, *dbus.Error) {
	var inputBytes []byte
	if v, ok := input.Value().([]byte); ok {
		inputBytes = v
	}

	sess, output, err := s.sessions.open(algorithm, inputBytes)
	if err != nil {
		return dbus.MakeVariant([]byte{}), NullPath, errUnsupported(err)
	}

	return dbus.MakeVariant(output), sess.path, nil
}

// CreateCollection implements org.freedesktop.Secret.Service.CreateCollection.
func (s *Service) CreateCollection(properties map[string]dbus.Variant, alias string) (dbus.ObjectPath, dbus.ObjectPath, *dbus.Error) {
	ctx := context.Background()

	label := variantString(properties, CollectionIface+".Label")

	name := alias
	if name == "" {
		name = label
	}
	if name == "" {
		name = "collection"
	}
	name = SanitizeName(name)

	if name == SessionCollectionName {
		return NullPath, NullPath, errExists(fmt.Errorf("collection name %q is reserved for the session collection", name))
	}

	if _, err := s.store.GetCollection(ctx, name); err == nil {
		return NullPath, NullPath, errExists(fmt.Errorf("collection already exists: %s", name))
	}

	if err := s.store.CreateCollection(ctx, name, label); err != nil {
		return NullPath, NullPath, errUnsupported(err)
	}

	coll, err := s.ensureCollection(ctx, name)
	if err != nil {
		return NullPath, NullPath, errUnsupported(err)
	}

	if alias != "" {
		if err := s.store.SetAlias(ctx, alias, name); err != nil {
			debug.Log("secret-service: set alias %s: %s", alias, err)
		}
		if err := s.syncAliases(ctx); err != nil {
			debug.Log("secret-service: syncing aliases: %s", err)
		}
	}

	s.emitServiceSignal(ServiceIface+".CollectionCreated", coll.path)
	_ = s.refreshCollectionsProperty(ctx)

	return coll.path, NullPath, nil
}

// SearchItems implements org.freedesktop.Secret.Service.SearchItems.
func (s *Service) SearchItems(attributes map[string]string) ([]dbus.ObjectPath, []dbus.ObjectPath, *dbus.Error) {
	ctx := context.Background()

	results, err := s.store.SearchAllItems(ctx, attributes)
	if err != nil {
		return nil, nil, errNotFound(err)
	}

	// The session collection is searched as well, but is never "locked".
	sessionResults, err := s.sessionStore.SearchAllItems(ctx, attributes)
	if err == nil {
		for name, items := range sessionResults {
			results[name] = append(results[name], items...)
		}
	}

	var unlocked, locked []dbus.ObjectPath
	for collName, items := range results {
		store := s.storeFor(collName)
		collData, _ := store.GetCollection(ctx, collName)
		isLocked := collData != nil && collData.Locked

		for _, it := range items {
			item, err := s.ensureItem(ctx, collName, it.ID)
			if err != nil {
				debug.Log("secret-service: cannot export item %s/%s: %s", collName, it.ID, err)

				continue
			}
			if isLocked {
				locked = append(locked, item.path)
			} else {
				unlocked = append(unlocked, item.path)
			}
		}
	}

	return unlocked, locked, nil
}

// Unlock implements org.freedesktop.Secret.Service.Unlock.
func (s *Service) Unlock(objects []dbus.ObjectPath) ([]dbus.ObjectPath, dbus.ObjectPath, *dbus.Error) {
	ctx := context.Background()

	var unlocked []dbus.ObjectPath
	for _, p := range objects {
		name, err := s.resolveCollectionName(ctx, p)
		if err != nil {
			continue
		}
		if err := s.storeFor(name).UnlockCollection(ctx, name); err != nil {
			continue
		}
		s.refreshLockState(ctx, name)
		unlocked = append(unlocked, p)
	}

	return unlocked, NullPath, nil
}

// Lock implements org.freedesktop.Secret.Service.Lock.
func (s *Service) Lock(objects []dbus.ObjectPath) ([]dbus.ObjectPath, dbus.ObjectPath, *dbus.Error) {
	ctx := context.Background()

	var locked []dbus.ObjectPath
	for _, p := range objects {
		name, err := s.resolveCollectionName(ctx, p)
		if err != nil {
			continue
		}
		if err := s.storeFor(name).LockCollection(ctx, name); err != nil {
			continue
		}
		s.refreshLockState(ctx, name)
		locked = append(locked, p)
	}

	return locked, NullPath, nil
}

// refreshLockState republishes the Locked property of a collection and of every
// exported item it contains.
func (s *Service) refreshLockState(ctx context.Context, name string) {
	if coll, ok := s.collection(name); ok {
		coll.refreshLocked(ctx)
	}

	s.mu.Lock()
	items := make([]*Item, 0, len(s.items))
	for _, item := range s.items {
		if item.collection == name {
			items = append(items, item)
		}
	}
	s.mu.Unlock()

	for _, item := range items {
		item.refreshLocked(ctx)
	}
}

// GetSecrets implements org.freedesktop.Secret.Service.GetSecrets.
func (s *Service) GetSecrets(items []dbus.ObjectPath, sessionPath dbus.ObjectPath) (map[dbus.ObjectPath]Secret, *dbus.Error) {
	sess, err := s.sessions.get(sessionPath)
	if err != nil {
		return nil, dbusError(errNoSession, err)
	}

	ctx := context.Background()
	out := make(map[dbus.ObjectPath]Secret)

	for _, p := range items {
		collection, id, err := parseItemPath(p)
		if err != nil {
			continue
		}

		item, err := s.storeFor(collection).GetItem(ctx, collection, ItemNameFromDBusID(id))
		if err != nil {
			continue
		}

		params, ciphertext, err := sess.encrypt(item.Secret)
		if err != nil {
			continue
		}

		s.notifyAccess(ctx, item.Label)

		out[p] = Secret{
			Session:     sessionPath,
			Parameters:  params,
			Value:       ciphertext,
			ContentType: item.ContentType,
		}
	}

	return out, nil
}

// ReadAlias implements org.freedesktop.Secret.Service.ReadAlias.
func (s *Service) ReadAlias(name string) (dbus.ObjectPath, *dbus.Error) {
	// The session collection is always available under the well-known alias
	// and never appears in the persistent alias map.
	if name == SessionCollectionName {
		return CollectionDBusPath(SessionCollectionName), nil
	}

	collName, err := s.store.GetAlias(context.Background(), name)
	if err != nil {
		return NullPath, nil
	}

	return CollectionDBusPath(collName), nil
}

// SetAlias implements org.freedesktop.Secret.Service.SetAlias.
func (s *Service) SetAlias(name string, collection dbus.ObjectPath) *dbus.Error {
	ctx := context.Background()

	if name == SessionCollectionName {
		return errUnsupported(fmt.Errorf("the session alias is read-only"))
	}

	if collection == NullPath {
		if err := s.store.SetAlias(ctx, name, ""); err != nil {
			return errUnsupported(err)
		}
		if err := s.syncAliases(ctx); err != nil {
			debug.Log("secret-service: syncing aliases: %s", err)
		}

		return nil
	}

	collName, err := parseCollectionPath(collection)
	if err != nil {
		return errNotFound(err)
	}

	if err := s.store.SetAlias(ctx, name, collName); err != nil {
		return errUnsupported(err)
	}
	if err := s.syncAliases(ctx); err != nil {
		debug.Log("secret-service: syncing aliases: %s", err)
	}

	return nil
}

// resolveCollectionName maps a collection, item or alias object path to a
// collection name.
func (s *Service) resolveCollectionName(ctx context.Context, p dbus.ObjectPath) (string, error) {
	if name, err := parseCollectionPath(p); err == nil {
		return name, nil
	}
	if collection, _, err := parseItemPath(p); err == nil {
		return collection, nil
	}

	// Alias paths: /org/freedesktop/secrets/aliases/NAME.
	if alias, ok := IsAliasPath(p); ok {
		return s.store.GetAlias(ctx, alias)
	}

	return "", fmt.Errorf("not a collection, item or alias path: %s", p)
}

// ensureDefaultCollection makes sure a collection exists and is aliased as
// "default".
func (s *Service) ensureDefaultCollection(ctx context.Context) error {
	name, err := s.store.GetAlias(ctx, "default")
	if err != nil {
		name = "default"
	}

	if _, err := s.store.GetCollection(ctx, name); err != nil {
		if err := s.store.CreateCollection(ctx, name, "Default"); err != nil {
			return fmt.Errorf("create default collection: %w", err)
		}
	}

	if _, err := s.ensureCollection(ctx, name); err != nil {
		return err
	}

	if err := s.store.SetAlias(ctx, "default", name); err != nil {
		debug.Log("secret-service: set default alias: %s", err)
	}

	return nil
}

// refreshCollectionsProperty updates the Service.Collections property.
func (s *Service) refreshCollectionsProperty(ctx context.Context) error {
	if s.serviceProps == nil {
		return nil
	}

	names, err := s.store.Collections(ctx)
	if err != nil {
		return err
	}

	paths := make([]dbus.ObjectPath, 0, len(names))
	for _, name := range names {
		if _, err := s.ensureCollection(ctx, name); err != nil {
			continue
		}
		paths = append(paths, CollectionDBusPath(name))
	}

	s.serviceProps.SetMust(ServiceIface, "Collections", paths)

	return nil
}

// removeAliasesFor deletes every alias that points at the given collection.
func (s *Service) removeAliasesFor(ctx context.Context, name string) error {
	aliases, err := s.store.Aliases(ctx)
	if err != nil {
		return err
	}

	for alias, target := range aliases {
		if target != name {
			continue
		}
		if err := s.store.SetAlias(ctx, alias, ""); err != nil {
			return err
		}
	}

	return s.syncAliases(ctx)
}

// notifyAccess sends a desktop notification about a secret read, if the daemon
// was started with --notify-on-access.
func (s *Service) notifyAccess(ctx context.Context, label string) {
	if !s.cfg.NotifyOnAccess {
		return
	}

	if label == "" {
		label = "a secret"
	}

	if err := notify.Notify(ctx, "gopass secret-service", "Read "+label); err != nil {
		debug.Log("secret-service: notification failed: %s", err)
	}
}

// variantString extracts a string from a properties map, tolerating absence.
func variantString(properties map[string]dbus.Variant, key string) string {
	v, ok := properties[key]
	if !ok {
		return ""
	}
	s, _ := v.Value().(string)

	return s
}

// variantAttributes extracts an attribute map from a properties map.
func variantAttributes(properties map[string]dbus.Variant, key string) map[string]string {
	out := make(map[string]string)

	v, ok := properties[key]
	if !ok {
		return out
	}

	switch attrs := v.Value().(type) {
	case map[string]string:
		for k, val := range attrs {
			out[k] = val
		}
	case map[string]dbus.Variant:
		for k, val := range attrs {
			if s, ok := val.Value().(string); ok {
				out[k] = s
			}
		}
	}

	return out
}

// introspectable serves a static introspection XML document.
type introspectable string

// Introspect implements org.freedesktop.DBus.Introspectable.
func (i introspectable) Introspect() (string, *dbus.Error) {
	return string(i), nil
}

const serviceIntrospection = `<node>
  <interface name="org.freedesktop.Secret.Service">
    <method name="OpenSession">
      <arg name="algorithm" type="s" direction="in"/>
      <arg name="input" type="v" direction="in"/>
      <arg name="output" type="v" direction="out"/>
      <arg name="result" type="o" direction="out"/>
    </method>
    <method name="CreateCollection">
      <arg name="properties" type="a{sv}" direction="in"/>
      <arg name="alias" type="s" direction="in"/>
      <arg name="collection" type="o" direction="out"/>
      <arg name="prompt" type="o" direction="out"/>
    </method>
    <method name="SearchItems">
      <arg name="attributes" type="a{ss}" direction="in"/>
      <arg name="unlocked" type="ao" direction="out"/>
      <arg name="locked" type="ao" direction="out"/>
    </method>
    <method name="Unlock">
      <arg name="objects" type="ao" direction="in"/>
      <arg name="unlocked" type="ao" direction="out"/>
      <arg name="prompt" type="o" direction="out"/>
    </method>
    <method name="Lock">
      <arg name="objects" type="ao" direction="in"/>
      <arg name="locked" type="ao" direction="out"/>
      <arg name="prompt" type="o" direction="out"/>
    </method>
    <method name="GetSecrets">
      <arg name="items" type="ao" direction="in"/>
      <arg name="session" type="o" direction="in"/>
      <arg name="secrets" type="a{o(oayays)}" direction="out"/>
    </method>
    <method name="ReadAlias">
      <arg name="name" type="s" direction="in"/>
      <arg name="collection" type="o" direction="out"/>
    </method>
    <method name="SetAlias">
      <arg name="name" type="s" direction="in"/>
      <arg name="collection" type="o" direction="in"/>
    </method>
    <signal name="CollectionCreated">
      <arg name="collection" type="o"/>
    </signal>
    <signal name="CollectionDeleted">
      <arg name="collection" type="o"/>
    </signal>
    <signal name="CollectionChanged">
      <arg name="collection" type="o"/>
    </signal>
    <property name="Collections" type="ao" access="read"/>
  </interface>
</node>`
