//go:build linux

package secretservice

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
)

// Item implements org.freedesktop.Secret.Item.
type Item struct {
	svc        *Service
	store      Store
	collection string
	id         string
	path       dbus.ObjectPath

	props *prop.Properties
}

// itemKey identifies an exported item.
func itemKey(collection, id string) string {
	return collection + "/" + id
}

// ensureItem returns the exported item object, creating and exporting it on
// first use.
func (s *Service) ensureItem(ctx context.Context, collection, id string) (*Item, error) {
	key := itemKey(collection, id)

	s.mu.Lock()
	if it, ok := s.items[key]; ok {
		s.mu.Unlock()

		return it, nil
	}
	s.mu.Unlock()

	store := s.storeFor(collection)

	data, err := store.GetItem(ctx, collection, id)
	if err != nil {
		return nil, err
	}

	it := &Item{
		svc:        s,
		store:      store,
		collection: collection,
		id:         id,
		path:       ItemDBusPath(collection, id),
	}

	if err := s.conn.Export(it, it.path, ItemIface); err != nil {
		return nil, fmt.Errorf("export item %s: %w", key, err)
	}

	if err := s.conn.Export(introspectable(itemIntrospection), it.path, IntrospectableIface); err != nil {
		return nil, fmt.Errorf("export item introspection %s: %w", key, err)
	}

	if err := it.exportProps(ctx, data); err != nil {
		_ = s.conn.Export(nil, it.path, ItemIface)

		return nil, err
	}

	s.mu.Lock()
	s.items[key] = it
	s.mu.Unlock()

	return it, nil
}

// exportProps exports the item's D-Bus properties.
func (i *Item) exportProps(ctx context.Context, data *ItemData) error {
	attrs := data.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}

	props, err := prop.Export(i.svc.conn, i.path, prop.Map{
		ItemIface: {
			"Locked":     {Value: i.locked(ctx), Writable: false, Emit: prop.EmitTrue},
			"Attributes": {Value: attrs, Writable: true, Emit: prop.EmitTrue, Callback: i.setAttributes},
			"Label":      {Value: data.Label, Writable: true, Emit: prop.EmitTrue, Callback: i.setLabel},
			"Created":    {Value: uint64(data.Created.Unix()), Writable: false, Emit: prop.EmitConst},
			"Modified":   {Value: uint64(data.Modified.Unix()), Writable: false, Emit: prop.EmitConst},
		},
	})
	if err != nil {
		return fmt.Errorf("export item properties: %w", err)
	}
	i.props = props

	return nil
}

// locked reports whether the item's collection is currently locked.
func (i *Item) locked(ctx context.Context) bool {
	data, err := i.store.GetCollection(ctx, i.collection)
	if err != nil {
		return false
	}

	return data.Locked
}

// refreshLocked republishes the item's Locked property.
func (i *Item) refreshLocked(ctx context.Context) {
	if i.props == nil {
		return
	}
	i.props.SetMust(ItemIface, "Locked", i.locked(ctx))
}

// refreshProps reloads and republishes the item's mutable properties.
func (i *Item) refreshProps(ctx context.Context) {
	if i.props == nil {
		return
	}

	data, err := i.store.GetItem(ctx, i.collection, i.id)
	if err != nil {
		return
	}

	attrs := data.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}
	i.props.SetMust(ItemIface, "Attributes", attrs)
	i.props.SetMust(ItemIface, "Label", data.Label)
	i.props.SetMust(ItemIface, "Modified", uint64(data.Modified.Unix()))
	i.props.SetMust(ItemIface, "Locked", i.locked(ctx))
}

// load returns the item's current data from the store.
func (i *Item) load(ctx context.Context) (*ItemData, error) {
	return i.store.GetItem(ctx, i.collection, i.id)
}

// setAttributes persists an attribute change made via the Properties
// interface.
func (i *Item) setAttributes(ch *prop.Change) *dbus.Error {
	attrs, ok := ch.Value.(map[string]string)
	if !ok {
		return errUnsupported(fmt.Errorf("attributes must be a string map"))
	}

	ctx := context.Background()
	data, err := i.load(ctx)
	if err != nil {
		return errNotFound(err)
	}
	data.Attributes = attrs

	if err := i.store.UpdateItem(ctx, i.collection, i.id, data); err != nil {
		return errUnsupported(err)
	}

	i.svc.emitItemChanged(i.collection, i.path)

	return nil
}

