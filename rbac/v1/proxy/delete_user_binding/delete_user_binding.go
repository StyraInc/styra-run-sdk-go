package delete_user_binding

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
	"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/shared"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type DeleteUserBindingResponse struct{}

type Settings struct {
	Rbac       rbac.Rbac
	GetSession api.GetSession
	GetId      types.GetVar
	OnTouched  shared.OnUserBindingTouched
}

func New(settings *Settings) *types.Route {
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

		if settings.OnTouched != nil {
			if code, err := settings.OnTouched(user); err != nil {
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

	return &types.Route{
		Method:  http.MethodDelete,
		Handler: handler,
	}
}
