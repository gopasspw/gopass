package store

import (
	"context"
	"fmt"

	"github.com/justwatchcom/gopass/pkg/backend"
	"github.com/justwatchcom/gopass/pkg/tree"
)

// RecipientCallback is a callback to verify the list of recipients
type RecipientCallback func(context.Context, string, []string) ([]string, error)

// ImportCallback is a callback to ask the user if he wants to import
// a certain recipients public key into his keystore
type ImportCallback func(context.Context, string, []string) bool

// FsckCallback is a callback to ask the user to confirm certain fsck
// corrective actions
type FsckCallback func(context.Context, string) bool

// TemplateStore is a store supporting templating operations
type TemplateStore interface {
	GetTemplate(context.Context, string) ([]byte, error)
	HasTemplate(context.Context, string) bool
	ListTemplates(context.Context, string) []string
	LookupTemplate(context.Context, string) ([]byte, bool)
	RemoveTemplate(context.Context, string) error
	SetTemplate(context.Context, string, []byte) error
	TemplateTree(context.Context) (tree.Tree, error)
}

// RecipientStore is a store supporting recipient operations
type RecipientStore interface {
	AddRecipient(context.Context, string) error
	GetRecipients(context.Context, string) ([]string, error)
	RemoveRecipient(context.Context, string) error
	SaveRecipients(context.Context) error
	SetRecipients(context.Context, []string) error
	Recipients(context.Context) []string
	ImportMissingPublicKeys(context.Context) error
	ExportMissingPublicKeys(context.Context, []string) (bool, error)
}

// Store is secrets store
type Store interface {
	fmt.Stringer

	TemplateStore
	RecipientStore

	Fsck(context.Context, string) error
	Path() string
	URL() string
	RCS() backend.RCS
	Crypto() backend.Crypto
	Storage() backend.Storage
	GitInit(context.Context, string, string) error
	Alias() string
	Copy(context.Context, string, string) error
	Delete(context.Context, string) error
	Equals(Store) bool
	Exists(context.Context, string) bool
	Get(context.Context, string) (Secret, error)
	GetRevision(context.Context, string, string) (Secret, error)
	Init(context.Context, string, ...string) error
	Initialized(context.Context) bool
	IsDir(context.Context, string) bool
	List(context.Context, string) ([]string, error)
	ListRevisions(context.Context, string) ([]backend.Revision, error)
	Move(context.Context, string, string) error
	Set(context.Context, string, Secret) error
	Prune(context.Context, string) error
	Valid() bool
}
