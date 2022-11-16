package shared

import (
	api "github.com/styrainc/styra-run-sdk-go/types"
)

// Default callback that adds session information to an input section.
func DefaultOnModifyInput() OnModifyInput {
	return func(session *api.Session, path string, input interface{}) (interface{}, error) {
		if input == nil {
			input = make(map[string]interface{})
		}

		if values, ok := input.(map[string]interface{}); ok {
			values["tenant"] = session.Tenant
			values["subject"] = session.Subject
		}

		return input, nil
	}
}
