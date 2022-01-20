package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGenerateAPIToken(t *testing.T) {
	var newToken string = generateAPIToken()
	if len(newToken) < 19 {
		fmt.Println("len(newToken)", len(newToken))
		t.Error("token length should be longer than 19 (including prefix)")
	}
	if newToken[:3] != "sk_" {
		t.Error("api secret tokens should start with sk_")
	}
}

func TestHashAPIToken(t *testing.T) {
	token := "sk_0d588b42621105e7f77b2c584a7be4e2"
	hashedToken := hashAPIToken(token)
	if len(hashedToken) != 32 {
		t.Error("should return a sha256 hashed string of length 32")
	}
	if token == hashedToken {
		errMsg := fmt.Sprint("hashed token can't be the same as input token! input: ", token, "output:", hashedToken)
		t.Error(errMsg)
	}
}

// for reference on gin testing https://github.com/gin-gonic/gin/blob/1b28e2b0303b6e5ecdea70890ba1ee8c5950892b/auth_test.go#L115

func TestGetTokenFromHeader(t *testing.T) {

	expectedToken := "sk_0d588b42621105e7f77b2c584a7be4e2"
	var resultToken string
	var err error

	// create a new router with the fuction to test
	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		resultToken, err = getTokenFromHeader(c)
	})

	// prepare a request to send to the endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	// use the expectedToken as the bearer token
	req.Header.Set("Authorization", "Bearer sk_0d588b42621105e7f77b2c584a7be4e2")
	//send the request
	router.ServeHTTP(w, req)

	// test the results
	if err != nil {
		t.Error("error not nil: ", err.Error())
	}
	if resultToken != expectedToken {
		t.Error("result token and expected token not equal")
	}
}
