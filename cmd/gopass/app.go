package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/prozod/gopass/internal/common"
	"github.com/prozod/gopass/internal/vault"
)

func Run(args []string) int {
	os.Args = args
	main()
	return 0
}

func runCLI() int {
	var configFlag string
	var config string
	flag.StringVar(&configFlag, "config", "", "Config for the vault file (Format: <filepath>:<password>)")
	flag.Parse()

	if configFlag != "" {
		parts := strings.Split(configFlag, ":")
		if len(parts) == 0 {
			fmt.Println("Invalid config format. Use -config <filepath>[:password]")
			return 1
		}
		filePath := parts[0]
		_ = vault.SaveVaultAccessToConfig(filePath)
		config = filePath
	} else {
		cfg, err := vault.GetVaultPathFromConfig()
		if err != nil {
			fmt.Printf("Error while getting vault path from config (.gopassrc).")
		}
		config = cfg
		if config == "" {
			fmt.Println(common.Red + "No saved config found. Use -config <filepath> at least once. It will generate a .gopassrc file in your home directory containing your vault path." + common.Reset)
			return 1
		}
	}

	lastVault, err := vault.GetLastVaultFilePath()
	if err != nil {
		fmt.Println(err)
	}
	if err := vault.ClearOldVaultPasswordIfNeeded(lastVault, config); err != nil {
		fmt.Println("Error clearing old vault password:", err)
	}

	v, err := vault.Load(config)
	if err != nil {
		fmt.Println("An error while loading file: ", err)
		return 1
	}
	if err := vault.SetLastVaultFilePath(config); err != nil {
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
			case "vault":
				fmt.Println(common.Cyan + "Current vault: " + common.Reset + config)
			case "list":
				if len(os.Args) > 2 && os.Args[2] == "-expose" {
					v.List(os.Args[2])
				} else {
					v.List()
				}
			case "export":
				v.Export(os.Args[2])
			case "import":
				file, err := os.ReadFile(os.Args[2])
				if err != nil {
					fmt.Printf("Error opening file: %v", os.Args[2])
				}
				v.Import(file, config)
			case "add":
				v.Add(os.Args[2], os.Args[3], config)
			case "remove":
				v.Remove(os.Args[2], config)
			case "get":
				_, err := v.Get(os.Args[2])
				if err != nil {
					fmt.Println(err)
				}
			case "-config":
				fmt.Println(common.Green + "Switching vault to " + config + common.Reset)
			default:
				fmt.Println(common.Red + "Unrecognised flag/command, use 'gopass help' for available commands." + common.Reset)
			}
		}
	}
	return 0
}
