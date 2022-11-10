package main

import (
	"flag"
	"log"

	api "github.com/styrainc/styra-run-sdk-go/api/v1"
	"github.com/styrainc/styra-run-sdk-go/tests/v1/server"
)

func main() {
	token := flag.String("token", "", "token")
	url := flag.String("url", "", "url")
	port := flag.Int("port", 0, "port")

	flag.Parse()

	if *token == "" || *url == "" || *port == 0 {
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
		},
	)

	if err := ws.Listen(); err != nil {
		log.Fatal(err)
	}
}
