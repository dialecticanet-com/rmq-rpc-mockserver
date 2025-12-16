package expectations

import (
	"encoding/json"
	"strings"
)

type Response struct {
	Body json.RawMessage
}

func NewResponse(body []byte) (*Response, error) {
	return &Response{
		Body: body,
	}, nil
}

func (r *Response) FormattedBody(offset int) string {
	raw, _ := json.MarshalIndent(r.Body, strings.Repeat(" ", offset), "  ")
	return string(raw)
}
