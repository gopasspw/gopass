package xc

import (
	"context"
	"fmt"
)

func (x *XC) Sign(ctx context.Context, message []byte) ([]byte, error) {
	return nil, fmt.Errorf("not yet implemented")
}

func (x *XC) Verify(ctx context.Context, message, signedMessage []byte) ([]byte, error) {
	return nil, fmt.Errorf("not yet implemented")
}
