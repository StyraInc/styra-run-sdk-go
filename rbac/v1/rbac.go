package v1

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/internal/errors"
	"github.com/styrainc/styra-run-sdk-go/types"
)

const (
	authzPath              = "rbac/manage/allow"
	getRolesPath           = "rbac/roles"
	listUserBindingsFormat = "rbac/user_bindings/%s"
	userBindingFormat      = "rbac/user_bindings/%s/%s"
)

var (
	authzError = errors.NewAuthzError()
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
	GetRoles(ctx context.Context, session *types.Session) ([]string, error)
	ListUserBindingsAll(ctx context.Context, session *types.Session) ([]*UserBinding, error)
	ListUserBindings(ctx context.Context, session *types.Session, users []*User) ([]*UserBinding, error)
	GetUserBinding(ctx context.Context, session *types.Session, user *User) (*UserBinding, error)
	PutUserBinding(ctx context.Context, session *types.Session, user *User, binding *UserBinding) error
	DeleteUserBinding(ctx context.Context, session *types.Session, user *User) error
}

type rbac struct {
	settings *Settings
}

func New(settings *Settings) Rbac {
	return &rbac{
		settings: settings,
	}
}

func (r *rbac) GetRoles(ctx context.Context, session *types.Session) ([]string, error) {
	if !r.authz(ctx, session) {
		return nil, authzError
	}

	result := make([]string, 0)
	if err := r.settings.Client.Query(ctx, getRolesPath, nil, &result); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (r *rbac) ListUserBindingsAll(ctx context.Context, session *types.Session) ([]*UserBinding, error) {
	if !r.authz(ctx, session) {
		return nil, authzError
	}

	data := make(map[string][]string)
	url := fmt.Sprintf(listUserBindingsFormat, session.Tenant)
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

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Id < result[j].Id
	})

	return result, nil
}

func (r *rbac) ListUserBindings(ctx context.Context, session *types.Session, users []*User) ([]*UserBinding, error) {
	if !r.authz(ctx, session) {
		return nil, authzError
	}

	result := make([]*UserBinding, 0)
	for _, user := range users {
		data := make([]string, 0)
		url := fmt.Sprintf(userBindingFormat, session.Tenant, user.Id)

		// To make this even remotely usable for customers at the moment, we are making
		// one request per user. Since most folks will use it through the proxy, which is
		// paginated, this will avoid querying for all users. We should switch to the new
		// pagination stuff in the data plane as soon as possible though. Here we check for
		// 404's and silently ignore them.
		if err := r.settings.Client.GetData(ctx, url, &data); err != nil {
			if httpError, ok := err.(errors.HttpError); ok && httpError.Code() == http.StatusNotFound {
				data = make([]string, 0)
			} else {
				return nil, err
			}
		}

		result = append(result, &UserBinding{
			Id:    user.Id,
			Roles: data,
		})
	}

	return result, nil
}

func (r *rbac) GetUserBinding(ctx context.Context, session *types.Session, user *User) (*UserBinding, error) {
	if !r.authz(ctx, session) {
		return nil, authzError
	}

	data := make([]string, 0)
	url := fmt.Sprintf(userBindingFormat, session.Tenant, user.Id)
	if err := r.settings.Client.GetData(ctx, url, &data); err != nil {
		return nil, err
	}

	return &UserBinding{
		Id:    user.Id,
		Roles: data,
	}, nil
}

func (r *rbac) PutUserBinding(ctx context.Context, session *types.Session, user *User, binding *UserBinding) error {
	if !r.authz(ctx, session) {
		return authzError
	}

	url := fmt.Sprintf(userBindingFormat, session.Tenant, user.Id)
	return r.settings.Client.PutData(ctx, url, binding.Roles)
}

func (r *rbac) DeleteUserBinding(ctx context.Context, session *types.Session, user *User) error {
	if !r.authz(ctx, session) {
		return authzError
	}

	url := fmt.Sprintf(userBindingFormat, session.Tenant, user.Id)
	return r.settings.Client.DeleteData(ctx, url)
}

func (r *rbac) authz(ctx context.Context, session *types.Session) bool {
	if result, err := r.settings.Client.Check(ctx, authzPath, session); err != nil {
		return false
	} else {
		return result
	}
}
