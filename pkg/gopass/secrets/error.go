package secrets

// PermanentError signals that parsing should not attempt other formats.
type PermanentError struct {
	Err error
}

// Error returns the underlying error.
func (p *PermanentError) Error() string {
	return p.Err.Error()
}
