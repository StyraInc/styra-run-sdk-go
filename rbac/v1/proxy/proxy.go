package proxy

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
)

type RouteType uint

const (
	proxyGetRolesPath         = "/roles"
	proxyListUserBindingsPath = "/user_bindings"
	proxyUserBindingsFormat   = "/user_bindings/{id}"

	// The route types.
	GetRoles RouteType = iota
	ListUserBindings
	GetUserBinding
	PutUserBinding
	DeleteUserBinding
)

type Route struct {
	Path    string
	Method  string
	Handler http.HandlerFunc
}

type GetUrlVar func(r *http.Request, key string) string

type GetUsers func(r *http.Request, bytes []byte) ([]*rbac.User, interface{}, error)

type OnGetUserBinding func(user *rbac.User) (int, error)

type OnPutUserBinding func(user *rbac.User) (int, error)

type OnDeleteUserBinding func(user *rbac.User) (int, error)

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
	All() map[RouteType]*Route
}

type proxy struct {
	settings *Settings
	rbac     rbac.Rbac
}

func New(settings *Settings) Proxy {
	return &proxy{
		settings: settings,
		rbac: rbac.New(
			&rbac.Settings{
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

		authz, err := p.settings.Callbacks.GetAuthz(r)
		if err != nil {
			p.authzError(w, err)
			return
		}

		if p.settings.Callbacks.GetUsers == nil {
			p.listUserBindingsAll(w, r, authz)
		} else {
			p.listUserBindings(w, r, authz)
		}
	}

	return &Route{
		Path:    proxyListUserBindingsPath,
		Method:  http.MethodGet,
		Handler: handler,
	}
}

func (p *proxy) listUserBindingsAll(w http.ResponseWriter, r *http.Request, authz *api.Authz) {
	bindings, err := p.rbac.ListUserBindingsAll(r.Context(), authz)
	if err != nil {
		utils.ForwardHttpError(w, err)
		return
	}

	response := &ListUserBindingsResponse{
		Result: bindings,
	}

	utils.WriteResponse(w, response)
}

func (p *proxy) listUserBindings(w http.ResponseWriter, r *http.Request, authz *api.Authz) {
	query, ok := utils.HasSingleQueryParameter(w, r, "page")
	if !ok {
		return
	}

	users, page, err := p.settings.Callbacks.GetUsers(r, []byte(query))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	bindings, err := p.rbac.ListUserBindings(r.Context(), authz, users)
	if err != nil {
		utils.ForwardHttpError(w, err)
		return
	}

	response := &ListUserBindingsResponse{
		Result: bindings,
		Page:   page,
	}

	utils.WriteResponse(w, response)
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

		user := &rbac.User{
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
			binding = &rbac.UserBinding{
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

		user := &rbac.User{
			Id: p.settings.GetUrlVar(r, "id"),
		}
		binding := &rbac.UserBinding{
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

		user := &rbac.User{
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

func (p *proxy) All() map[RouteType]*Route {
	return map[RouteType]*Route{
		GetRoles:          p.GetRoles(),
		ListUserBindings:  p.ListUserBindings(),
		GetUserBinding:    p.GetUserBinding(),
		PutUserBinding:    p.PutUserBinding(),
		DeleteUserBinding: p.DeleteUserBinding(),
	}
}

func (p *proxy) authzError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}
