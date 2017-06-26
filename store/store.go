package store

// RecipientCallback is a callback to verify the list of recipients
type RecipientCallback func(string, []string) ([]string, error)

// ImportCallback is a callback to ask the user if he wants to import
// a certain recipients public key into his keystore
type ImportCallback func(string) bool

// FsckCallback is a callback to ask the user to confirm certain fsck
// corrective actions
type FsckCallback func(string) bool
