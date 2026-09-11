package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/validation"
)

// Request is what a handler method receives: the HTTP request, the model it
// fills in, and the binding result for a body-bound argument.
type Request struct {
	// HTTP is the underlying request.
	HTTP *http.Request
	// Model is the model the view renders against.
	Model *Model
	// BindingResult carries the binding and validation state of the body
	// argument, as a controller's BindingResult parameter does.
	BindingResult *validation.BindingResult
}

// The failures that map onto Spring MVC's status codes.
var (
	// ErrMissingParameter is what a missing required @RequestParam produces:
	// MissingServletRequestParameterException, answered with 400.
	ErrMissingParameter = errors.New("Required request parameter is not present")
	// ErrNotReadable is HttpMessageNotReadableException, answered with 400.
	ErrNotReadable = errors.New("Required request body is missing or not readable")
	// ErrUnsupportedMediaType is HttpMediaTypeNotSupportedException,
	// answered with 415.
	ErrUnsupportedMediaType = errors.New("Content type not supported")
)

// MissingParameterError names the parameter that was absent.
type MissingParameterError struct{ Name string }

func (e *MissingParameterError) Error() string {
	return fmt.Sprintf("Required String parameter '%s' is not present", e.Name)
}

func (e *MissingParameterError) Unwrap() error { return ErrMissingParameter }

// RequestParam reads a required @RequestParam. A parameter that is present but
// empty satisfies the requirement, exactly as it does in Spring.
func (r *Request) RequestParam(name string) (string, error) {
	values, ok := r.HTTP.URL.Query()[name]
	if !ok || len(values) == 0 {
		if err := r.HTTP.ParseForm(); err == nil {
			if v, ok := r.HTTP.PostForm[name]; ok && len(v) > 0 {
				return v[0], nil
			}
		}
		return "", &MissingParameterError{Name: name}
	}
	return values[0], nil
}

// maxBodyBytes bounds how much of a request body is read, so a malformed
// client cannot exhaust memory. The container imposed an equivalent limit.
const maxBodyBytes = 1 << 20

// RequestBody binds the JSON request body into target, the role
// @RequestBody plus Jackson play.
//
// The media-type check comes first, because Spring picks the message
// converter by Content-Type and answers 415 before it ever looks at the bytes.
func (r *Request) RequestBody(target any) error {
	contentType := r.HTTP.Header.Get("Content-Type")
	if mediaType := strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]); mediaType != "application/json" {
		return ErrUnsupportedMediaType
	}
	body, err := io.ReadAll(io.LimitReader(r.HTTP.Body, maxBodyBytes))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNotReadable, err)
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return ErrNotReadable
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("%w: %v", ErrNotReadable, err)
	}
	return nil
}
