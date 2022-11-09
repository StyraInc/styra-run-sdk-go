package list_user_bindings

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type ListUserBindingsResponse struct {
	Result []*rbac.UserBinding `json:"result"`
	Page   interface{}         `json:"page,omitempty"`
}

type GetUsers func(r *http.Request, bytes []byte) ([]*rbac.User, interface{}, error)

type Settings struct {
	Rbac       rbac.Rbac
	GetSession api.GetSession
	GetUsers   GetUsers
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

		query, ok := utils.HasSingleQueryParameter(w, r, "page")
		if !ok {
			return
		}

		users, page, err := settings.GetUsers(r, []byte(query))
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
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

	return &types.Route{
		Method:  http.MethodGet,
		Handler: handler,
	}
}
