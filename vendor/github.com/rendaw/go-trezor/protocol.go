package trezor

import "github.com/rendaw/go-trezor/messages"

type Protocol interface {
	SessionBegin(transport Transport) error
	SessionEnd(transport Transport) error
	Read(transport Transport) (messages.MessageType, []byte, error)
	Write(transport Transport, messageType messages.MessageType, data []byte) error
}
