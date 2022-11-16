package list_user_bindings_all

import (
	"net/http"

	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type ListUserBindingsAllResponse struct {
	Result []*rbac.UserBinding `json:"result"`
	Page   interface{}         `json:"page,omitempty"`
}

type Settings struct {
	// The SDK rbac instance.
	Rbac rbac.Rbac

	// A callback to get session information.
	GetSession types.GetSession
}

func New(settings *Settings) *types.Proxy {
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

	return &types.Proxy{
		Method:  http.MethodGet,
		Handler: handler,
	}
}
