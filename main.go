package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/alaa/catalog/vault"
	"github.com/gambol99/go-marathon"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/health", Index)
	router.NotFoundHandler = http.HandlerFunc(NotFound)

	router.HandleFunc("/service/{name}", ReadService).
		Methods("POST").
		HeadersRegexp("X-Vault-Token", vault.TokenPattern)

	log.Fatal(http.ListenAndServe(":8080", router))
}

// NotFound default handler for undefined routes.
func NotFound(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Not Found")
}

// Index default index handler
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Catalog is running!")
}

// ReadService handles the communication between vault secrets and
// the requested marathon job.
func ReadService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	secretPath := fmt.Sprintf("/secret/%s", vars["name"])

	defer r.Body.Close()
	vaultToken := r.Header["X-Vault-Token"][0]
	app, err := parseService(r.Body)
	if err != nil {
		fmt.Fprintf(w, "unprocessable entity, service.json is not valid")
		return
	}

	secret, err := getSecret(secretPath, vaultToken)
	if err != nil || secret == nil {
		fmt.Fprintf(w, "No value found at %s, %s", secretPath, err)
		return
	}

	newapp, err := merge(secret, app)
	if err != nil {
		fmt.Fprintf(w, "Failed to merge, error: %s", err)
		return
	}

	id, err := deploy(newapp)
	if err != nil {
		log.Printf("Error deploying to marathon: %s", err)
	}

	log.Printf("Marathon Deployment ID: %s with vault secret on %s", id, secretPath)

	fmt.Fprintf(w, id)
	return
}

// parse and unmarshall json request into service.
func parseService(rbody io.Reader) (marathon.Application, error) {
	var app marathon.Application
	decoder := json.NewDecoder(rbody)
	if err := decoder.Decode(&app); err != nil {
		log.Printf("Error: unmarshalling service %s", err)
		return marathon.Application{}, err
	}
	return app, nil
}

// get the secret from vault
func getSecret(path string, vaultToken string) (map[string]interface{}, error) {
	vault, err := vault.New(vaultToken)
	if err != nil {
		log.Println("Cloud not initialize vault client", err)
	}

	secret, err := vault.ReadSecret(path)
	if err != nil || secret == nil {
		return nil, err
	}
	return secret.Data, nil
}

// merge marathon application env with the service request
func merge(secret map[string]interface{}, app marathon.Application) (marathon.Application, error) {
	var t marathon.Application
	err := mapstructure.Decode(secret, &t)
	if err != nil {
		log.Printf("Could not unmarshal vault secrets %s", err)
		return marathon.Application{}, err
	}

	for key := range *t.Env {
		(*app.Env)[key] = (*t.Env)[key]
	}
	return app, nil
}

// deploy to marathon
func deploy(task marathon.Application) (string, error) {
	url, err := envFetch("MARATHON_URL")
	if err != nil {
		return "", err
	}

	config := marathon.NewDefaultConfig()
	config.URL = url
	client, err := marathon.NewClient(config)
	if err != nil {
		log.Printf("Failed to initialize marathon client %s", err)
	}

	id, err := client.UpdateApplication(&task, false)
	if err != nil {
		log.Printf("Failed to update application: %s, error: %s", task.ID, err)
	}
	return id.DeploymentID, err
}

func envFetch(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", errors.New(fmt.Sprintf("Key: %s is not set!", key))
	}
	return value, nil
}
