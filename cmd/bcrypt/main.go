package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

/*
CLI utility for hashing passwords. Use examples:

1. BUILD: 	go build ./cmd/bcrypt
2. HASH:  	bcrypt hash "some password here"
3. COMPARE: bcrypt compare "some password here" "some hash here"

During development, you may want to build/run in the same step:

1. BUILD/RUN/HASH: 		go run cmd/bcrypt/main.go hash "some password here"
2. BUILD/RUN/COMPARE: 	go run cmd/bcrypt/main.go compare "abcd" 'hashed'

IMPORTANT: Use single quotes around the hashed password, so the $ in the string
are not interpreted as parameter expansion!
*/
func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	command := os.Args[1]
	arg1 := os.Args[2]
	switch command {
	case "hash":
		if len(os.Args) != 3 {
			fmt.Println("usage: bcrypt hash <password>")
			os.Exit(2)
		}
		hash(arg1)
	case "compare":
		if len(os.Args) != 4 {
			fmt.Println("usage: bcrypt compare <password> <hash>")
			os.Exit(2)
		}
		arg2 := os.Args[3]
		compare(arg1, arg2)
	default:
		fmt.Printf("invalid command: %s\n", command)
		usage()
		os.Exit(2)
	}
}

func hash(password string) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("error hashing: %v\n", password)
	}
	hashString := string(hashedBytes)
	fmt.Println(hashString)
}

func compare(password string, hash string) {
	fmt.Println("comparing", password, hash)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Printf("error comparing: %v\n", password)
		return
	}
	fmt.Println("password matches!")
}

func usage() {
	fmt.Println("usage:")
	fmt.Println("  bcrypt hash <password>")
	fmt.Println("  bcrypt compare <password> <hash>")
}
