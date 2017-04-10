package main

import (
	"log"
	"net/http"

	"github.com/alaa/catalog/deploy"
	"github.com/alaa/catalog/router"
)

func main() {
	routes := router.Routes{
		router.Route{"Index", "GET", "/", router.Index},
		router.Route{"Health", "GET", "/health", router.Index},
		router.Route{"Deploy", "POST", "/service/{name}", deploy.DeployWithSecrets},
	}

	r := router.NewRouter(routes)
	r.NotFoundHandler = http.HandlerFunc(router.NotFound)
	log.Fatal(http.ListenAndServe(":8080", r))
}
