package jsonapi

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/gopasspw/gopass/internal/clipboard"
	"github.com/gopasspw/gopass/internal/otp"
	"github.com/gopasspw/gopass/internal/secrets"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/pwgen"

	"github.com/pkg/errors"
)

var (
	sep = "/"
)

func (api *API) respondMessage(ctx context.Context, msgBytes []byte) error {
	var message messageType
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return errors.Wrapf(err, "failed to unmarshal JSON message")
	}

	switch message.Type {
	case "query":
		return api.respondQuery(ctx, msgBytes)
	case "queryHost":
		return api.respondHostQuery(ctx, msgBytes)
	case "getLogin":
		return api.respondGetLogin(ctx, msgBytes)
	case "getData":
		return api.respondGetData(ctx, msgBytes)
	case "create":
		return api.respondCreateEntry(ctx, msgBytes)
	case "copyToClipboard":
		return api.respondCopyToClipboard(ctx, msgBytes)
	case "getVersion":
		return api.respondGetVersion()
	default:
		return fmt.Errorf("unknown message of type %s", message.Type)
	}
}

func (api *API) respondHostQuery(ctx context.Context, msgBytes []byte) error {
	var message queryHostMessage
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return errors.Wrapf(err, "failed to unmarshal JSON message")
	}

	l, err := api.Store.List(ctx)
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

func (api *API) respondQuery(ctx context.Context, msgBytes []byte) error {
	var message queryMessage
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return errors.Wrapf(err, "failed to unmarshal JSON message")
	}

	l, err := api.Store.List(ctx)
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

	sec, err := api.Store.Get(ctx, message.Entry, "latest")
	if err != nil {
		return errors.Wrapf(err, "failed to get secret")
	}

	return sendSerializedJSONMessage(loginResponse{
		Username: api.getUsername(message.Entry, sec),
		Password: sec.Password(),
	}, api.Writer)
}

func (api *API) respondGetData(ctx context.Context, msgBytes []byte) error {
	var message getDataMessage
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return errors.Wrapf(err, "failed to unmarshal JSON message")
	}

	sec, err := api.Store.Get(ctx, message.Entry, "latest")
	if err != nil {
		return errors.Wrapf(err, "failed to get secret")
	}

	keys := sec.Keys()
	responseData := make(map[string]string, len(keys))
	for _, k := range keys {
		v, ok := sec.Get(k)
		if !ok {
			continue
		}
		responseData[k] = v
	}
	currentTotp, _, err := otp.Calculate("_", sec)
	if err == nil {
		responseData["current_totp"] = currentTotp.OTP()
	}

	converted := convertMixedMapInterfaces(interface{}(responseData))
	return sendSerializedJSONMessage(converted, api.Writer)
}

func (api *API) getUsername(name string, sec gopass.Secret) string {
	// look for a meta-data entry containing the username first
	for _, key := range []string{"login", "username", "user"} {
		if v, ok := sec.Get(key); ok && v != "" {
			return v
		}
	}

	// if no meta-data was found return the name of the secret itself
	// as the username, e.g. providers/amazon.com/foobar -> foobar
	if strings.Contains(name, sep) {
		return path.Base(name)
	}

	return ""
}

func (api *API) respondCreateEntry(ctx context.Context, msgBytes []byte) error {
	var message createEntryMessage
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return errors.Wrapf(err, "failed to unmarshal JSON message")
	}

	if _, err := api.Store.Get(ctx, message.Name, "latest"); err == nil {
		return fmt.Errorf("secret %s already exists", message.Name)
	}

	if message.Generate {
		message.Password = pwgen.GeneratePassword(message.PasswordLength, message.UseSymbols)
	}

	sec := secrets.New()
	sec.SetPassword(message.Password)
	if len(message.Login) > 0 {
		sec.Set("user", message.Login)
	}
	if err := api.Store.Set(ctx, message.Name, sec); err != nil {
		return errors.Wrapf(err, "failed to store secret")
	}

	return sendSerializedJSONMessage(loginResponse{
		Username: message.Login,
		Password: message.Password,
	}, api.Writer)
}

func (api *API) respondGetVersion() error {
	return sendSerializedJSONMessage(getVersionMessage{
		Version: api.Version.String(),
		Major:   api.Version.Major,
		Minor:   api.Version.Minor,
		Patch:   api.Version.Patch,
	}, api.Writer)
}

func (api *API) respondCopyToClipboard(ctx context.Context, msgBytes []byte) error {
	var message copyToClipboard
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return errors.Wrapf(err, "failed to unmarshal JSON message")
	}

	sec, err := api.Store.Get(ctx, message.Entry, "latest")
	if err != nil {
		return errors.Wrapf(err, "failed to get secret")
	}
	var val string
	if message.Key == "" {
		val = sec.Password()
	} else {
		val, _ = sec.Get(message.Key)
	}

	if val == "" {
		return fmt.Errorf("entry not found")
	}

	if err := clipboard.CopyTo(ctx, message.Entry, []byte(val)); err != nil {
		return errors.Wrapf(err, "failed to copy to clipboard")
	}

	return sendSerializedJSONMessage(statusResponse{
		Status: "ok",
	}, api.Writer)
}
