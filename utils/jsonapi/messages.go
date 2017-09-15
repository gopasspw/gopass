package jsonapi

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
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

type errorResponse struct {
	Error string `json:"error"`
}

func readMessage() ([]byte, error) {
	stdin := bufio.NewReader(os.Stdin)
	lenBytes := make([]byte, 4)
	_, err := stdin.Read(lenBytes)
	if err != nil {
		return nil, eofReturn(err)
	}

	length, err := getMessageLength(lenBytes)
	if err != nil {
		return nil, err
	}

	msgBytes := make([]byte, length)
	_, err = stdin.Read(msgBytes)
	if err != nil {
		return nil, eofReturn(err)
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

func sendSerializedJSONMessage(message interface{}) error {
	var msgBuf bytes.Buffer

	serialized, err := json.Marshal(message)
	if err != nil {
		return err
	}

	if err := writeMessageLength(serialized); err != nil {
		return err
	}

	_, err = msgBuf.Write(serialized)
	if err != nil {
		return err
	}

	_, err = msgBuf.WriteTo(os.Stdout)
	return err
}

func writeMessageLength(msg []byte) error {
	return binary.Write(os.Stdout, binary.LittleEndian, uint32(len(msg)))
}
