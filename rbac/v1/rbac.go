package v1

import (
	"context"
	"errors"
	"fmt"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
)

const (
	authzPath              = "rbac/manage/allow"
	getRolesPath           = "rbac/roles"
	listUserBindingsFormat = "rbac/user_bindings/%s"
	userBindingFormat      = "rbac/user_bindings/%s/%s"
)

var (
	authzError = errors.New("permission denied")
)

type User struct {
	Id string
}

type UserBinding struct {
	Id    string   `json:"id"`
	Roles []string `json:"roles"`
}

type Settings struct {
	Client api.Client
}

type Rbac interface {
	GetRoles(ctx context.Context, authz *api.Authz) ([]string, error)
	ListUserBindingsAll(ctx context.Context, authz *api.Authz) ([]*UserBinding, error)
	ListUserBindings(ctx context.Context, authz *api.Authz, users []*User) ([]*UserBinding, error)
	GetUserBinding(ctx context.Context, authz *api.Authz, user *User) (*UserBinding, error)
	PutUserBinding(ctx context.Context, authz *api.Authz, user *User, binding *UserBinding) error
	DeleteUserBinding(ctx context.Context, authz *api.Authz, user *User) error
}

type rbac struct {
	settings *Settings
}

func New(settings *Settings) Rbac {
	return &rbac{
		settings: settings,
	}
}

func (r *rbac) GetRoles(ctx context.Context, authz *api.Authz) ([]string, error) {
	if !r.authz(ctx, authz) {
		return nil, authzError
	}

	result := make([]string, 0)
	if err := r.settings.Client.Query(ctx, getRolesPath, nil, &result); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (r *rbac) ListUserBindingsAll(ctx context.Context, authz *api.Authz) ([]*UserBinding, error) {
	if !r.authz(ctx, authz) {
		return nil, authzError
	}

	data := make(map[string][]string)
	url := fmt.Sprintf(listUserBindingsFormat, authz.Tenant)
	if err := r.settings.Client.GetData(ctx, url, &data); err != nil {
		return nil, err
	}

	result := make([]*UserBinding, 0)
	for id, roles := range data {
		result = append(
			result,
			&UserBinding{
				Id:    id,
				Roles: roles,
			},
		)
	}

	return result, nil
}

func (r *rbac) ListUserBindings(ctx context.Context, authz *api.Authz, users []*User) ([]*UserBinding, error) {
	if !r.authz(ctx, authz) {
		return nil, authzError
	}

	data := make(map[string][]string)
	url := fmt.Sprintf(listUserBindingsFormat, authz.Tenant)
	if err := r.settings.Client.GetData(ctx, url, &data); err != nil {
		return nil, err
	}

	result := make([]*UserBinding, 0)
	for _, user := range users {
		roles := make([]string, 0)
		if values, ok := data[user.Id]; ok {
			roles = values
		}

		result = append(
			result,
			&UserBinding{
				Id:    user.Id,
				Roles: roles,
			},
		)
	}

	return result, nil
}

func (r *rbac) GetUserBinding(ctx context.Context, authz *api.Authz, user *User) (*UserBinding, error) {
	if !r.authz(ctx, authz) {
		return nil, authzError
	}

	data := make([]string, 0)
	url := fmt.Sprintf(userBindingFormat, authz.Tenant, user.Id)
	if err := r.settings.Client.GetData(ctx, url, &data); err != nil {
		return nil, err
	}

	return &UserBinding{
		Id:    user.Id,
		Roles: data,
	}, nil
}

func (r *rbac) PutUserBinding(ctx context.Context, authz *api.Authz, user *User, binding *UserBinding) error {
	if !r.authz(ctx, authz) {
		return authzError
	}

	url := fmt.Sprintf(userBindingFormat, authz.Tenant, user.Id)
	return r.settings.Client.PutData(ctx, url, binding.Roles)
}

func (r *rbac) DeleteUserBinding(ctx context.Context, authz *api.Authz, user *User) error {
	if !r.authz(ctx, authz) {
		return authzError
	}

	url := fmt.Sprintf(userBindingFormat, authz.Tenant, user.Id)
	return r.settings.Client.DeleteData(ctx, url)
}

func (r *rbac) authz(ctx context.Context, authz *api.Authz) bool {
	if result, err := r.settings.Client.Check(ctx, authzPath, authz); err != nil {
		return false
	} else {
		return result
	}
}
