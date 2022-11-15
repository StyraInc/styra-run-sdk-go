package put_data

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type PutDataResponse struct{}

type Settings struct {
	Client  api.Client
	GetPath types.GetVar
}

func New(settings *Settings) *types.Proxy {
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

		path := settings.GetPath(r)

		if err := settings.Client.PutData(r.Context(), path, data); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &PutDataResponse{}
		utils.WriteResponse(w, response)
	}

	return &types.Proxy{
		Method:  http.MethodPut,
		Handler: handler,
	}
}
