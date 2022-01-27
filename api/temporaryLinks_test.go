package main

import (
	"fmt"
	"testing"
)

func testPublishAPILinkEmailTask(t *testing.T) {

	url := "https://ethz.ch"
	content := fmt.Sprintf(`Welcome to the Dataset DJ

	below is a single use link that returns a API Key.
	This API Key should be kept secret, and not disclosed to users or client side code.

	%v

	`, url)
	fmt.Println("lkansldknalksnlaksnd")
	fmt.Print("content: ", content)

	if true {
		t.Errorf("text test")
	}
}
