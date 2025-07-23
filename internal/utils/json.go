package utils

import "github.com/prozod/gopass/internal/vault"

type Importer interface {
	ImportJSON(path string) error
}

type Exporter interface {
	ExportJSON(path string) error
}

type ImportJSON struct {
	filepath string
	entries  vault.Vault
}
