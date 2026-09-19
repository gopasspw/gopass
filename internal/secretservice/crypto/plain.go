package crypto

// plainSession implements the "plain" algorithm: the secret value is passed
// through unmodified. The session bus is a local UNIX socket, so the kernel
// protects the transport.
type plainSession struct{}

// Encrypt returns the plaintext unchanged with no IV.
func (plainSession) Encrypt(plaintext []byte) ([]byte, []byte, error) {
	return nil, plaintext, nil
}

// Decrypt returns the ciphertext unchanged.
func (plainSession) Decrypt(_, ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}

// Close is a no-op for the plain algorithm.
func (plainSession) Close() error {
	return nil
}
