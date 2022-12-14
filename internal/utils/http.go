package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/styrainc/styra-run-sdk-go/internal/errors"
)

const (
	ApplicationJson = "application/json"
)

func JoinPath(base string, paths ...string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	for _, p := range paths {
		u.Path = path.Join(u.Path, p)
	}

	return u.String(), nil
}

func InternalServerError(w http.ResponseWriter) {
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func AuthzError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}

func HasMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}

	return true
}

func HasContentType(w http.ResponseWriter, r *http.Request, contentType string) bool {
	if headers, ok := r.Header["Content-Type"]; !ok {
		http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
		return false
	} else {
		values := make(map[string]bool)
		for _, v := range headers {
			values[v] = true
		}

		if _, ok := values[contentType]; !ok {
			http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
			return false
		}
	}

	return true
}

func HasSingleQueryParameter(w http.ResponseWriter, r *http.Request, name string) (string, bool) {
	if params, ok := r.URL.Query()[name]; !ok {
		message := fmt.Sprintf("missing query parameter: %s", name)
		http.Error(w, message, http.StatusBadRequest)
		return "", false
	} else if len(params) > 1 {
		message := fmt.Sprintf("query parameter %s should have exactly one value", name)
		http.Error(w, message, http.StatusBadRequest)
		return "", false
	} else {
		return params[0], true
	}
}

func ReadRequest(w http.ResponseWriter, r *http.Request, request interface{}) bool {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return false
	}

	if err := json.Unmarshal(body, request); err != nil {
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return false
	}

	return true
}

func WriteResponse(w http.ResponseWriter, response interface{}) bool {
	if bytes, err := json.Marshal(response); err != nil {
		InternalServerError(w)
		return false
	} else {
		w.Header().Set("Content-Type", ApplicationJson)

		if _, err := w.Write(bytes); err != nil {
			InternalServerError(w)
			return false
		}
	}

	return true
}

func ForwardHttpError(w http.ResponseWriter, err error) {
	if httpError, ok := err.(errors.HttpError); ok && httpError.Details() != nil {
		if bytes, err := json.Marshal(httpError.Details()); err != nil {
			InternalServerError(w)
		} else {
			w.Header().Set("Content-Type", ApplicationJson)
			w.WriteHeader(httpError.Code())

			if _, err := w.Write(bytes); err != nil {
				InternalServerError(w)
			}
		}
	} else if _, ok := err.(errors.AuthzError); ok {
		http.Error(w, "forbidden", http.StatusForbidden)
	} else {
		InternalServerError(w)
	}
}
