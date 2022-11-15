package delete_data

import (
	"net/http"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
	"github.com/styrainc/styra-run-sdk-go/types"
)

type DeleteDataResponse struct{}

type Settings struct {
	Client  api.Client
	GetPath types.GetVar
}

func New(settings *Settings) *types.Proxy {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if !utils.HasMethod(w, r, http.MethodDelete) {
			return
		}

		path := settings.GetPath(r)

		if err := settings.Client.DeleteData(r.Context(), path); err != nil {
			utils.ForwardHttpError(w, err)
			return
		}

		response := &DeleteDataResponse{}
		utils.WriteResponse(w, response)
	}

	return &types.Proxy{
		Method:  http.MethodDelete,
		Handler: handler,
	}
}
