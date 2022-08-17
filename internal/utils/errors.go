package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
}

type HttpError interface {
	error

	Code() int
	Details() *ErrorResponse
}

type httpError struct {
	code    int
	details *ErrorResponse
	message string
}

func NewHttpError(code int, details *ErrorResponse) error {
	message := fmt.Sprintf("request failed with code %d", code)

	if details != nil {
		if details.Code != "" {
			message += " - code: " + details.Code
		}

		if details.Message != "" {
			message += " - reason: " + details.Message
		}

		if len(details.Errors) > 0 {
			message += " - details: " + strings.Join(details.Errors, " - ")
		}
	}

	return &httpError{
		code:    code,
		details: details,
		message: message,
	}
}

func (h *httpError) Code() int {
	return h.code
}

func (h *httpError) Details() *ErrorResponse {
	return h.details
}

func (h *httpError) Error() string {
	return h.message
}

func HttpErrorDecoder(value interface{}) Decoder {
	return func(code int, bytes []byte) error {
		if code >= http.StatusOK && code <= http.StatusIMUsed {
			if err := json.Unmarshal(bytes, value); err != nil {
				return err
			}
		} else {
			details := &ErrorResponse{}

			if err := json.Unmarshal(bytes, details); err != nil {
				return err
			}

			return NewHttpError(code, details)
		}

		return nil
	}
}

func IsHttpError(err error, code int) bool {
	httpError, ok := err.(HttpError)
	return ok && httpError.Code() == code
}
