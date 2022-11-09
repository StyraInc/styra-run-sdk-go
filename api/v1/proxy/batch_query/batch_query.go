package batch_query

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/shared"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type BatchQueryRequestItem struct {
	Path  string      `json:"path"`
	Input interface{} `json:"input,omitempty"`
}

type BatchQueryRequest struct {
	Items []*BatchQueryRequestItem `json:"items"`
	Input interface{}              `json:"input,omitempty"`
}

type BatchQueryResponseItem struct {
	Result interface{} `json:"result,omitempty"`
}

type BatchQueryResponse struct {
	Result []*BatchQueryResponseItem `json:"result"`
}

type Settings struct {
	Client        api.Client
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

		request := &BatchQueryRequest{}

		if !utils.ReadRequest(w, r, request) {
			return
		}

		queries := make([]api.Query, 0)
		for _, item := range request.Items {
			queries = append(
				queries,
				api.Query{
					Path:  item.Path,
					Input: item.Input,
				},
			)
		}

		// Allow the user to modify inputs if the callback is set.
		if settings.GetSession != nil && settings.OnModifyInput != nil {
			session, err := settings.GetSession(r)
			if err != nil {
				utils.AuthzError(w, err)
				return
			}

			request.Input = settings.OnModifyInput(session, "", request.Input)

			for i := range queries {
				queries[i].Input = settings.OnModifyInput(session, queries[i].Path, queries[i].Input)
			}
		}

		// Make the request. If an error occurs, and if it's a http error, forward
		// the payload on from the backend with the appropriate status code.
		if err := settings.Client.BatchQuery(r.Context(), queries, request.Input); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &BatchQueryResponse{
			Result: make([]*BatchQueryResponseItem, 0),
		}

		for _, query := range queries {
			response.Result = append(
				response.Result,
				&BatchQueryResponseItem{
					Result: query.Result,
				},
			)
		}

		utils.WriteResponse(w, response)
	}

	return &types.Route{
		Method:  http.MethodPost,
		Handler: handler,
	}
}
