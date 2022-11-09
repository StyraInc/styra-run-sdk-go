package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
	"github.com/styrainc/styra-run-sdk-go/types"

	"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/batch_query"
	"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/check"
	"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/delete_data"
	"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/get_data"
	"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/put_data"
	"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/query"
	"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/shared"
	"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/delete_user_binding"
	"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/get_roles"
	"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/get_user_binding"
	"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/list_user_bindings_all"
	"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/put_user_binding"
)

const (
	tenant  = "acmecorp"
	subject = "alice"
)

type WebServerSettings struct {
	Port   int
	Client api.Client
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

	getSession := api.SessionFromValues(tenant, subject)

	// Client handlers.
	{
		// Get data.
		install(get_data.New(
			&get_data.Settings{
				Client:  w.settings.Client,
				GetPath: key("path"),
			}), "/data/{path:.*}",
		)

		// Put data.
		install(put_data.New(
			&put_data.Settings{
				Client:  w.settings.Client,
				GetPath: key("path"),
			}), "/data/{path:.*}",
		)

		// Delete data.
		install(delete_data.New(
			&delete_data.Settings{
				Client:  w.settings.Client,
				GetPath: key("path"),
			}), "/data/{path:.*}",
		)

		// Query.
		install(query.New(
			&query.Settings{
				Client:  w.settings.Client,
				GetPath: key("path"),
			}), "/query/{path:.*}",
		)

		// Check.
		install(check.New(
			&check.Settings{
				Client:  w.settings.Client,
				GetPath: key("path"),
			}), "/check/{path:.*}",
		)

		// Batch query.
		install(batch_query.New(
			&batch_query.Settings{
				Client:        w.settings.Client,
				GetSession:    getSession,
				OnModifyInput: shared.DefaultOnModifyInput(),
			}), "/batch_query",
		)
	}

	// Rbac handlers.
	{
		myRbac := rbac.New(
			&rbac.Settings{
				Client: w.settings.Client,
			},
		)

		// Get roles.
		install(get_roles.New(
			&get_roles.Settings{
				Rbac:       myRbac,
				GetSession: getSession,
			}), "/roles",
		)

		// List user bindings all.
		install(list_user_bindings_all.New(
			&list_user_bindings_all.Settings{
				Rbac:       myRbac,
				GetSession: getSession,
			}), "/user_bindings_all",
		)

		// List user bindings.
		//install(list_user_bindings.New(
		//	&list_user_bindings.Settings{
		//		Rbac:       myRbac,
		//		GetSession: getSession,
		//	}), "/user_bindings_all",
		//)

		// Get user binding.
		install(get_user_binding.New(
			&get_user_binding.Settings{
				Rbac:       myRbac,
				GetSession: getSession,
				GetId:      key("id"),
			}), "/user_bindings/{id}",
		)

		// Put user binding.
		install(put_user_binding.New(
			&put_user_binding.Settings{
				Rbac:       myRbac,
				GetSession: getSession,
				GetId:      key("id"),
			}), "/user_bindings/{id}",
		)

		// Delete user binding.
		install(delete_user_binding.New(
			&delete_user_binding.Settings{
				Rbac:       myRbac,
				GetSession: getSession,
				GetId:      key("id"),
			}), "/user_bindings/{id}",
		)
	}

	port := fmt.Sprintf(":%d", w.settings.Port)

	return http.ListenAndServe(port, router)
}
