package common

import "fmt"

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"

	Bold = "\033[1m"
)

func PrintHelp() {
	fmt.Println()
	fmt.Printf(Bold + `Usage:` + Reset + "\n")
	fmt.Println(`  ` + Green + `gopass add <name> <password>` + Reset + ` — Add a new secret`)
	fmt.Println(`  ` + Blue + `gopass get <name>` + Reset + ` — Retrieve a password, copied to clipboard automatically.`)
	fmt.Println(`  ` + Yellow + `gopass list` + Reset + ` — List all stored secret names`)
	fmt.Println(`  ` + Purple + `gopass export <filename> (ex: mydata)` + Reset + ` — Export secrets to JSON`)
	fmt.Println(`  ` + Red + `gopass import <filepath> (ex: mydata.json)` + Reset + ` — Import secrets from JSON`)
	fmt.Println(`  ` + Cyan + `gopass -config <absolute filepath> (ex: ~/myvault.dat)` + Reset + ` — Import secrets from JSON`)
	fmt.Println()
	fmt.Println(Bold + `Current vault is cached and saved in a local config file (~/.gopassrc).` + Reset)
}
