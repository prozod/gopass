// Package vault provides functionality for secure password storage,
// encryption and decryption of vault files, and management of entries.
// It contains AES-GCM encryption with password-derived keys and
// importing/exporting entries in JSON format.
package vault

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"strings"

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
	if name == "" || value == "" {
		return fmt.Errorf("key and value cannot be empty")
	}
	if _, exists := v.Entries[name]; exists {
		fmt.Printf(common.Red+"Entry with name '%s' already exists in '%s', skipping...\n"+common.Reset, name, filepath)
		return fmt.Errorf("entry with name '%s' already exists", name)
	}
	v.Entries[name] = value
	if err := v.Save(filepath); err != nil {
		fmt.Printf("Error adding %v to file: %v. -> %v", os.Args[2], filepath, err)
		return err
	} else {
		fmt.Println(common.Green + "Added " + common.Reset + os.Args[2] + common.Green + " to " + common.Reset + filepath)
		return nil
	}
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

func (v *Vault) Remove(name, filepath string) error {
	if _, exists := v.Entries[name]; exists {
		delete(v.Entries, name)
		fmt.Println(common.Green + "Deleted " + common.Reset + name + common.Green + " from vault" + common.Reset)
		return v.Save(filepath)
	} else {
		return fmt.Errorf("entry with name '%s' doesn't exists", name)
	}
}

func (v Vault) List(args ...string) {
	fmt.Println(common.Green + "INFO: " + common.Reset + "Entries are separated by ':' (<" + common.Blue + "name" + common.Reset + ">:<" + common.Yellow + "value" + common.Reset + ">)")
	fmt.Println()
	fmt.Println("---------- VAULT STORAGE ----------")
	if len(args) > 0 {
		if args[0] == "-expose" {
			for n, v := range v.Entries {
				fmt.Printf("|> "+common.Blue+"%s"+common.Reset+":"+common.Yellow+"%s"+common.Reset+"\n", n, v)
			}
		} else {
			fmt.Printf(common.Yellow+"WARNING: "+common.Reset+"Unknown argument: %s\n", args[0])
		}
	} else {
		fmt.Println(common.Purple + "Hidden mode, use flag '-expose' to display passwords." + common.Reset)
		for n, v := range v.Entries {
			fmt.Printf("|> "+common.Blue+"%s"+common.Reset+":"+common.Yellow+"%s"+common.Reset+"\n", n, strings.Repeat("*", len(strings.Split(v, ""))))
		}
	}
	fmt.Println("-----------------------------------")
	fmt.Println()
}

func (v Vault) Export(path string) error {
	fmt.Println(common.Green + "Exporting vault to JSON..." + common.Reset)
	dataToExport := make(map[string]string)
	maps.Copy(dataToExport, v.Entries)
	data, err := json.MarshalIndent(dataToExport, "", "\t")
	if err != nil {
		fmt.Println("Error exporting to JSON. ", err)
	}
	err = os.WriteFile(path, data, 0o644)
	if err != nil {
		fmt.Printf("Error creating JSON file")
	}
	return err
}

func (v Vault) Import(jsonFile []byte, filepath string) error {
	fmt.Println(common.Green + "Importing JSON to vault..." + common.Reset)

	var raw map[string]any
	if err := json.Unmarshal(jsonFile, &raw); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	dataToImport := make(map[string]string)
	for name, val := range raw {
		strVal, ok := val.(string)
		if !ok {
			return fmt.Errorf("invalid value for key '%s': only flat key-value strings are supported", name)
		}
		dataToImport[name] = strVal
	}

	for name, value := range dataToImport {
		if err := v.Add(name, value, filepath); err != nil {
			fmt.Println(err)
		}
	}

	return nil
}
