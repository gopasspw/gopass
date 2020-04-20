package tests

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getMessageLength(t *testing.T, msg []byte) int {
	var length uint32
	buf := bytes.NewBuffer(msg)
	err := binary.Read(buf, binary.LittleEndian, &length)
	require.NoError(t, err)
	return int(length)
}

func readAndVerifyMessageLength(t *testing.T, rawMessage []byte) string {
	stdin := bytes.NewReader(rawMessage)
	lenBytes := make([]byte, 4)

	_, err := stdin.Read(lenBytes)
	require.NoError(t, err)

	length := getMessageLength(t, lenBytes)
	require.NoError(t, err)
	assert.Equal(t, len(rawMessage)-4, length)

	msgBytes := make([]byte, length)
	_, err = stdin.Read(msgBytes)
	require.NoError(t, err)
	return string(msgBytes)
}

func writeMessageWithLength(message string) io.Reader {
	buffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(buffer, binary.LittleEndian, uint32(len(message)))
	buffer.WriteString(message)
	return buffer
}

func getMessageResponse(t *testing.T, ts *tester, message string) string {
	out, err := ts.runWithInputReader("jsonapi listen", writeMessageWithLength(message))
	require.NoError(t, err)
	return readAndVerifyMessageLength(t, out)
}

func TestJSONAPI(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()
	ts.initSecrets("awesomePrefix/")

	// message has length specified but invalid json
	strout, err := ts.runWithInput("jsonapi listen", "1234Xabcd")
	require.NoError(t, err)
	assert.Equal(t, "{\"error\":\"incomplete message read\"}", readAndVerifyMessageLength(t, strout))

	// message with empty object
	response := getMessageResponse(t, ts, "{}")
	assert.Equal(t, "{\"error\":\"unknown message of type \"}", response)

	// query for keys with matching one
	response = getMessageResponse(t, ts, "{\"type\":\"query\",\"query\":\"foo\"}")
	assert.Equal(t, "[\"awesomePrefix\\\\foo\\\\bar\"]", response)
}
