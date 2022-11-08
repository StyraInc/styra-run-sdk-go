package types

import "net/http"

type Route struct {
	Method  string
	Handler http.HandlerFunc
}

type GetVar func(r *http.Request) string
