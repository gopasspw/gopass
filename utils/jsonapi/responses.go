package jsonapi

import (
	"context"

	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/justwatchcom/gopass/store"
)

func (api *API) respondMessage(ctx context.Context, msgBytes []byte) error {
	var message messageType
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return err
	}

	switch message.Type {
	case "query":
		return api.respondQuery(msgBytes)
	case "queryHost":
		return api.respondHostQuery(msgBytes)
	case "getLogin":
		return api.respondGetLogin(ctx, msgBytes)
	default:
		return fmt.Errorf("Unknown message of type %s", message.Type)
	}
}

func (api *API) respondHostQuery(msgBytes []byte) error {
	var message queryHostMessage
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return err
	}

	l, err := api.Store.List(0)
	if err != nil {
		return err
	}
	choices := make([]string, 0, 10)

	for !isPublicSuffix(message.Host) {
		// only query for paths and files in the store fully matching the hostname.
		reQuery := fmt.Sprintf("(^|.*/)%s($|/.*)", regexSafeLower(message.Host))
		if err := searchAndAppendChoices(reQuery, l, &choices); err != nil {
			return err
		}
		if len(choices) > 0 {
			break
		} else {
			message.Host = strings.SplitN(message.Host, ".", 2)[1]
		}
	}

	return sendSerializedJSONMessage(choices, api.Writer)
}

func (api *API) respondQuery(msgBytes []byte) error {
	var message queryMessage
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return err
	}

	l, err := api.Store.List(0)
	if err != nil {
		return err
	}

	choices := make([]string, 0, 10)
	reQuery := fmt.Sprintf(".*%s.*", regexSafeLower(message.Query))
	if err := searchAndAppendChoices(reQuery, l, &choices); err != nil {
		return err
	}

	return sendSerializedJSONMessage(choices, api.Writer)
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

func (api *API) respondGetLogin(ctx context.Context, msgBytes []byte) error {
	var message getLoginMessage
	var response loginResponse

	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return err
	}

	secret, err := api.Store.Get(ctx, message.Entry)
	if err != nil {
		return err
	}

	response.Username, err = api.getUsername(ctx, message.Entry)
	if err != nil {
		return err
	}

	response.Password = secret.Password()

	return sendSerializedJSONMessage(response, api.Writer)
}

func (api *API) getUsername(ctx context.Context, name string) (string, error) {
	for _, key := range []string{"login", "username", "user"} {
		secret, err := api.Store.Get(ctx, name)
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
