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
	ClientPrefix    string
	ClientCallbacks *aproxy.Callbacks
	RbacPrefix      string
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

		w.InstallClient(proxy, w.settings.ClientPrefix, router)
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

		w.InstallRbac(proxy, w.settings.RbacPrefix, router)
	}

	port := fmt.Sprintf(":%d", w.settings.Port)

	return http.ListenAndServe(port, router)
}

func (w *webServer) InstallClient(proxy aproxy.Proxy, prefix string, router *mux.Router) {
	install := func(route *aproxy.Route) {
		router.HandleFunc(prefix+route.Path, route.Handler).Methods(route.Method)
	}

	install(proxy.BatchQuery())
}

func (w *webServer) InstallRbac(proxy rproxy.Proxy, prefix string, router *mux.Router) {
	install := func(route *rproxy.Route) {
		router.HandleFunc(prefix+route.Path, route.Handler).Methods(route.Method)
	}

	install(proxy.GetRoles())
	install(proxy.ListUserBindings())
	install(proxy.GetUserBinding())
	install(proxy.PutUserBinding())
	install(proxy.DeleteUserBinding())
}
