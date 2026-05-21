package main

import (
	"fmt"
	"os"

	"github.com/stpotter16/gather/internal/password"
	"golang.org/x/term"
)

func main() {
	fmt.Fprint(os.Stderr, "Password: ")

	raw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading password:", err)
		os.Exit(1)
	}

	hash, err := password.Hash(string(raw))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error hashing password:", err)
		os.Exit(1)
	}

	fmt.Println(hash)
}
