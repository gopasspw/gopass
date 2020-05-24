package store

// Byter is a minimal secrets write interface
type Byter interface {
	Bytes() ([]byte, error)
}

// Secret is an in-memory secret with a key/value part
// DEPRECATION WARNING: This interface is going to change.
type Secret interface {
	Byter

	Body() string
	Data() map[string]interface{}
	DeleteKey(string) error
	Equal(Secret) bool
	Password() string
	SetBody(string) error
	SetPassword(string)
	SetValue(string, string) error
	String() string
	Value(string) (string, error)
}
