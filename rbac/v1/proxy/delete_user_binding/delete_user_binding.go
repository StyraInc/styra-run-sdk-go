package delete_user_binding

import (
	"net/http"

	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
	"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/shared"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type DeleteUserBindingResponse struct{}

type Settings struct {
	// The SDK rbac instance.
	Rbac rbac.Rbac

	// A callback to get session information.
	GetSession types.GetSession

	// A callback to get the user id.
	GetId types.GetVar

	// An optional callback called before user bindings are accessed.
	OnBeforeAccess shared.OnBeforeAccess
}

func New(settings *Settings) *types.Proxy {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodDelete) {
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

		if settings.OnBeforeAccess != nil {
			if code, err := settings.OnBeforeAccess(user); err != nil {
				http.Error(w, err.Error(), code)
				return
			}
		}

		if err := settings.Rbac.DeleteUserBinding(r.Context(), session, user); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &DeleteUserBindingResponse{}
		utils.WriteResponse(w, response)
	}

	return &types.Proxy{
		Method:  http.MethodDelete,
		Handler: handler,
	}
}
