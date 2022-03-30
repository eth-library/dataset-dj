package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
)

// prints a new admin API Key to the command line
// useful if admin token needs to be replaced.
// generate the new token and replace the existing ADMIN_KEY in the .env file
// delete the existing admin key (with field  permission: admin)
// restart the service to load the new key from the .env file

func generateAdminAPIToken() string {
	length := 16
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	if err != nil {
		return ""
	}
	return "ak_" + hex.EncodeToString(b)
}

func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

func main() {

	token := generateAdminAPIToken()
	fmt.Println("token: ", token)

	showHash := flag.Bool("hash", false, "also display the hashed token (e.g. for testing/debugging purposes)")
	flag.Parse()

	if *showHash {
		hashedToken := hashToken(token)
		fmt.Println("hashedToken: ", hashedToken)
	}

}
