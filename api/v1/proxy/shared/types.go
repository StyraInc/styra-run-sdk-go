package shared

import (
	api "github.com/styrainc/styra-run-sdk-go/types"
)

type OnModifyInput func(session *api.Session, path string, input interface{}) (interface{}, error)
