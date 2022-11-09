package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultTimeout = time.Second * 30
)

type (
	Encoder func() ([]byte, error)
	Decoder func(code int, bytes []byte) error
)

func JsonEncoder(value interface{}) Encoder {
	return func() ([]byte, error) {
		return json.Marshal(value)
	}
}

func JsonDecoder(value interface{}) Decoder {
	return func(code int, bytes []byte) error {
		return json.Unmarshal(bytes, value)
	}
}

type Rest struct {
	Url     string
	Method  string
	Client  *http.Client
	Headers map[string]string
	Queries map[string]string
	Encoder Encoder
	Decoder Decoder
	Code    int
}

func (r *Rest) Execute(ctx context.Context) error {
	var requestBody []byte

	// Encode the request body.
	if r.Encoder != nil {
		if bytes, err := r.Encoder(); err != nil {
			return err
		} else {
			requestBody = bytes
		}
	}

	// Create the request.
	httpRequest, err := http.NewRequest(r.Method, r.Url, bytes.NewReader(requestBody))
	if err != nil {
		return err
	}

	httpRequest = httpRequest.WithContext(ctx)

	// Add any headers.
	for k, v := range r.Headers {
		httpRequest.Header.Set(k, v)
	}

	// Add any queries.
	if len(r.Queries) > 0 {
		queries := url.Values{}

		for k, v := range r.Queries {
			queries.Add(k, v)
		}

		httpRequest.URL.RawQuery = queries.Encode()
	}

	// Construct the client.
	var client *http.Client
	if r.Client != nil {
		client = r.Client
	} else {
		client = &http.Client{
			Timeout: DefaultTimeout,
		}
	}

	// Make the request.
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return err
	}

	defer httpResponse.Body.Close()

	r.Code = httpResponse.StatusCode

	// Read the response.
	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}

	// Decode the response body.
	if r.Decoder != nil {
		if err := r.Decoder(r.Code, body); err != nil {
			return err
		}
	}

	return nil
}
