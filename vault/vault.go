package vault

import vaultapi "github.com/hashicorp/vault/api"

var TokenPattern = "^{?([a-z0-9]{8})-([a-z0-9]{4})-([1-5][a-z0-9]{3})-([a-z0-9]{4})-([a-z0-9]{12})\\}?$"

type Vault struct {
	client  *vaultapi.Client
	logical *vaultapi.Logical
}

func New(vaultToken string) (*Vault, error) {
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
