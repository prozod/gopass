package vault

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"

	"github.com/atotto/clipboard"
	"github.com/prozod/gopass/internal/common"
)

type VaultKVJson struct {
	Key   string
	Value string
}

type VaultStore interface {
	Load() error
	Save() error
	Add(name, value string) error
	Get(name string) (string, error)
	Remove(name string) error
	List() []string
}

type Vault struct {
	Entries map[string]string `json:"entries"`
}

func (v *Vault) Add(name, value, filepath string) error {
	if _, exists := v.Entries[name]; exists {
		fmt.Printf(common.Red+"Entry with name '%s' already exists in '%s', skipping...\n"+common.Reset, name, filepath)
		return fmt.Errorf("entry with name '%s' already exists", name)
	}
	v.Entries[name] = value
	return v.Save(filepath)
}

func (v Vault) Get(name string) (string, error) {
	value, exists := v.Entries[name]
	if exists {
		err := clipboard.WriteAll(value)
		if err != nil {
			log.Fatalf("Failed to copy to clipboard: %v", err)
		}
		fmt.Printf("Copied value for \"%s\" to clipboard.\n", name)
		return value, nil
	} else {
		return "", fmt.Errorf("'%s' doesnt exist in vault", name)
	}
}

func (v *Vault) Remove(name, value, filepath, password string) error {
	if _, exists := v.Entries[name]; exists {
		delete(v.Entries, name)
		return v.Save(filepath)
	} else {
		return fmt.Errorf("entry with name '%s' doesn't exists", name)
	}
}

func (v Vault) List() {
	fmt.Println(common.Green + "INFO: " + common.Reset + "Entries are separated by ':' (<" + common.Blue + "name" + common.Reset + ">:<" + common.Yellow + "value" + common.Reset + ">)")
	fmt.Println()
	fmt.Println("---------- VAULT STORAGE ----------")
	for n, v := range v.Entries {
		fmt.Printf("|> "+common.Blue+"%s"+common.Reset+":"+common.Yellow+"%s"+common.Reset+"\n", n, v)
	}
	fmt.Println("-----------------------------------")
	fmt.Println()
}

func (v Vault) Export(path string) {
	fmt.Println(common.Green + "Exporting vault to JSON..." + common.Reset)
	dataToExport := make(map[string]string)
	maps.Copy(dataToExport, v.Entries)
	data, err := json.MarshalIndent(dataToExport, "", "\t")
	if err != nil {
		fmt.Println("Error exporting to JSON. ", err)
	}
	fmt.Println(dataToExport)
	fmt.Println(string(data))
	err = os.WriteFile(path+".json", data, 0o644)
	if err != nil {
		fmt.Printf("Error creating JSON file")
	}
}

func (v Vault) Import(jsonFile []byte, filepath string) {
	fmt.Println(common.Green + "Importing JSON  to vault..." + common.Reset)
	dataToImport := make(map[string]string)
	json.Unmarshal(jsonFile, &dataToImport)
	for name, value := range dataToImport {
		err := v.Add(name, value, filepath)
		if err != nil {
			fmt.Println(err)
		}
	}
}
