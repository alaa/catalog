package deploy

// deploy to marathon
import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/alaa/catalog/merger"
	"github.com/alaa/catalog/vault"
	marathon "github.com/gambol99/go-marathon"
	"github.com/gorilla/mux"
	multierror "github.com/hashicorp/go-multierror"
)

// DeployWithSecrets Merge and overrwrite vault secrets with marathon env vars.
func DeployWithSecrets(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	secretPath := fmt.Sprintf("/secret/%s", vars["name"])
	vaultToken := r.Header["X-Vault-Token"][0]

	// parse http request body into marathon apps.
	marathonApps, err := parseService(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "unprocessable entity, invalid JSON payload.")
		return
	}

	// get vault secrets.
	vault, err := vault.New(vaultToken)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Forbiddin, Could not connect to vault. %s", err)
		return
	}

	marathonSecrets, err := vault.GetMarathonSecret(secretPath, vaultToken)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Secret path not found at %s, %s", secretPath, err)
		return
	}

	// merge marathon application environment map with vault secrets.
	var errs error
	var deploymentIDs []string

	for _, marathonApp := range marathonApps {
		merger.EnvMerge(*marathonApp.Env, *marathonSecrets.Env)

		// deploy to marathon.
		deploymentID, err := toMarathon(marathonApp)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		deploymentIDs = append(deploymentIDs, deploymentID)
	}

	if errs != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error deploying to marathon, error: %s", errs)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf(strings.Join(deploymentIDs, ", ")))
	return
}

func parseService(rbody io.Reader) ([]marathon.Application, error) {
	var apps []marathon.Application
	decoder := json.NewDecoder(rbody)
	if err := decoder.Decode(&apps); err != nil {
		log.Printf("Error: unmarshalling marathon definition %s", err)
		return []marathon.Application{}, err
	}
	return apps, nil
}

func toMarathon(task marathon.Application) (string, error) {
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

	deploy, err := client.UpdateApplication(&task, false)
	if err != nil {
		log.Printf("Failed to update application: %s, error: %s", task.ID, err)
		return "", err
	}
	return deploy.DeploymentID, nil
}

func envFetch(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", errors.New(fmt.Sprintf("Key: %s is not set!", key))
	}
	return value, nil
}
