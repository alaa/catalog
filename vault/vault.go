package vault

import (
	"errors"
	"log"
	"regexp"

	marathon "github.com/gambol99/go-marathon"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
)

type Vault struct {
	client  *vaultapi.Client
	logical *vaultapi.Logical
}

func New(vaultToken string) (*Vault, error) {
	if err := IsValidUUID(vaultToken); err != nil {
		return nil, err
	}

	client, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	if err != nil {
		return &Vault{}, err
	}

	if vaultToken != "" {
		client.SetToken(vaultToken)
	}

	return &Vault{
		client:  client,
		logical: client.Logical(),
	}, nil
}

func (v *Vault) ReadSecret(path string) (*vaultapi.Secret, error) {
	secret, err := v.logical.Read(path)
	if err != nil {
		return &vaultapi.Secret{}, err
	}
	return secret, nil
}

// validate Vault Token 8-4-4-4-12 hex sequence.
func IsValidUUID(uuid string) error {
	if valid := isValidUUID(uuid); !valid {
		return errors.New("Invalid UUID Token. Token should match 8-4-4-4-12 hex sequence")
	}
	return nil
}

func isValidUUID(uuid string) bool {
	regexString := "^[a-fa-f0-9]{8}-[a-fa-f0-9]{4}-[a-fa-f0-9]{4}-[a-fa-f0-9]{4}-[a-fa-f0-9]{12}$"
	r := regexp.MustCompile(regexString)
	return r.MatchString(uuid)
}

func (v *Vault) GetMarathonSecret(path string, vaultToken string) (marathon.Application, error) {
	secret, err := v.ReadSecret(path)
	if err != nil {
		return marathon.Application{}, err
	}

	if secret == nil {
		return marathon.Application{}, errors.New("Vault secret path not found")
	}

	var app marathon.Application
	if err := mapstructure.Decode(secret.Data, &app); err != nil {
		log.Printf("Error: unmarshalling vault secret for marathon %s", err)
		return marathon.Application{}, err
	}

	return app, nil
}
