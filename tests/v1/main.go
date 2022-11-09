package main

import (
	"flag"
	"log"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	aproxy "github.com/styrainc/styra-run-sdk-go/api/v1/proxy"
	"github.com/styrainc/styra-run-sdk-go/examples/v1/proxies/gorilla_mux/server"
	rproxy "github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy"
	"github.com/styrainc/styra-run-sdk-go/types"
)

// var (
// 	users = []*rbac.User{
// 		{Id: "alice"},
// 		{Id: "bob"},
// 		{Id: "bryan"},
// 		{Id: "cesar"},
// 		{Id: "emily"},
// 		{Id: "gary"},
// 		{Id: "henry"},
// 		{Id: "kevin"},
// 		{Id: "lynn"},
// 		{Id: "jiri"},
// 		{Id: "larry"},
// 		{Id: "alan"},
// 	}
// )

func main() {
	token := flag.String("token", "", "token")
	url := flag.String("url", "", "url")
	port := flag.Int("port", 0, "port")

	flag.Parse()

	if *token == "" || *url == "" || *port == 0 {
		flag.PrintDefaults()
		return
	}

	client := api.Default(
		&api.DefaultSettings{
			Token: *token,
			Url:   *url,
		},
	)

	ws := server.NewWebServer(
		&server.WebServerSettings{
			Port:   *port,
			Client: client,
			ClientCallbacks: aproxy.DefaultCallbacks(
				&aproxy.DefaultCallbackSettings{
					GetSession: types.SessionFromCookie(),
				},
			),
			RbacCallbacks: &rproxy.Callbacks{
				GetSession: types.SessionFromCookie(),
			},
			// RbacCallbacks: rproxy.ArrayCallbacks(
			// 	&rproxy.ArrayCallbackSettings{
			// 		GetSession: api.SessionFromCookie(),
			// 		Users:      users,
			// 		PageSize:   3,
			// 	},
			// ),
		},
	)

	if err := ws.Listen(); err != nil {
		log.Fatal(err)
	}
}
