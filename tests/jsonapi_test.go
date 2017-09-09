package tests

import (
	"bytes"
	"testing"
	"io"

	"encoding/binary"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getMessageLength(t *testing.T, msg []byte) int {
	var length uint32
	buf := bytes.NewBuffer(msg)
	err := binary.Read(buf, binary.LittleEndian, &length)
	assert.NoError(t, err)
	return int(length)
}

func readAndVerifyMessageLength(t *testing.T, rawMessage []byte) string {
	stdin := bytes.NewReader(rawMessage)
	lenBytes := make([]byte, 4)

	_, err := stdin.Read(lenBytes)
	assert.NoError(t, err)

	length := getMessageLength(t, lenBytes)
	assert.NoError(t, err)
	assert.Equal(t, len(rawMessage)-4, length)

	msgBytes := make([]byte, length)
	_, err = stdin.Read(msgBytes)
	assert.NoError(t, err)
	return string(msgBytes)
}

func writeMessageWithLength(message string) io.Reader {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.LittleEndian, uint32(len(message)))
	buffer.WriteString(message)
	return buffer
}

func getMessageResponse(t *testing.T, ts *tester, message string) string {
	out, err := ts.runWithInputReader("jsonapi", writeMessageWithLength(message))
	assert.NoError(t, err)
	return readAndVerifyMessageLength(t, out)
}

func TestJSONAPI(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()
	ts.initSecrets("awesomePrefix/")

	out, err := ts.runCmd([]string{ts.Binary, "insert", "awesomePrefix/fixed/yamllogin"}, []byte("thesecret\n---\nlogin: muh"))
	require.NoError(ts.t, err, "failed to insert password:\n%s", out)

	out, err = ts.runCmd([]string{ts.Binary, "insert", "awesomePrefix/fixed/yamlother"}, []byte("thesecret\n---\nother: meh"))
	require.NoError(ts.t, err, "failed to insert password:\n%s", out)

	// message has length specified but invalid json
	strout, err := ts.runWithInput("jsonapi", "1234Xabcd")
	assert.NoError(t, err)
	assert.Equal(t, "{\"error\":\"invalid character 'X' looking for beginning of value\"}", readAndVerifyMessageLength(t, strout))

	// empty message, no json to parse
	response := getMessageResponse(t, ts, "")
	assert.Equal(t, "{\"error\":\"unexpected end of JSON input\"}", response)

	// message with empty object
	response = getMessageResponse(t, ts, "{}")
	assert.Equal(t, "{\"error\":\"Unknown message of type \"}", response)

	// query for keys without any matching
	response = getMessageResponse(t, ts, "{\"type\":\"query\",\"query\":\"notfound\"}")
	assert.Equal(t, "[]", response)

	// query for keys with matching one
	response = getMessageResponse(t, ts, "{\"type\":\"query\",\"query\":\"foo\"}")
	assert.Equal(t, "[\"awesomePrefix/foo/bar\"]", response)

	// query for keys with matching multiple
	response = getMessageResponse(t, ts, "{\"type\":\"query\",\"query\":\"yaml\"}")
	assert.Equal(t, "[\"awesomePrefix/fixed/yamllogin\",\"awesomePrefix/fixed/yamlother\"]", response)

	// get username / password for key without value in yaml
	response = getMessageResponse(t, ts, "{\"type\":\"getLogin\",\"entry\":\"awesomePrefix/fixed/secret\"}")
	assert.Equal(t, "{\"username\":\"secret\",\"password\":\"moar\"}", response)

	// get username / password for key with login in yaml
	response = getMessageResponse(t, ts, "{\"type\":\"getLogin\",\"entry\":\"awesomePrefix/fixed/yamllogin\"}")
	assert.Equal(t, "{\"username\":\"muh\",\"password\":\"thesecret\"}", response)

	// get username / password for key with no login in yaml (fallback)
	response = getMessageResponse(t, ts, "{\"type\":\"getLogin\",\"entry\":\"awesomePrefix/fixed/yamlother\"}")
	assert.Equal(t, "{\"username\":\"yamlother\",\"password\":\"thesecret\"}", response)
}
