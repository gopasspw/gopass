package store

// Secret is an in-memory secret with a key/value part
type Secret interface {
	Body() string
	Bytes() ([]byte, error)
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
