package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return false
	} else {
		w.Header().Set("Content-Type", "application/json")

		if _, err := w.Write(bytes); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return false
		}
	}

	return true
}

func ForwardHttpError(w http.ResponseWriter, err error) {
	if httpError, ok := err.(HttpError); ok && httpError.Details() != nil {
		if bytes, err := json.Marshal(httpError.Details()); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(httpError.Code())

			if _, err := w.Write(bytes); err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
	} else if _, ok := err.(AuthzError); ok {
		http.Error(w, "forbidden", http.StatusForbidden)
	} else {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
