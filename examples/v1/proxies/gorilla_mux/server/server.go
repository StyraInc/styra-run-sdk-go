package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	aproxy "github.com/styrainc/styra-run-sdk-go/api/v1/proxy"
	rproxy "github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy"
)

type WebServerSettings struct {
	Port            int
	Client          api.Client
	ClientCallbacks *aproxy.Callbacks
	RbacCallbacks   *rproxy.Callbacks
}

type WebServer interface {
	Listen() error
}

type webServer struct {
	settings *WebServerSettings
}

func NewWebServer(settings *WebServerSettings) WebServer {
	return &webServer{
		settings: settings,
	}
}

func (w *webServer) Listen() error {
	router := mux.NewRouter()

	// Client handlers.
	{
		proxy := aproxy.New(
			&aproxy.Settings{
				Client:    w.settings.Client,
				Callbacks: w.settings.ClientCallbacks,
			},
		)

		for _, route := range proxy.All() {
			router.HandleFunc(route.Path, route.Handler).Methods(route.Method)
		}
	}

	// Rbac handlers.
	{
		proxy := rproxy.New(
			&rproxy.Settings{
				Client:    w.settings.Client,
				Callbacks: w.settings.RbacCallbacks,
				GetUrlVar: func(r *http.Request, key string) string {
					return mux.Vars(r)[key]
				},
			},
		)

		for _, route := range proxy.All() {
			router.HandleFunc(route.Path, route.Handler).Methods(route.Method)
		}
	}

	port := fmt.Sprintf(":%d", w.settings.Port)

	return http.ListenAndServe(port, router)
}
