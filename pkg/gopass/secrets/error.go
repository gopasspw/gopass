package secrets

// PermanentError signal that parsing should not attempt other formats.
type PermanentError struct {
	Err error
}

func (p *PermanentError) Error() string {
	return p.Err.Error()
}
