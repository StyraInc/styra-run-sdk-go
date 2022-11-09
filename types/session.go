package types

import (
	"errors"
	"net/http"
	"strings"
)

var (
	credentialsError = errors.New("could not extract credentials")
)

type Session struct {
	Tenant  string `json:"tenant"`
	Subject string `json:"subject"`
}

type GetSession func(r *http.Request) (*Session, error)

func SessionFromValues(tenant, subject string) GetSession {
	return func(r *http.Request) (*Session, error) {
		return &Session{
			Tenant:  tenant,
			Subject: subject,
		}, nil
	}
}

func SessionFromCookie() GetSession {
	return func(r *http.Request) (*Session, error) {
		cookie, err := r.Cookie("user")
		if err != nil {
			return nil, err
		}

		values := strings.Split(cookie.Value, "/")
		if len(values) != 2 {
			return nil, credentialsError
		}

		tenant := strings.TrimSpace(values[0])
		if tenant == "" {
			return nil, credentialsError
		}

		subject := strings.TrimSpace(values[1])
		if subject == "" {
			return nil, credentialsError
		}

		return &Session{
			Tenant:  tenant,
			Subject: subject,
		}, nil
	}
}

func SessionFromContext() GetSession {
	return func(r *http.Request) (*Session, error) {
		var tenant, subject string

		if value, ok := r.Context().Value("tenant").(string); ok && value != "" {
			tenant = value
		} else {
			return nil, credentialsError
		}

		if value, ok := r.Context().Value("subject").(string); ok && value != "" {
			subject = value
		} else {
			return nil, credentialsError
		}

		return &Session{
			Tenant:  tenant,
			Subject: subject,
		}, nil
	}
}
