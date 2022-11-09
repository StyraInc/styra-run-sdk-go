package query

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/shared"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type QueryRequest struct {
	Input interface{} `json:"input"`
}

type QueryResponse struct {
	Result interface{} `json:"result"`
}

type Settings struct {
	Client        api.Client
	GetPath       types.GetVar
	GetSession    api.GetSession
	OnModifyInput shared.OnModifyInput
}

func New(settings *Settings) *types.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodPost) {
			return
		}

		if !utils.HasContentType(w, r, utils.ApplicationJson) {
			return
		}

		request := &QueryRequest{}
		if !utils.ReadRequest(w, r, &request) {
			return
		}

		path := settings.GetPath(r)

		// Allow the user to modify inputs if the callback is set.
		if settings.GetSession != nil && settings.OnModifyInput != nil {
			session, err := settings.GetSession(r)
			if err != nil {
				utils.AuthzError(w, err)
				return
			}

			request.Input = settings.OnModifyInput(session, path, request.Input)
		}

		var data interface{}
		if err := settings.Client.Query(r.Context(), path, request.Input, &data); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &QueryResponse{
			Result: data,
		}

		utils.WriteResponse(w, response)
	}

	return &types.Route{
		Method:  http.MethodPost,
		Handler: handler,
	}
}