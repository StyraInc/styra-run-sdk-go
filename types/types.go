package types

import "net/http"

type Proxy struct {
	Method  string
	Handler http.HandlerFunc
}

type GetVar func(r *http.Request) string
