package jsonapi

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundTrip(t *testing.T) {
	a := assert.New(t)
	var receivedMessage queryMessage

	message := queryMessage{Query: "holla"}
	var buffer bytes.Buffer

	err := sendSerializedJSONMessage(message, &buffer)
	a.NoError(err)

	received, err := readMessage(&buffer)
	a.NoError(err)

	a.NoError(json.Unmarshal(received, &receivedMessage))
	a.Equal(message.Query, receivedMessage.Query)
}
