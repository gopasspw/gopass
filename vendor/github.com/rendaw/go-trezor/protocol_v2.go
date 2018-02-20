package trezor

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/rendaw/go-trezor/messages"
)

const REPLEN_V2 = 64

type ProtocolV2 struct {
	hasSession bool
	session    uint32
}

func ProtocolV2New() (Protocol, error) {
	return &ProtocolV2{
		hasSession: false,
		session:    0,
	}, nil
}

func (self *ProtocolV2) SessionBegin(transport Transport) error {
	chunk := [REPLEN_V2]byte{}
	chunk[0] = 3
	for i := 1; i < len(chunk); i++ {
		chunk[i] = 0
	}
	err := transport.WriteChunk(chunk[:])
	if err != nil {
		return err
	}
	resp, err := transport.ReadChunk()
	if err != nil {
		return err
	}
	err = self.ParseSessionOpen(resp)
	if err != nil {
		return err
	}
	return nil
}

func (self *ProtocolV2) SessionEnd(transport Transport) error {
	if !self.hasSession {
		return nil
	}
	chunk := [REPLEN_V2]byte{}
	chunk[0] = 0x04
	binary.BigEndian.PutUint32(chunk[1:5], self.session)
	err := transport.WriteChunk(chunk[:])
	if err != nil {
		return err
	}
	resp, err := transport.ReadChunk()
	if err != nil {
		return err
	}
	magic := resp[0]
	if magic != 0x04 {
		return fmt.Errorf("Expected session close (0x04), got %d", hex.EncodeToString([]byte{magic}))
	}
	self.hasSession = false
	return nil
}

func (self *ProtocolV2) Write(transport Transport, messageType messages.MessageType, data []byte) error {
	if !self.hasSession {
		return fmt.Errorf("Missing session for v2 protocol")
	}
	dataHeader := [8]byte{}
	binary.BigEndian.PutUint32(dataHeader[0:4], uint32(messageType))
	binary.BigEndian.PutUint32(dataHeader[4:8], uint32(len(data)))
	data = append(dataHeader[:], data...)
	var seq int32 = -1
	for len(data) > 0 {
		var repHeader []byte
		if seq < 0 {
			_repHeader := [5]byte{}
			repHeader = _repHeader[:]
			repHeader[0] = 0x01
			binary.BigEndian.PutUint32(repHeader[1:5], self.session)
		} else {
			_repHeader := [9]byte{}
			repHeader = _repHeader[:]
			repHeader[0] = 0x02
			binary.BigEndian.PutUint32(repHeader[1:5], self.session)
			binary.BigEndian.PutUint32(repHeader[1:5], uint32(seq))
		}
		chunk := [REPLEN_V2]byte{}
		off := 0
		off += copy(chunk[off:], repHeader)
		dataLen := copy(chunk[off:], data)
		err := transport.WriteChunk(chunk[:])
		if err != nil {
			return err
		}
		data = data[dataLen:]
		seq += 1
	}
	return nil
}

func (self *ProtocolV2) Read(transport Transport) (messages.MessageType, []byte, error) {
	if !self.hasSession {
		return 0, nil, fmt.Errorf("Missing session for v2 protocol")
	}

	chunk, err := transport.ReadChunk()
	if err != nil {
		return 0, nil, err
	}
	messageType, dataLen, data, err := ParseFirstV2(self, chunk)
	if err != nil {
		return 0, nil, err
	}

	for uint32(len(data)) < dataLen {
		chunk, err := transport.ReadChunk()
		if err != nil {
			return 0, nil, err
		}
		nextData, err := ParseNextV2(self, chunk)
		if err != nil {
			return 0, nil, err
		}
		data = append(data, nextData...)
	}

	return messageType, data[:dataLen], nil
}

func ParseFirstV2(proto *ProtocolV2, chunk []byte) (messages.MessageType, uint32, []byte, error) {
	offset := 0
	magic := chunk[offset]
	offset += 1
	if magic != 0x01 {
		return 0, 0, nil, fmt.Errorf("Expected magic character 0x01, got %s", hex.EncodeToString([]byte{magic}))
	}
	sessionBytes := chunk[offset : offset+4]
	session := binary.BigEndian.Uint32(sessionBytes)
	offset += 4
	if session != proto.session {
		protoSessionBytes := [4]byte{}
		binary.BigEndian.PutUint32(protoSessionBytes[:], proto.session)
		return 0, 0, nil, fmt.Errorf("Session mismatch, expected %s, got %s", hex.EncodeToString(protoSessionBytes[:]), hex.EncodeToString(sessionBytes))
	}
	messageType := messages.MessageType(binary.BigEndian.Uint32(chunk[offset : offset+4]))
	offset += 4
	dataLen := binary.BigEndian.Uint32(chunk[offset : offset+4])
	offset += 4
	return messageType, dataLen, chunk[offset:], nil
}

func ParseNextV2(proto *ProtocolV2, chunk []byte) ([]byte, error) {
	offset := 0
	magic := chunk[offset]
	offset += 1
	if magic != 0x02 {
		return nil, fmt.Errorf("Expected magic character 0x02, got %s", hex.EncodeToString([]byte{magic}))
	}
	sessionBytes := chunk[offset : offset+4]
	session := binary.BigEndian.Uint32(sessionBytes)
	offset += 4
	if session != proto.session {
		protoSessionBytes := [4]byte{}
		binary.BigEndian.PutUint32(protoSessionBytes[:], proto.session)
		return nil, fmt.Errorf("Session mismatch, expected %s, got %s", hex.EncodeToString(protoSessionBytes[:]), hex.EncodeToString(sessionBytes))
	}
	offset += 4 // skip sequence
	return chunk[offset:], nil
}

func (self *ProtocolV2) ParseSessionOpen(resp []byte) error {
	magic := resp[0]
	session := binary.BigEndian.Uint32(resp[1:5])
	if magic != 0x03 {
		return fmt.Errorf("Expected magic character 0x03, got %s", hex.EncodeToString([]byte{magic}))
	}
	self.session = session
	self.hasSession = true
	return nil
}
