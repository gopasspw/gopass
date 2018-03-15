package jsonapi

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

type messageType struct {
	Type string `json:"type"`
}

type queryMessage struct {
	Query string `json:"query"`
}

type queryHostMessage struct {
	Host string `json:"host"`
}

type getLoginMessage struct {
	Entry string `json:"entry"`
}

type loginResponse struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type createEntryMessage struct {
	Name           string `json:"entry_name"`
	Login          string `json:"login"`
	Password       string `json:"password"`
	PasswordLength int    `json:"length"`
	Generate       bool   `json:"generate"`
	UseSymbols     bool   `json:"use_symbols"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func readMessage(r io.Reader) ([]byte, error) {
	stdin := bufio.NewReader(r)
	lenBytes := make([]byte, 4)
	count, err := stdin.Read(lenBytes)
	if err != nil {
		return nil, eofReturn(err)
	}
	if count != 4 {
		return nil, fmt.Errorf("not enough bytes read to deterimine message size")
	}

	length, err := getMessageLength(lenBytes)
	if err != nil {
		return nil, err
	}

	msgBytes := make([]byte, length)
	count, err = stdin.Read(msgBytes)
	if err != nil {
		return nil, eofReturn(err)
	}
	if count != length {
		return nil, fmt.Errorf("incomplete message read")
	}

	return msgBytes, nil
}

func getMessageLength(msg []byte) (int, error) {
	var length uint32
	buf := bytes.NewBuffer(msg)
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return 0, err
	}

	return int(length), nil
}

func eofReturn(err error) error {
	if err == io.EOF {
		return nil
	}
	return err
}

func sendSerializedJSONMessage(message interface{}, w io.Writer) error {
	// we can't use json.NewEncoder(w).Encode because we need to send the final
	// message length before the actul JSON
	serialized, err := json.Marshal(message)
	if err != nil {
		return err
	}

	if err := writeMessageLength(serialized, w); err != nil {
		return err
	}

	var msgBuf bytes.Buffer
	count, err := msgBuf.Write(serialized)
	if err != nil {
		return err
	}
	if count != len(serialized) {
		return fmt.Errorf("message not fully written")
	}

	wcount, err := msgBuf.WriteTo(w)
	if wcount != int64(len(serialized)) {
		return fmt.Errorf("message not fully written")
	}
	return err
}

func writeMessageLength(msg []byte, w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, uint32(len(msg)))
}
