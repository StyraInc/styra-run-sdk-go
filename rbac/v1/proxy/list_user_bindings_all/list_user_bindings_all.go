package list_user_bindings_all

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type ListUserBindingsAllResponse struct {
	Result []*rbac.UserBinding `json:"result"`
	Page   interface{}         `json:"page,omitempty"`
}

type Settings struct {
	Rbac       rbac.Rbac
	GetSession api.GetSession
}

func New(settings *Settings) *types.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodGet) {
			return
		}

		session, err := settings.GetSession(r)
		if err != nil {
			utils.AuthzError(w, err)
			return
		}

		bindings, err := settings.Rbac.ListUserBindingsAll(r.Context(), session)
		if err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &ListUserBindingsAllResponse{
			Result: bindings,
		}

		utils.WriteResponse(w, response)
	}

	return &types.Route{
		Method:  http.MethodGet,
		Handler: handler,
	}
}
