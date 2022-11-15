package get_data

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type GetDataResponse struct {
	Result interface{} `json:"result"`
}

type Settings struct {
	Client  api.Client
	GetPath types.GetVar
}

func New(settings *Settings) *types.Proxy {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodGet) {
			return
		}

		path := settings.GetPath(r)

		var data interface{}
		if err := settings.Client.GetData(r.Context(), path, &data); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &GetDataResponse{
			Result: data,
		}

		utils.WriteResponse(w, response)
	}

	return &types.Proxy{
		Method:  http.MethodGet,
		Handler: handler,
	}
}
