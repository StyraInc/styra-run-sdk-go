package shared

import api "github.com/styrainc/styra-run-sdk-go/api/v1"

type OnModifyInput func(session *api.Session, path string, input interface{}) interface{}
