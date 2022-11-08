package proxy

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type OnModifyBatchQueryInput func(session *api.Session, input interface{}) interface{}

type Callbacks struct {
	GetSession              api.GetSession
	OnModifyBatchQueryInput OnModifyBatchQueryInput
}

type Settings struct {
	Client    api.Client
	Callbacks *Callbacks
}

type Proxy interface {
	GetData(getPath types.GetVar) *types.Route
	PutData(getPath types.GetVar) *types.Route
	DeleteData(getPath types.GetVar) *types.Route
	Query(getPath types.GetVar) *types.Route
	Check(getPath types.GetVar) *types.Route
	BatchQuery() *types.Route
}

type proxy struct {
	settings *Settings
}

func New(settings *Settings) Proxy {
	return &proxy{
		settings: settings,
	}
}

func (p *proxy) GetData(getPath types.GetVar) *types.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodGet) {
			return
		}

		path := getPath(r)

		var data interface{}
		if err := p.settings.Client.GetData(r.Context(), path, &data); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &GetDataResponse{
			Result: data,
		}

		utils.WriteResponse(w, response)
	}

	return &types.Route{
		Method:  http.MethodGet,
		Handler: handler,
	}
}

func (p *proxy) PutData(getPath types.GetVar) *types.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodPut) {
			return
		}

		if !utils.HasContentType(w, r, utils.ApplicationJson) {
			return
		}

		var data interface{}
		if !utils.ReadRequest(w, r, &data) {
			return
		}

		path := getPath(r)

		if err := p.settings.Client.PutData(r.Context(), path, data); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &PutDataResponse{}
		utils.WriteResponse(w, response)
	}

	return &types.Route{
		Method:  http.MethodPut,
		Handler: handler,
	}
}

func (p *proxy) DeleteData(getPath types.GetVar) *types.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodDelete) {
			return
		}

		path := getPath(r)

		if err := p.settings.Client.DeleteData(r.Context(), path); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &DeleteDataResponse{}
		utils.WriteResponse(w, response)
	}

	return &types.Route{
		Method:  http.MethodDelete,
		Handler: handler,
	}
}

func (p *proxy) Query(getPath types.GetVar) *types.Route {
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

		path := getPath(r)

		var data interface{}
		if err := p.settings.Client.Query(r.Context(), path, request.Input, &data); err != nil {
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

func (p *proxy) Check(getPath types.GetVar) *types.Route {
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

		path := getPath(r)

		result, err := p.settings.Client.Check(r.Context(), path, request.Input)
		if err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &CheckResponse{
			Result: result,
		}

		utils.WriteResponse(w, response)
	}

	return &types.Route{
		Method:  http.MethodPost,
		Handler: handler,
	}
}

func (p *proxy) BatchQuery() *types.Route {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodPost) {
			return
		}

		if !utils.HasContentType(w, r, utils.ApplicationJson) {
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
			session, err := p.settings.Callbacks.GetSession(r)
			if err != nil {
				p.authzError(w, err)
				return
			}

			request.Input = p.settings.Callbacks.OnModifyBatchQueryInput(session, request.Input)

			for i := range queries {
				queries[i].Input = p.settings.Callbacks.OnModifyBatchQueryInput(session, queries[i].Input)
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

	return &types.Route{
		Method:  http.MethodPost,
		Handler: handler,
	}
}

func (p *proxy) authzError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}
