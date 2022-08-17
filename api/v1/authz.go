package v1

import (
	"errors"
	"net/http"
	"strings"
)

var (
	credentialsError = errors.New("could not extract credentials")
)

type Authz struct {
	Tenant  string `json:"tenant"`
	Subject string `json:"subject"`
}

type GetAuthz func(r *http.Request) (*Authz, error)

func AuthzFromValues(tenant, subject string) GetAuthz {
	return func(r *http.Request) (*Authz, error) {
		return &Authz{
			Tenant:  tenant,
			Subject: subject,
		}, nil
	}
}

func AuthzFromCookie() GetAuthz {
	return func(r *http.Request) (*Authz, error) {
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

		return &Authz{
			Tenant:  tenant,
			Subject: subject,
		}, nil
	}
}

func AuthzFromContext() GetAuthz {
	return func(r *http.Request) (*Authz, error) {
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

		return &Authz{
			Tenant:  tenant,
			Subject: subject,
		}, nil
	}
}
