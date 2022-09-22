package proxy

import (
	"encoding/json"
	"errors"
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
)

type DefaultCallbackSettings struct {
	GetAuthz api.GetAuthz
}

func DefaultCallbacks(settings *DefaultCallbackSettings) *Callbacks {
	return &Callbacks{
		GetAuthz: settings.GetAuthz,
	}
}

type ArrayCallbackSettings struct {
	GetAuthz api.GetAuthz
	Users    []*rbac.User
	PageSize int
}

func ArrayCallbacks(settings *ArrayCallbackSettings) *Callbacks {
	return &Callbacks{
		GetAuthz:            settings.GetAuthz,
		GetUsers:            NewGetUsers(settings.Users, settings.PageSize),
		OnGetUserBinding:    NewUserExists(settings.Users),
		OnPutUserBinding:    NewUserExists(settings.Users),
		OnDeleteUserBinding: NewUserExists(settings.Users),
	}
}

func NewGetUsers(users []*rbac.User, size int) GetUsers {
	return func(r *http.Request, bytes []byte) ([]*rbac.User, interface{}, error) {
		var page int
		if err := json.Unmarshal(bytes, &page); err != nil {
			return nil, nil, err
		}

		if page < 0 {
			page = 0
		}

		if size < 0 {
			size = 0
		}

		values := make([]*rbac.User, 0)
		for i, end := page*size, page*size+size; i < end; i++ {
			if i < len(users) {
				values = append(values, users[i])
			}
		}

		info := &struct {
			Index int `json:"index"`
			Total int `json:"total"`
		}{
			Index: page,
			Total: len(users) / size,
		}

		return values, info, nil
	}
}

func NewUserExists(users []*rbac.User) func(user *rbac.User) (int, error) {
	return func(user *rbac.User) (int, error) {
		for _, u := range users {
			if user.Id == u.Id {
				return http.StatusOK, nil
			}
		}

		return http.StatusBadRequest, errors.New("user does not exist")
	}
}
