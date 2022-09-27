package tickethub

import (
	"net/http"

	"github.com/gorilla/mux"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	aproxy "github.com/styrainc/styra-run-sdk-go/api/v1/proxy"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
	rproxy "github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy"
)

func DefaultClient(token, url string) api.Client {
	return api.New(
		&api.Settings{
			Token:             token,
			Url:               url,
			DiscoveryStrategy: api.Simple,
			MaxRetries:        3,
		},
	)
}

func DefaultClientProxy(client api.Client) aproxy.Proxy {
	return aproxy.New(
		&aproxy.Settings{
			Client: client,
			Callbacks: aproxy.DefaultCallbacks(
				&aproxy.DefaultCallbackSettings{
					GetAuthz: api.AuthzFromCookie(),
				},
			),
		},
	)
}

func InstallDefaultClientProxy(router *mux.Router, client api.Client) {
	proxy := DefaultClientProxy(client).BatchQuery()

	router.HandleFunc("/api/authz", proxy.Handler).Methods(proxy.Method)
}

func DefaultRbac(client api.Client) rbac.Rbac {
	return rbac.New(
		&rbac.Settings{
			Client: client,
		},
	)
}

func DefaultRbacProxy(client api.Client) rproxy.Proxy {
	return rproxy.New(
		&rproxy.Settings{
			Client: client,
			GetUrlVar: func(r *http.Request, key string) string {
				return mux.Vars(r)[key]
			},
			Callbacks: rproxy.DefaultCallbacks(
				&rproxy.DefaultCallbackSettings{
					GetAuthz: api.AuthzFromCookie(),
				},
			),
		},
	)
}

func InstallDefaultRbacProxy(router *mux.Router, client api.Client) {
	proxy := DefaultRbacProxy(client)
	install := func(route *rproxy.Route) {
		router.HandleFunc("/api/rbac"+route.Path, route.Handler).Methods(route.Method)
	}

	install(proxy.GetRoles())
	install(proxy.ListUserBindings())
	install(proxy.GetUserBinding())
	install(proxy.PutUserBinding())
	install(proxy.DeleteUserBinding())
}
