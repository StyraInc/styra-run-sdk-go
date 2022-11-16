package shared

import (
	"net/http"

	"github.com/styrainc/styra-run-sdk-go/types"
)

func DefaultOnModifyInput(getSession types.GetSession) OnModifyInput {
	return func(r *http.Request, path string, input interface{}) (interface{}, error) {
		if input == nil {
			input = make(map[string]interface{})
		}

		session, err := getSession(r)
		if err != nil {
			return nil, err
		}

		if values, ok := input.(map[string]interface{}); ok {
			values["tenant"] = session.Tenant
			values["subject"] = session.Subject
		}

		return input, nil
	}
}
