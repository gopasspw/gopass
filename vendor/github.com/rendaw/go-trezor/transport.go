package trezor

import (
	"github.com/rendaw/go-trezor/messages"

	"github.com/golang/protobuf/proto"
)

type Transport interface {
	Open() error
	Close() error
	Read() (messages.MessageType, []byte, error)
	Write(message proto.Message) error
	String() string

	// Used by Protocol
	ReadChunk() ([]byte, error)
	WriteChunk([]byte) error
}