// setLabel persists a label change made via the Properties interface.
func (i *Item) setLabel(ch *prop.Change) *dbus.Error {
	label, _ := ch.Value.(string)

	ctx := context.Background()
	data, err := i.load(ctx)
	if err != nil {
		return errNotFound(err)
	}
	data.Label = label

	if err := i.store.UpdateItem(ctx, i.collection, i.id, data); err != nil {
		return errUnsupported(err)
	}

	i.svc.emitItemChanged(i.collection, i.path)

	return nil
}

// GetSecret implements org.freedesktop.Secret.Item.GetSecret.
func (i *Item) GetSecret(sessionPath dbus.ObjectPath) (Secret, *dbus.Error) {
	ctx := context.Background()

	coll, err := i.store.GetCollection(ctx, i.collection)
	if err != nil {
		return Secret{}, errNotFound(err)
	}
	if coll.Locked {
		return Secret{}, dbusError(errIsLocked, fmt.Errorf("collection is locked: %s", i.collection))
	}

	sess, err := i.svc.sessions.get(sessionPath)
	if err != nil {
		return Secret{}, dbusError(errNoSession, err)
	}

	data, err := i.load(ctx)
	if err != nil {
		return Secret{}, errNotFound(err)
	}

	params, ciphertext, err := sess.encrypt(data.Secret)
	if err != nil {
		return Secret{}, errUnsupported(err)
	}

	i.svc.notifyAccess(ctx, data.Label)

	return Secret{
		Session:     sessionPath,
		Parameters:  params,
		Value:       ciphertext,
		ContentType: data.ContentType,
	}, nil
}

// SetSecret implements org.freedesktop.Secret.Item.SetSecret.
func (i *Item) SetSecret(secret Secret) *dbus.Error {
	ctx := context.Background()

	coll, err := i.store.GetCollection(ctx, i.collection)
	if err != nil {
		return errNotFound(err)
	}
	if coll.Locked {
		return dbusError(errIsLocked, fmt.Errorf("collection is locked: %s", i.collection))
	}

	value, err := i.svc.decryptSecret(secret)
	if err != nil {
		return errUnsupported(err)
	}

	data, err := i.load(ctx)
	if err != nil {
		return errNotFound(err)
	}
	data.Secret = value
	if secret.ContentType != "" {
		data.ContentType = secret.ContentType
	}

	if err := i.store.UpdateItem(ctx, i.collection, i.id, data); err != nil {
		return errUnsupported(err)
	}

	i.svc.emitItemChanged(i.collection, i.path)

	return nil
}

// Delete implements org.freedesktop.Secret.Item.Delete.
func (i *Item) Delete() (dbus.ObjectPath, *dbus.Error) {
	ctx := context.Background()

	if err := i.store.DeleteItem(ctx, i.collection, i.id); err != nil {
		return NullPath, errNotFound(err)
	}

	_ = i.svc.conn.Export(nil, i.path, ItemIface)
	_ = i.svc.conn.Export(nil, i.path, PropertiesIface)
	_ = i.svc.conn.Export(nil, i.path, IntrospectableIface)

	i.svc.mu.Lock()
	delete(i.svc.items, itemKey(i.collection, i.id))
	i.svc.mu.Unlock()

	i.svc.emitCollectionSignal(i.collection, CollectionIface+".ItemDeleted", i.path)
	i.svc.serviceChanged(i.collection)

	if coll, ok := i.svc.collection(i.collection); ok {
		coll.refreshItems(ctx)
	}

	return NullPath, nil
}

// decryptSecret decrypts a D-Bus Secret using the session it references.
func (s *Service) decryptSecret(secret Secret) ([]byte, error) {
	sess, err := s.sessions.get(secret.Session)
	if err != nil {
		return nil, err
	}

	return sess.decrypt(secret.Parameters, secret.Value)
}

// emitItemChanged announces an item property change on its collection.
func (s *Service) emitItemChanged(collection string, p dbus.ObjectPath) {
	s.emitCollectionSignal(collection, CollectionIface+".ItemChanged", p)
	s.serviceChanged(collection)
}
