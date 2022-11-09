package main

import (
	"flag"
	"log"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/examples/v1/proxies/gorilla_mux/server"
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
)

var (
	users = []*rbac.User{
		{Id: "alice"},
		{Id: "bob"},
		{Id: "bryan"},
		{Id: "cesar"},
		{Id: "emily"},
		{Id: "gary"},
		{Id: "henry"},
		{Id: "kevin"},
		{Id: "lynn"},
		{Id: "jiri"},
		{Id: "larry"},
		{Id: "alan"},
		{Id: "terence"},
		{Id: "eckhart"},
	}
)

func main() {
	token := flag.String("token", "", "token")
	url := flag.String("url", "", "url")
	port := flag.Int("port", 3000, "port")

	flag.Parse()

	if *token == "" || *url == "" {
		flag.PrintDefaults()
		return
	}

	client := api.New(
		&api.Settings{
			Token:             *token,
			Url:               *url,
			DiscoveryStrategy: api.Simple,
			MaxRetries:        3,
		},
	)

	ws := server.NewWebServer(
		&server.WebServerSettings{
			Port:   *port,
			Client: client,
			//RbacCallbacks: rproxy.ArrayCallbacks(
			//	&rproxy.ArrayCallbackSettings{
			//		GetSession: api.SessionFromValues(tenant, subject),
			//		Users:      users,
			//		PageSize:   3,
			//	},
			//),
		},
	)

	if err := ws.Listen(); err != nil {
		log.Fatal(err)
	}
}
