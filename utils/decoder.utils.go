package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type HTTPError interface {
	error
	Status() int
}

type err struct {
	code int
	msg  string
}

func (e *err) Error() string { return e.msg }
func (e *err) Status() int   { return e.code }

func newBadRequest(msg string) HTTPError      { return &err{code: 400, msg: msg} }
func newRequestTooLarge(msg string) HTTPError { return &err{code: 413, msg: msg} }
func newInternal(msg string) HTTPError        { return &err{code: 500, msg: msg} }

// DecodeJSON decodes JSON from r into v (v must be a pointer).
// If isStrict==true then unknown JSON fields are rejected.
// Returns HTTPError for client-visible errors so handler can map to proper HTTP status.
func DecodeJSON(v any, r io.Reader, isStrict bool) error {
	dec := json.NewDecoder(r)
	if isStrict {
		dec.DisallowUnknownFields()
	}

	if err := dec.Decode(v); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxErr):
			return newBadRequest(fmt.Sprintf("request body contains badly-formed JSON (at position %d)", syntaxErr.Offset))
		case errors.Is(err, io.ErrUnexpectedEOF):
			return newBadRequest("request body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeErr):
			field := unmarshalTypeErr.Field
			if field == "" {
				field = "unknown"
			}
			return newBadRequest(fmt.Sprintf("request body contains an invalid value for the %q field", field))
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			f := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return newBadRequest(fmt.Sprintf("request body contains unknown field %s", f))
		case errors.Is(err, io.EOF):
			return newBadRequest("request body must not be empty")
		case strings.Contains(err.Error(), "request body too large"):
			return newRequestTooLarge("request body must not be larger than allowed")
		default:
			return newInternal("failed to parse JSON")
		}
	}

	// Ensure there is only one top-level JSON value.
	// Decode(&struct{}{}) should return io.EOF if there is no extra data.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return newBadRequest("request body must only contain a single JSON object")
	}

	return nil
}
