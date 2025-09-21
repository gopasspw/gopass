package ctxutil

import (
	"context"
	"strings"
)

// hasString is a helper function for checking if a string has been set in
// the provided context.
func hasString(ctx context.Context, key contextKey) bool {
	_, ok := ctx.Value(key).(string)

	return ok
}

// hasBool is a helper function for checking if a bool has been set in
// the provided context.
func hasBool(ctx context.Context, key contextKey) bool {
	_, ok := ctx.Value(key).(bool)

	return ok
}

// is is a helper function for returning the value of a bool from the context
// or the provided default.
func is(ctx context.Context, key contextKey, def bool) bool {
	bv, ok := ctx.Value(key).(bool)
	if !ok {
		return def
	}

	return bv
}

// HeadedText is a helper struct for storing a commit message with a subject and body.
type HeadedText struct {
	head string
	body *strings.Builder
}

// SetHead sets the head of the text.
func (h *HeadedText) SetHead(s string) {
	h.head = s
}

// GetHead returns the head of the text.
func (h *HeadedText) GetHead() string {
	return h.head
}

// AddToBody adds a string to the body of the text.
func (h *HeadedText) AddToBody(s string) {
	if h.body == nil {
		var realBody strings.Builder
		h.body = &realBody
		realBody.WriteString(s)

		return
	}
	(*h.body).WriteString("\n" + s)
}

// ClearBody clears the body of the text.
func (h *HeadedText) ClearBody() {
	h.body = nil
}

// GetBody returns the body of the text.
func (h *HeadedText) GetBody() string {
	if h.body == nil {
		return ""
	}

	return (*h.body).String()
}

// HasBody returns true if the body of the text is not empty.
func (h *HeadedText) HasBody() bool {
	ok := h.body != nil

	return ok && (*h.body).Len() > 0
}

// GetText returns the full text, including the head and body.
func (h *HeadedText) GetText() string {
	body := h.GetBody()
	if body == "" && h.head == "" {
		return ""
	}
	if h.head == "" {
		return body
	}
	if body == "" {
		return h.head
	}

	return h.head + "\n\n" + body
}
