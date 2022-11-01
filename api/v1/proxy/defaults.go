package proxy

import (
	api "github.com/styrainc/styra-run-sdk-go/api/v1"
)

type DefaultSettings struct {
	Client     api.Client
	GetSession api.GetSession
}

func Default(settings *DefaultSettings) Proxy {
	return New(
		&Settings{
			Client: settings.Client,
			Callbacks: DefaultCallbacks(
				&DefaultCallbackSettings{
					GetSession: settings.GetSession,
				},
			),
		},
	)
}

type DefaultCallbackSettings struct {
	GetSession api.GetSession
}

func DefaultCallbacks(settings *DefaultCallbackSettings) *Callbacks {
	return &Callbacks{
		GetSession:              settings.GetSession,
		OnModifyBatchQueryInput: NewOnModifyBatchQueryInput(),
	}
}

func NewOnModifyBatchQueryInput() OnModifyBatchQueryInput {
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
