package get_roles

import (
	"net/http"

	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type GetRolesResponse struct {
	Result []string `json:"result"`
}

type Settings struct {
	Rbac       rbac.Rbac
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

		roles, err := settings.Rbac.GetRoles(r.Context(), session)
		if err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &GetRolesResponse{
			Result: roles,
		}

		utils.WriteResponse(w, response)
	}

	return &types.Proxy{
		Method:  http.MethodGet,
		Handler: handler,
	}
}
