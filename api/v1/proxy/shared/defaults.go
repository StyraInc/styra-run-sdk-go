package shared

import api "github.com/styrainc/styra-run-sdk-go/api/v1"

func NewOnModifyInput() OnModifyInput {
	return func(session *api.Session, path string, input interface{}) interface{} {
		if input == nil {
			input = make(map[string]interface{})
		}

		if values, ok := input.(map[string]interface{}); ok {
			values["tenant"] = session.Tenant
			values["subject"] = session.Subject
		}

		return input
	}
}
