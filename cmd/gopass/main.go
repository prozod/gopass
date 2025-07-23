package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/prozod/gopass/internal/common"
	"github.com/prozod/gopass/internal/vault"
)

func main() {
	var configFlag string
	var config []string
	flag.StringVar(&configFlag, "config", "", "Config for the vault file (Format: <filepath>:<password>)")
	flag.Parse()

	if configFlag != "" {
		parts := strings.Split(configFlag, ":")
		if len(parts) == 0 {
			fmt.Println("Invalid config format. Use -config <filepath>[:password]")
			os.Exit(1)
		}
		filePath := parts[0]
		_ = vault.SaveVaultAccessToConfig(filePath)
		config = []string{filePath}
	} else {
		config = vault.GetVaultPathFromConfig()
		if len(config) == 0 {
			fmt.Println(common.Red + "No saved config found. Use -config <filepath> at least once. It will generate a .gopassrc file in your home directory containing your vault path." + common.Reset)
			os.Exit(1)
		}
	}

	lastVault, _ := vault.GetLastVaultFilePath()
	if err := vault.ClearOldVaultPasswordIfNeeded(lastVault, config[0]); err != nil {
		fmt.Println("Error clearing old vault password:", err)
	}

	v, err := vault.Load(config[0])
	if err != nil {
		fmt.Println("An error while loading file: ", err)
		os.Exit(1)
	}
	fmt.Println(common.Green + "Current vault is: " + common.Reset + config[0])

	if err := vault.SetLastVaultFilePath(config[0]); err != nil {
		fmt.Println("Warning: failed to update last vault file:", err)
	}

	if len(os.Args) < 2 {
		fmt.Println("Welcome to Gopass, a simple password storage and encrypter.")
		fmt.Println("Type 'gopass help' for more info.")
	} else {
		if os.Args[1] == "help" {
			common.PrintHelp()
		} else {
			switch os.Args[1] {
			case "list":
				v.List()
			case "export":
				v.Export(os.Args[2])
			case "import":
				file, err := os.ReadFile(os.Args[2])
				if err != nil {
					fmt.Printf("Error opening file: %v", os.Args[2])
				}
				v.Import(file, config[0])
			case "add":
				v.Add(os.Args[2], os.Args[3], config[0])
				fmt.Printf("Added '%s' to %s\n", os.Args[2], config[0])
			case "get":
				_, err := v.Get(os.Args[2])
				if err != nil {
					fmt.Println(err)
				}
			default:
				common.PrintHelp()
			}
		}
	}
}
