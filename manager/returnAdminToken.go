package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

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

func main() {

	token := generateAdminAPIToken()
	fmt.Println("token: ", token)
}
