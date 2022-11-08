package proxy

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type GetUsers func(r *http.Request, bytes []byte) ([]*rbac.User, interface{}, error)

type OnGetUserBinding func(user *rbac.User) (int, error)

type OnPutUserBinding func(user *rbac.User) (int, error)

type OnDeleteUserBinding func(user *rbac.User) (int, error)

type Callbacks struct {
	GetSession          api.GetSession
	GetUsers            GetUsers
	OnGetUserBinding    OnGetUserBinding    // optional
	OnPutUserBinding    OnPutUserBinding    // optional
	OnDeleteUserBinding OnDeleteUserBinding // optional
}

type Settings struct {
	Client    api.Client
	Callbacks *Callbacks
}

type Proxy interface {
	GetRoles() *types.Route
	ListAllUserBindings() *types.Route
	ListUserBindings() *types.Route
	GetUserBinding(getId types.GetVar) *types.Route
	PutUserBinding(getId types.GetVar) *types.Route
	DeleteUserBinding(getId types.GetVar) *types.Route
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

func (p *proxy) GetRoles() *types.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodGet) {
			return
		}

		session, err := p.settings.Callbacks.GetSession(r)
		if err != nil {
			p.authzError(w, err)
			return
		}

		roles, err := p.rbac.GetRoles(r.Context(), session)
		if err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &GetRolesResponse{
			Result: roles,
		}

		utils.WriteResponse(w, response)
	}

	return &types.Route{
		Method:  http.MethodGet,
		Handler: handler,
	}
}

func (p *proxy) ListAllUserBindings() *types.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodGet) {
			return
		}

		session, err := p.settings.Callbacks.GetSession(r)
		if err != nil {
			p.authzError(w, err)
			return
		}

		bindings, err := p.rbac.ListUserBindingsAll(r.Context(), session)
		if err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &ListUserBindingsResponse{
			Result: bindings,
		}

		utils.WriteResponse(w, response)
	}

	return &types.Route{
		Method:  http.MethodGet,
		Handler: handler,
	}
}

func (p *proxy) ListUserBindings() *types.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodGet) {
			return
		}

		session, err := p.settings.Callbacks.GetSession(r)
		if err != nil {
			p.authzError(w, err)
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

		bindings, err := p.rbac.ListUserBindings(r.Context(), session, users)
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

	return &types.Route{
		Method:  http.MethodGet,
		Handler: handler,
	}
}

func (p *proxy) GetUserBinding(getId types.GetVar) *types.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodGet) {
			return
		}

		session, err := p.settings.Callbacks.GetSession(r)
		if err != nil {
			p.authzError(w, err)
			return
		}

		user := &rbac.User{
			Id: getId(r),
		}

		if p.settings.Callbacks.OnGetUserBinding != nil {
			if code, err := p.settings.Callbacks.OnGetUserBinding(user); err != nil {
				http.Error(w, err.Error(), code)
				return
			}
		}

		binding, err := p.rbac.GetUserBinding(r.Context(), session, user)
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

	return &types.Route{
		Method:  http.MethodGet,
		Handler: handler,
	}
}

func (p *proxy) PutUserBinding(getId types.GetVar) *types.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodPut) {
			return
		}

		if !utils.HasContentType(w, r, utils.ApplicationJson) {
			return
		}

		roles := make(PutUserBindingRequest, 0)
		if !utils.ReadRequest(w, r, &roles) {
			return
		}

		session, err := p.settings.Callbacks.GetSession(r)
		if err != nil {
			p.authzError(w, err)
			return
		}

		user := &rbac.User{
			Id: getId(r),
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

		if err := p.rbac.PutUserBinding(r.Context(), session, user, binding); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &PutUserBindingResponse{}
		utils.WriteResponse(w, response)
	}

	return &types.Route{
		Method:  http.MethodPut,
		Handler: handler,
	}
}

func (p *proxy) DeleteUserBinding(getId types.GetVar) *types.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodDelete) {
			return
		}

		session, err := p.settings.Callbacks.GetSession(r)
		if err != nil {
			p.authzError(w, err)
			return
		}

		user := &rbac.User{
			Id: getId(r),
		}

		if p.settings.Callbacks.OnDeleteUserBinding != nil {
			if code, err := p.settings.Callbacks.OnDeleteUserBinding(user); err != nil {
				http.Error(w, err.Error(), code)
				return
			}
		}

		if err := p.rbac.DeleteUserBinding(r.Context(), session, user); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &DeleteUserBindingResponse{}
		utils.WriteResponse(w, response)
	}

	return &types.Route{
		Method:  http.MethodDelete,
		Handler: handler,
	}
}

func (p *proxy) authzError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}
