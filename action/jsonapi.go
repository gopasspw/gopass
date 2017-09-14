package action

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"regexp"

	"github.com/justwatchcom/gopass/store"
	"github.com/urfave/cli"
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

// JSONAPI reads a json message on stdin and responds on stdout
func (s *Action) JSONAPI(ctx context.Context, c *cli.Context) error {
	if err := readAndRespond(ctx, s); err != nil {
		var response errorResponse
		response.Error = err.Error()

		if err := sendSerializedJSONMessage(response); err != nil {
			return err
		}
	}
	return nil
}

func readAndRespond(ctx context.Context, s *Action) error {
	message, err := readMessage()
	if message == nil || err != nil {
		return err
	}

	return respondMessage(ctx, s, message)
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

func respondMessage(ctx context.Context, s *Action, msgBytes []byte) error {
	var message messageType
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return err
	}

	switch message.Type {
	case "query":
		return respondQuery(s, msgBytes)
	case "queryHost":
		return respondHostQuery(s, msgBytes)
	case "getLogin":
		return respondGetLogin(ctx, s, msgBytes)
	default:
		return fmt.Errorf("Unknown message of type %s", message.Type)
	}
}

func respondHostQuery(s *Action, msgBytes []byte) error {
	var message queryHostMessage
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return err
	}

	l, err := s.Store.List(0)
	if err != nil {
		return err
	}
	choices := make([]string, 0, 10)

	reQuery := fmt.Sprintf("(^|.*/)%s($|/.*)", regexSafeLower(message.Host))
	if err := searchAndAppendChoices(reQuery, l, &choices); err != nil {
		return err
	}

	for len(choices) == 0 && strings.Count(message.Host, ".") > 1 {
		message.Host = strings.SplitN(message.Host, ".", 2)[1]
		reQuery = fmt.Sprintf("(^|.*/)%s($|/.*)", regexSafeLower(message.Host))
		if err := searchAndAppendChoices(reQuery, l, &choices); err != nil {
			return err
		}
	}

	return sendSerializedJSONMessage(choices)
}

func respondQuery(s *Action, msgBytes []byte) error {
	var message queryMessage
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return err
	}

	l, err := s.Store.List(0)
	if err != nil {
		return err
	}

	choices := make([]string, 0, 10)
	reQuery := fmt.Sprintf(".*%s.*", regexSafeLower(message.Query))
	if err := searchAndAppendChoices(reQuery, l, &choices); err != nil {
		return err
	}

	return sendSerializedJSONMessage(choices)
}

func regexSafeLower(str string) string {
	return regexp.QuoteMeta(strings.ToLower(str))
}

func searchAndAppendChoices(reQuery string, list []string, choices *[]string) error {
	re, err := regexp.Compile(reQuery)
	if err != nil {
		return err
	}
	for _, value := range list {
		if re.MatchString(strings.ToLower(value)) {
			*choices = append(*choices, value)
		}
	}
	return nil
}

func respondGetLogin(ctx context.Context, s *Action, msgBytes []byte) error {
	var message getLoginMessage
	var response loginResponse

	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return err
	}

	secret, err := s.Store.Get(ctx, message.Entry)
	if err != nil {
		return err
	}

	response.Username, err = getUsername(ctx, s, message.Entry)
	if err != nil {
		return err
	}

	response.Password = secret.Password()

	return sendSerializedJSONMessage(response)
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

func getUsername(ctx context.Context, s *Action, name string) (string, error) {
	for _, key := range []string{"login", "username", "user"} {
		secret, err := s.Store.Get(ctx, name)
		if err != nil {
			return "", err
		}
		value, err := secret.Value(key)
		if err != nil {
			if err != store.ErrYAMLNoKey {
				continue
			}
		} else {
			return value, nil
		}
	}

	if strings.Count(name, "/") >= 1 {
		return filepath.Base(name), nil
	}

	return "", nil
}
