package proxy

import (
	api "github.com/styrainc/styra-run-sdk-go/api/v1"
)

type DefaultSettings struct {
	Client   api.Client
	GetAuthz api.GetAuthz
}

func Default(settings *DefaultSettings) Proxy {
	return New(
		&Settings{
			Client: settings.Client,
			Callbacks: DefaultCallbacks(
				&DefaultCallbackSettings{
					GetAuthz: settings.GetAuthz,
				},
			),
		},
	)
}

type DefaultCallbackSettings struct {
	GetAuthz api.GetAuthz
}

func DefaultCallbacks(settings *DefaultCallbackSettings) *Callbacks {
	return &Callbacks{
		GetAuthz:                settings.GetAuthz,
		OnModifyBatchQueryInput: NewOnModifyBatchQueryInput(),
	}
}

func NewOnModifyBatchQueryInput() OnModifyBatchQueryInput {
	return func(authz *api.Authz, input interface{}) interface{} {
		if input == nil {
			input = make(map[string]interface{})
		}

		if values, ok := input.(map[string]interface{}); ok {
			values["tenant"] = authz.Tenant
			values["subject"] = authz.Subject
		}

		return input
	}
}
