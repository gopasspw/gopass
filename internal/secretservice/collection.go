//go:build linux

package secretservice

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Collection implements org.freedesktop.Secret.Collection.
type Collection struct {
	svc   *Service
	store Store
	name  string
	path  dbus.ObjectPath

	props *prop.Properties
}

// ensureCollection returns the exported collection object at its canonical
// path, creating and exporting it on first use.
func (s *Service) ensureCollection(ctx context.Context, name string) (*Collection, error) {
	return s.ensureCollectionFor(ctx, name, s.store)
}

// ensureCollectionFor returns the exported collection object for a specific
// backend, creating and exporting it on first use. It is used for the volatile
// session collection, which is not served by the gopass store.
func (s *Service) ensureCollectionFor(ctx context.Context, name string, store Store) (*Collection, error) {
	return s.ensureCollectionForPath(ctx, name, store, CollectionDBusPath(name))
}

// ensureCollectionForPath exports a collection object at an explicit object
// path. The same collection can be exported at several paths: its canonical
// path plus one per alias.
func (s *Service) ensureCollectionForPath(ctx context.Context, name string, store Store, p dbus.ObjectPath) (*Collection, error) {
	if c, ok := s.collectionAt(p); ok {
		return c, nil
	}

	c := &Collection{
		svc:   s,
		store: store,
		name:  name,
		path:  p,
	}

	if err := s.conn.Export(c, c.path, CollectionIface); err != nil {
		return nil, fmt.Errorf("export collection %s at %s: %w", name, p, err)
	}

	if err := s.conn.Export(introspectable(collectionIntrospection), c.path, IntrospectableIface); err != nil {
		return nil, fmt.Errorf("export collection introspection %s: %w", name, err)
	}

	if err := c.exportProps(ctx); err != nil {
		_ = s.conn.Export(nil, c.path, CollectionIface)

		return nil, err
	}

	s.mu.Lock()
	s.collections[string(c.path)] = c
	s.mu.Unlock()

	return c, nil
}

// collection returns the canonical object of an already-exported collection.
func (s *Service) collection(name string) (*Collection, bool) {
	return s.collectionAt(CollectionDBusPath(name))
}

// exportProps loads the collection's data and exports its D-Bus properties.
func (c *Collection) exportProps(ctx context.Context) error {
	data, err := c.store.GetCollection(ctx, c.name)
	if err != nil {
		return err
	}

	items, err := c.itemPaths(ctx)
	if err != nil {
		return err
	}

	props, err := prop.Export(c.svc.conn, c.path, prop.Map{
		CollectionIface: {
			"Items":    {Value: items, Writable: false, Emit: prop.EmitTrue},
			"Label":    {Value: data.Label, Writable: true, Emit: prop.EmitTrue, Callback: c.setLabel},
			"Locked":   {Value: data.Locked, Writable: false, Emit: prop.EmitTrue},
			"Created":  {Value: uint64(data.Created.Unix()), Writable: false, Emit: prop.EmitConst},
			"Modified": {Value: uint64(data.Modified.Unix()), Writable: false, Emit: prop.EmitConst},
		},
	})
	if err != nil {
		return fmt.Errorf("export collection properties: %w", err)
	}
	c.props = props

	return nil
}

// itemPaths returns the object paths of the collection's items.
func (c *Collection) itemPaths(ctx context.Context) ([]dbus.ObjectPath, error) {
	ids, err := c.store.Items(ctx, c.name)
	if err != nil {
		return nil, err
	}

	paths := make([]dbus.ObjectPath, 0, len(ids))
	for _, id := range ids {
		paths = append(paths, ItemDBusPath(c.name, id))
	}

	return paths, nil
}

// refreshItems recomputes and publishes the Items property on every exported
// object of the collection.
func (c *Collection) refreshItems(ctx context.Context) {
	paths, err := c.itemPaths(ctx)
	if err != nil {
		return
	}

	for _, o := range c.svc.collectionObjects(c.name) {
		if o.props == nil {
			continue
		}
		o.props.SetMust(CollectionIface, "Items", paths)
	}
}

// refreshLocked recomputes and publishes the Locked property on every exported
// object of the collection.
func (c *Collection) refreshLocked(ctx context.Context) {
	data, err := c.store.GetCollection(ctx, c.name)
	if err != nil {
		return
	}

	for _, o := range c.svc.collectionObjects(c.name) {
		if o.props == nil {
			continue
		}
		o.props.SetMust(CollectionIface, "Locked", data.Locked)
	}
}

// setLabel persists a label change made via the Properties interface. The prop
// package updates its own value after this returns; it must not be called
// re-entrantly.
func (c *Collection) setLabel(ch *prop.Change) *dbus.Error {
	label, _ := ch.Value.(string)
	if err := c.store.SetCollectionLabel(context.Background(), c.name, label); err != nil {
		return errUnsupported(err)
	}

	// Keep the other exported objects (aliases) in sync.
	for _, o := range c.svc.collectionObjects(c.name) {
		if o == c || o.props == nil {
			continue
		}
		o.props.SetMust(CollectionIface, "Label", label)
	}

	return nil
}

