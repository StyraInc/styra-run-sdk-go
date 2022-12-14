package list_user_bindings

import (
	"net/http"

	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
	"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/shared"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type ListUserBindingsResponse struct {
	Result []*rbac.UserBinding `json:"result"`
	Page   interface{}         `json:"page,omitempty"`
}

type Settings struct {
	// The SDK rbac instance.
	Rbac rbac.Rbac

	// A callback to get session information.
	GetSession types.GetSession

	// A callback that, given an HTTP request and `page` query parameter
	// details, emits a list of users and corresponding page information.
	GetUsers shared.GetUsers
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

		query, ok := utils.HasSingleQueryParameter(w, r, "page")
		if !ok {
			return
		}

		users, page, err := settings.GetUsers(r, []byte(query))
		if err != nil {
			utils.InternalServerError(w)
			return
		}

		bindings, err := settings.Rbac.ListUserBindings(r.Context(), session, users)
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

	return &types.Proxy{
		Method:  http.MethodGet,
		Handler: handler,
	}
}
