package trezor

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/rendaw/go-trezor/messages"
)

const REPLEN_V1 = 64

type ProtocolV1 struct {
}

func ProtocolV1New() (Protocol, error) {
	return &ProtocolV1{}, nil
}

func (self *ProtocolV1) SessionBegin(transport Transport) error {
	return nil
}

func (self *ProtocolV1) SessionEnd(transport Transport) error {
	return nil
}

func (self *ProtocolV1) Write(transport Transport, messageType messages.MessageType, data []byte) error {
	dataHeader := [2 + 2 + 4]byte{}
	dataHeader[0] = '#'
	dataHeader[1] = '#'
	binary.BigEndian.PutUint16(dataHeader[2:4], uint16(messageType))
	binary.BigEndian.PutUint32(dataHeader[4:8], uint32(len(data)))
	data = append(dataHeader[:], data...)
	for len(data) > 0 {
		chunk := [REPLEN_V1]byte{}
		chunk[0] = '?'
		dataLen := copy(chunk[1:], data)
		data = data[dataLen:]
		err := transport.WriteChunk(chunk[:])
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *ProtocolV1) Read(transport Transport) (messages.MessageType, []byte, error) {
	chunk, err := transport.ReadChunk()
	if err != nil {
		return 0, nil, err
	}
	messageType, dataLen, data, err := parseFirstV1(self, chunk)
	if err != nil {
		return 0, nil, err
	}
	for uint32(len(data)) < dataLen {
		chunk, err := transport.ReadChunk()
		if err != nil {
			return 0, nil, err
		}
		parsed, err := parseNextV1(self, chunk)
		if err != nil {
			return 0, nil, err
		}
		data = append(data, parsed...)
	}
	return messageType, data[:dataLen], nil
}

func parseFirstV1(proto *ProtocolV1, chunk []byte) (messages.MessageType, uint32, []byte, error) {
	magic := []byte{'?', '#', '#'}
	if !bytes.Equal(chunk[:3], magic) {
		return 0, 0, nil, fmt.Errorf("Expected magic characters %s, got %s", hex.EncodeToString(magic), hex.EncodeToString(chunk[0:3]))
	}
	offset := 3
	messageType := messages.MessageType(binary.BigEndian.Uint16(chunk[offset : offset+2]))
	offset += 2
	dataLen := binary.BigEndian.Uint32(chunk[offset : offset+4])
	offset += 4
	return messageType, dataLen, chunk[offset:], nil
}

func parseNextV1(proto *ProtocolV1, chunk []byte) ([]byte, error) {
	if chunk[0] != '?' {
		return nil, fmt.Errorf("Expected magic character 0x3f, got %s", hex.EncodeToString(chunk[0:1]))
	}
	return chunk[1:], nil
}