// Delete implements org.freedesktop.Secret.Collection.Delete.
func (c *Collection) Delete() (dbus.ObjectPath, *dbus.Error) {
	ctx := context.Background()

	// The volatile session collection cannot be deleted; it disappears with
	// the daemon.
	if c.name == SessionCollectionName {
		return NullPath, errUnsupported(fmt.Errorf("the session collection cannot be deleted"))
	}

	if err := c.store.DeleteCollection(ctx, c.name); err != nil {
		return NullPath, errNotFound(err)
	}

	// Drop aliases that pointed at the deleted collection.
	if err := c.svc.removeAliasesFor(ctx, c.name); err != nil {
		debug.Log("secret-service: removing aliases for %s: %s", c.name, err)
	}

	// Unexport every object of the collection (canonical path and aliases) and
	// all of its items.
	objects := c.svc.collectionObjects(c.name)
	paths := c.svc.collectionPaths(c.name)

	c.svc.mu.Lock()
	for key, item := range c.svc.items {
		if item.collection == c.name {
			_ = c.svc.conn.Export(nil, item.path, ItemIface)
			_ = c.svc.conn.Export(nil, item.path, PropertiesIface)
			_ = c.svc.conn.Export(nil, item.path, IntrospectableIface)
			delete(c.svc.items, key)
		}
	}
	for _, o := range objects {
		delete(c.svc.collections, string(o.path))
	}
	c.svc.mu.Unlock()

	for _, p := range paths {
		_ = c.svc.conn.Export(nil, p, CollectionIface)
		_ = c.svc.conn.Export(nil, p, PropertiesIface)
		_ = c.svc.conn.Export(nil, p, IntrospectableIface)

		c.svc.emitServiceSignal(ServiceIface+".CollectionDeleted", p)
	}

	_ = c.svc.refreshCollectionsProperty(ctx)

	return NullPath, nil
}

// SearchItems implements org.freedesktop.Secret.Collection.SearchItems.
func (c *Collection) SearchItems(attributes map[string]string) ([]dbus.ObjectPath, *dbus.Error) {
	ctx := context.Background()

	items, err := c.store.SearchItems(ctx, c.name, attributes)
	if err != nil {
		return nil, errNotFound(err)
	}

	paths := make([]dbus.ObjectPath, 0, len(items))
	for _, it := range items {
		if _, err := c.svc.ensureItem(ctx, c.name, it.ID); err != nil {
			continue
		}
		paths = append(paths, ItemDBusPath(c.name, it.ID))
	}

	return paths, nil
}

// CreateItem implements org.freedesktop.Secret.Collection.CreateItem.
//
// When replace is true and an existing item has exactly the same attributes as
// the new one, that item is overwritten instead of creating a duplicate. This
// matches libsecret's ReplaceItems behaviour.
func (c *Collection) CreateItem(properties map[string]dbus.Variant, secret Secret, replace bool) (dbus.ObjectPath, dbus.ObjectPath, *dbus.Error) {
	ctx := context.Background()

	data, err := c.store.GetCollection(ctx, c.name)
	if err != nil {
		return NullPath, NullPath, errNotFound(err)
	}
	if data.Locked {
		return NullPath, NullPath, dbusError(errIsLocked, fmt.Errorf("collection is locked: %s", c.name))
	}

	value, err := c.svc.decryptSecret(secret)
	if err != nil {
		return NullPath, NullPath, errUnsupported(err)
	}

	item := &ItemData{
		Secret:      value,
		Label:       variantString(properties, ItemIface+".Label"),
		ContentType: secret.ContentType,
		Attributes:  variantAttributes(properties, ItemIface+".Attributes"),
	}
	if item.ContentType == "" {
		item.ContentType = "text/plain"
	}

	// Replacement: update an item with an identical attribute set.
	if replace && len(item.Attributes) > 0 {
		if matched, err := c.store.SearchItems(ctx, c.name, item.Attributes); err == nil && len(matched) == 1 {
			return c.replaceItem(ctx, matched[0].ID, item)
		}
	}

	id, err := c.store.CreateItem(ctx, c.name, item)
	if err != nil {
		return NullPath, NullPath, errUnsupported(err)
	}

	exported, err := c.svc.ensureItem(ctx, c.name, id)
	if err != nil {
		return NullPath, NullPath, errUnsupported(err)
	}

	c.svc.emitCollectionSignal(c.name, CollectionIface+".ItemCreated", exported.path)
	c.svc.serviceChanged(c.name)
	c.refreshItems(ctx)

	return exported.path, NullPath, nil
}

// replaceItem overwrites the item with the given ID and emits the ItemChanged
// signal.
func (c *Collection) replaceItem(ctx context.Context, id string, item *ItemData) (dbus.ObjectPath, dbus.ObjectPath, *dbus.Error) {
	if err := c.store.UpdateItem(ctx, c.name, id, item); err != nil {
		return NullPath, NullPath, errUnsupported(err)
	}

	exported, err := c.svc.ensureItem(ctx, c.name, id)
	if err != nil {
		return NullPath, NullPath, errUnsupported(err)
	}
	exported.refreshProps(ctx)

	c.svc.emitCollectionSignal(c.name, CollectionIface+".ItemChanged", exported.path)
	c.svc.serviceChanged(c.name)
	c.refreshItems(ctx)

	return exported.path, NullPath, nil
}
