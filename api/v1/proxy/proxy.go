package proxy

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
)

const (
	proxyBatchQueryPath = "/"
)

type Route struct {
	Path    string
	Method  string
	Handler http.HandlerFunc
}

type OnModifyBatchQueryInput func(authz *api.Authz, input interface{}) interface{}

type Callbacks struct {
	GetAuthz                api.GetAuthz
	OnModifyBatchQueryInput OnModifyBatchQueryInput
}

type Settings struct {
	Client    api.Client
	Callbacks *Callbacks
}

type Proxy interface {
	BatchQuery() *Route
}

type proxy struct {
	settings *Settings
}

func New(settings *Settings) Proxy {
	return &proxy{
		settings: settings,
	}
}

func (p *proxy) BatchQuery() *Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodPost) {
			return
		}

		if !utils.HasContentType(w, r, "application/json") {
			return
		}

		request := &BatchProxyRequest{}

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

		// Allow the user to modify inputs.
		if p.settings.Callbacks.OnModifyBatchQueryInput != nil {
			authz, err := p.settings.Callbacks.GetAuthz(r)
			if err != nil {
				p.authzError(w, err)
				return
			}

			request.Input = p.settings.Callbacks.OnModifyBatchQueryInput(authz, request.Input)

			for _, query := range queries {
				query.Input = p.settings.Callbacks.OnModifyBatchQueryInput(authz, query.Input)
			}
		}

		// Make the request. If an error occurs, and if it's a http error, forward
		// the payload on from the backend with the appropriate status code.
		if err := p.settings.Client.BatchQuery(r.Context(), queries, request.Input); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &BatchProxyResponse{
			Result: make([]*BatchProxyResult, 0),
		}

		for _, query := range queries {
			response.Result = append(
				response.Result,
				&BatchProxyResult{
					Result: query.Result,
				},
			)
		}

		utils.WriteResponse(w, response)
	}

	return &Route{
		Path:    proxyBatchQueryPath,
		Method:  http.MethodPost,
		Handler: handler,
	}
}

func (p *proxy) authzError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}
