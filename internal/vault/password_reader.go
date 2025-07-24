package vault

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

type PasswordReader interface {
	Read(prompt string) (string, error)
}

type TerminalPasswordReader struct{}

func (TerminalPasswordReader) Read(prompt string) (string, error) {
	fmt.Print(prompt)
	passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(passBytes), nil
}

type StaticPasswordReader struct {
	Password string
}

func (s StaticPasswordReader) Read(prompt string) (string, error) {
	return s.Password, nil
}
