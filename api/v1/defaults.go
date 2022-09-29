package v1

type DefaultSettings struct {
	Token string
	Url   string
}

func Default(settings *DefaultSettings) Client {
	return New(
		&Settings{
			Token:             settings.Token,
			Url:               settings.Url,
			DiscoveryStrategy: Simple,
			MaxRetries:        3,
		},
	)
}
