package shared

import "net/http"

type OnModifyInput func(r *http.Request, path string, input interface{}) (interface{}, error)
