package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	aproxy "github.com/styrainc/styra-run-sdk-go/api/v1/proxy"
	rproxy "github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy"
	"github.com/styrainc/styra-run-sdk-go/types"
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

	key := func(key string) types.GetVar {
		return func(r *http.Request) string {
			return mux.Vars(r)[key]
		}
	}

	install := func(route *types.Route, path string) {
		router.HandleFunc(path, route.Handler).Methods(route.Method)
	}

	// Client handlers.
	{
		proxy := aproxy.New(
			&aproxy.Settings{
				Client:    w.settings.Client,
				Callbacks: w.settings.ClientCallbacks,
			},
		)

		install(proxy.GetData(key("path")), "/data/{path:.*}")
		install(proxy.PutData(key("path")), "/data/{path:.*}")
		install(proxy.DeleteData(key("path")), "/data/{path:.*}")
		install(proxy.Query(key("path")), "/query/{path:.*}")
		install(proxy.Check(key("path")), "/check/{path:.*}")
		install(proxy.BatchQuery(), "/batch_query")
	}

	// Rbac handlers.
	{
		proxy := rproxy.New(
			&rproxy.Settings{
				Client:    w.settings.Client,
				Callbacks: w.settings.RbacCallbacks,
			},
		)

		install(proxy.GetRoles(), "/roles")
		install(proxy.ListUserBindings(), "/user_bindings")
		install(proxy.GetUserBinding(key("id")), "/user_bindings/{id}")
		install(proxy.PutUserBinding(key("id")), "/user_bindings/{id}")
		install(proxy.DeleteUserBinding(key("id")), "/user_bindings/{id}")
	}

	port := fmt.Sprintf(":%d", w.settings.Port)

	return http.ListenAndServe(port, router)
}
