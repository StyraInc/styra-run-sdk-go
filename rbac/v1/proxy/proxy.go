package proxy

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	v1 "github.com/styrainc/styra-run-sdk-go/rbac/v1"
)

const (
	proxyGetRolesPath         = "/roles"
	proxyListUserBindingsPath = "/user_bindings"
	proxyUserBindingsFormat   = "/user_bindings/{id}"
)

type Route struct {
	Path    string
	Method  string
	Handler http.HandlerFunc
}

type GetUrlVar func(r *http.Request, key string) string

type GetUsers func(r *http.Request, bytes []byte) ([]*v1.User, interface{}, error)

type OnGetUserBinding func(user *v1.User) (int, error)

type OnPutUserBinding func(user *v1.User) (int, error)

type OnDeleteUserBinding func(user *v1.User) (int, error)

type Callbacks struct {
	GetAuthz            api.GetAuthz
	GetUsers            GetUsers
	OnGetUserBinding    OnGetUserBinding    // optional
	OnPutUserBinding    OnPutUserBinding    // optional
	OnDeleteUserBinding OnDeleteUserBinding // optional
}

type Settings struct {
	Client    api.Client
	GetUrlVar GetUrlVar
	Callbacks *Callbacks
}

type Proxy interface {
	GetRoles() *Route
	ListUserBindings() *Route
	GetUserBinding() *Route
	PutUserBinding() *Route
	DeleteUserBinding() *Route
}

type proxy struct {
	settings *Settings
	rbac     v1.Rbac
}

func New(settings *Settings) Proxy {
	return &proxy{
		settings: settings,
		rbac: v1.New(
			&v1.Settings{
				Client: settings.Client,
			},
		),
	}
}

func (p *proxy) GetRoles() *Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodGet) {
			return
		}

		authz, err := p.settings.Callbacks.GetAuthz(r)
		if err != nil {
			p.authzError(w, err)
			return
		}

		roles, err := p.rbac.GetRoles(r.Context(), authz)
		if err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &GetRolesResponse{
			Result: roles,
		}

		utils.WriteResponse(w, response)
	}

	return &Route{
		Path:    proxyGetRolesPath,
		Method:  http.MethodGet,
		Handler: handler,
	}
}

func (p *proxy) ListUserBindings() *Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodGet) {
			return
		}

		query, ok := utils.HasSingleQueryParameter(w, r, "page")
		if !ok {
			return
		}

		users, page, err := p.settings.Callbacks.GetUsers(r, []byte(query))
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		authz, err := p.settings.Callbacks.GetAuthz(r)
		if err != nil {
			p.authzError(w, err)
			return
		}

		bindings, err := p.rbac.ListUserBindings(r.Context(), authz, users)
		if err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &ListUserBindingsResponse{
			Result: make([]*ListUserBinding, len(users)),
			Page:   page,
		}

		for i, binding := range bindings {
			value := &ListUserBinding{
				Id: users[i].Id,
			}

			if binding != nil {
				value.Roles = binding.Roles
			}

			response.Result[i] = value
		}

		utils.WriteResponse(w, response)
	}

	return &Route{
		Path:    proxyListUserBindingsPath,
		Method:  http.MethodGet,
		Handler: handler,
	}
}

func (p *proxy) GetUserBinding() *Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodGet) {
			return
		}

		authz, err := p.settings.Callbacks.GetAuthz(r)
		if err != nil {
			p.authzError(w, err)
			return
		}

		user := &v1.User{
			Id: p.settings.GetUrlVar(r, "id"),
		}

		if p.settings.Callbacks.OnGetUserBinding != nil {
			if code, err := p.settings.Callbacks.OnGetUserBinding(user); err != nil {
				http.Error(w, err.Error(), code)
				return
			}
		}

		binding, err := p.rbac.GetUserBinding(r.Context(), authz, user)
		if utils.IsHttpError(err, http.StatusNotFound) {
			binding = &v1.UserBinding{
				Roles: make([]string, 0),
			}
		} else if err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &GetUserBindingResponse{
			Result: binding.Roles,
		}

		utils.WriteResponse(w, response)
	}

	return &Route{
		Path:    proxyUserBindingsFormat,
		Method:  http.MethodGet,
		Handler: handler,
	}
}

func (p *proxy) PutUserBinding() *Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodPut) {
			return
		}

		if !utils.HasContentType(w, r, "application/json") {
			return
		}

		roles := make(PutUserBindingRequest, 0)
		if !utils.ReadRequest(w, r, &roles) {
			return
		}

		authz, err := p.settings.Callbacks.GetAuthz(r)
		if err != nil {
			p.authzError(w, err)
			return
		}

		user := &v1.User{
			Id: p.settings.GetUrlVar(r, "id"),
		}
		binding := &v1.UserBinding{
			Roles: roles,
		}

		if p.settings.Callbacks.OnPutUserBinding != nil {
			if code, err := p.settings.Callbacks.OnPutUserBinding(user); err != nil {
				http.Error(w, err.Error(), code)
				return
			}
		}

		if err := p.rbac.PutUserBinding(r.Context(), authz, user, binding); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &PutUserBindingResponse{}
		utils.WriteResponse(w, response)
	}

	return &Route{
		Path:    proxyUserBindingsFormat,
		Method:  http.MethodPut,
		Handler: handler,
	}
}

func (p *proxy) DeleteUserBinding() *Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodDelete) {
			return
		}

		authz, err := p.settings.Callbacks.GetAuthz(r)
		if err != nil {
			p.authzError(w, err)
			return
		}

		user := &v1.User{
			Id: p.settings.GetUrlVar(r, "id"),
		}

		if p.settings.Callbacks.OnDeleteUserBinding != nil {
			if code, err := p.settings.Callbacks.OnDeleteUserBinding(user); err != nil {
				http.Error(w, err.Error(), code)
				return
			}
		}

		if err := p.rbac.DeleteUserBinding(r.Context(), authz, user); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &DeleteUserBindingResponse{}
		utils.WriteResponse(w, response)
	}

	return &Route{
		Path:    proxyUserBindingsFormat,
		Method:  http.MethodDelete,
		Handler: handler,
	}
}

func (p *proxy) authzError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}
