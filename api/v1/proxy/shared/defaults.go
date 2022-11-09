package shared

import api "github.com/styrainc/styra-run-sdk-go/api/v1"

func NewOnModifyInput() OnModifyInput {
	return func(session *api.Session, input interface{}) interface{} {
		if input != nil {
			return input
		}

		return map[string]string{
			"tenant":  session.Tenant,
			"subject": session.Subject,
		}
	}
}
