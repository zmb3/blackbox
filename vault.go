package main

import (
	vault "github.com/hashicorp/vault/api"
)

type vaultStorer struct {
	client *vault.Client
}

// Store stores the data to a Hashicorp Vault instance.
func (v *vaultStorer) Store(path string, data map[string]interface{}) error {
	l := v.client.Logical()
	_, err := l.Write(path, data)
	return err
}
