package check

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/shared"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type CheckRequest struct {
	Input interface{} `json:"input"`
}

type CheckResponse struct {
	Result bool `json:"result"`
}

type Settings struct {
	// The SDK client.
	Client api.Client

	// A callback to get the query path.
	GetPath types.GetVar

	// Two optional callbacks that must be specified together:
	// GetSession:    A callback to get session information.
	// OnModifyInput: A callback to modify query inputs.
	GetSession    types.GetSession
	OnModifyInput shared.OnModifyInput
}

func New(settings *Settings) *types.Proxy {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodPost) {
			return
		}

		if !utils.HasContentType(w, r, utils.ApplicationJson) {
			return
		}

		request := &CheckRequest{}
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

			if input, err := settings.OnModifyInput(session, path, request.Input); err != nil {
				utils.InternalServerError(w)
				return
			} else {
				request.Input = input
			}
		}

		result, err := settings.Client.Check(r.Context(), path, request.Input)
		if err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &CheckResponse{
			Result: result,
		}

		utils.WriteResponse(w, response)
	}

	return &types.Proxy{
		Method:  http.MethodPost,
		Handler: handler,
	}
}
