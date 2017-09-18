package jsonapi

import (
	"context"

	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/justwatchcom/gopass/store/secret"
	"github.com/pkg/errors"
)

var (
	sep = string(filepath.Separator)
)

func (api *API) respondMessage(ctx context.Context, msgBytes []byte) error {
	var message messageType
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return errors.Wrapf(err, "failed to unmarshal JSON message")
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
		return errors.Wrapf(err, "failed to unmarshal JSON message")
	}

	l, err := api.Store.List(0)
	if err != nil {
		return errors.Wrapf(err, "failed to list store")
	}
	choices := make([]string, 0, 10)

	for !isPublicSuffix(message.Host) {
		// only query for paths and files in the store fully matching the hostname.
		reQuery := fmt.Sprintf("(^|.*/)%s($|/.*)", regexSafeLower(message.Host))
		if err := searchAndAppendChoices(reQuery, l, &choices); err != nil {
			return errors.Wrapf(err, "failed to append search results")
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
		return errors.Wrapf(err, "failed to unmarshal JSON message")
	}

	l, err := api.Store.List(0)
	if err != nil {
		return errors.Wrapf(err, "failed to list store")
	}

	choices := make([]string, 0, 10)
	reQuery := fmt.Sprintf(".*%s.*", regexSafeLower(message.Query))
	if err := searchAndAppendChoices(reQuery, l, &choices); err != nil {
		return errors.Wrapf(err, "failed to append search results")
	}

	return sendSerializedJSONMessage(choices, api.Writer)
}

func searchAndAppendChoices(reQuery string, list []string, choices *[]string) error {
	re, err := regexp.Compile(reQuery)
	if err != nil {
		return errors.Wrapf(err, "failed to compile regexp '%s': %s", reQuery, err)
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
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return errors.Wrapf(err, "failed to unmarshal JSON message")
	}

	sec, err := api.Store.Get(ctx, message.Entry)
	if err != nil {
		return errors.Wrapf(err, "failed to get secret")
	}

	return sendSerializedJSONMessage(loginResponse{
		Username: api.getUsername(message.Entry, sec),
		Password: sec.Password(),
	}, api.Writer)
}

func (api *API) getUsername(name string, sec *secret.Secret) string {
	// look for a meta-data entry containing the username first
	for _, key := range []string{"login", "username", "user"} {
		value, err := sec.Value(key)
		if err != nil {
			continue
		}
		return value
	}

	// if no meta-data was found return the name of the secret itself
	// as the username, e.g. providers/amazon.com/foobar -> foobar
	if strings.Contains(name, sep) {
		return filepath.Base(name)
	}

	return ""
}
