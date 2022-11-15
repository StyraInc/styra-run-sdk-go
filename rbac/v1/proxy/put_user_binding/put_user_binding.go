package put_user_binding

import (
	"net/http"

	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
	"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/shared"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type PutUserBindingRequest []string

type PutUserBindingResponse struct{}

type Settings struct {
	Rbac           rbac.Rbac
	GetSession     types.GetSession
	GetId          types.GetVar
	OnBeforeAccess shared.OnBeforeAccess
}

func New(settings *Settings) *types.Proxy {
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

		session, err := settings.GetSession(r)
		if err != nil {
			utils.AuthzError(w, err)
			return
		}

		user := &rbac.User{
			Id: settings.GetId(r),
		}

		binding := &rbac.UserBinding{
			Roles: roles,
		}

		if settings.OnBeforeAccess != nil {
			if code, err := settings.OnBeforeAccess(user); err != nil {
				http.Error(w, err.Error(), code)
				return
			}
		}

		if err := settings.Rbac.PutUserBinding(r.Context(), session, user, binding); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &PutUserBindingResponse{}
		utils.WriteResponse(w, response)
	}

	return &types.Proxy{
		Method:  http.MethodPut,
		Handler: handler,
	}
}
