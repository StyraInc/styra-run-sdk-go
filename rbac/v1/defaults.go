package v1

import (
	api "github.com/styrainc/styra-run-sdk-go/api/v1"
)

type DefaultSettings struct {
	Client api.Client
}

func Default(settings *DefaultSettings) Rbac {
	return New(
		&Settings{
			Client: settings.Client,
		},
	)
}
