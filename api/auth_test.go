package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mockCollection struct {
	// add a Mock object instance
	mock.Mock
	// other fields go here as normal
}

func (m *mockCollection) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {

	r := mongo.DeleteResult{
		DeletedCount: 0,
	}
	err := errors.New("apiKey not found")

	return &r, err
}

func (m *mockCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	r := &mongo.DeleteResult{
		DeletedCount: 0,
	}
	return r, nil
}

func (m *mockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	r := &mongo.SingleResult{}
	return r
}

// func TestDeleteToken(t *testing.T) {

// 	//case: provided apiKey not in database
// 	mockCol := mockCollection{}

// 	// r, e := mockCol.DeleteMany(context.TODO(), struct{}{})
// 	// fmt.Println(e.Error())
// 	// fmt.Printf("r: %v\ne: %v\n", r, e)
// 	err := deleteToken(mockCol, "foo")

// 	if err != nil {
// 		t.Errorf("Error: in deleteToken: %v", err)
// 	}

// }

// func dropCollection(ctx context.Context, client *mongo.Client) {
// 	col := client.Database(config.DbName).Collection("apiKeys")
// 	result, err := col.DeleteMany(ctx, bson.M{})
// }

// When implementing the methods of an interface, you wire your functions up
// to call the Mock.Called(args...) method, and return the appropriate values.
//
// For example, to mock a method that saves the name and age of a person and returns
// the year of their birth or an error, you might write this:
//

var ErrNoDocuments = errors.New("mongo: no documents in result")

// start with empty database

//

// test admin token init

// test create invite link

//

func TestGenerateAPIToken(t *testing.T) {
	var newToken string = generateAPIToken("service")
	if len(newToken) < 20 {
		fmt.Println("len(newToken)", len(newToken))
		t.Error("token length should be longer than 20 (including prefix)")
	}
	if newToken[:3] != "sk_" {
		t.Error("api tokens for services should start with sk_")
	}
}

func TestHashAPIToken(t *testing.T) {
	token := "sk_0d588b42621105e7f77b2c584a7be4e2"
	hashedToken1 := hashAPIToken(token)
	hashedToken2 := hashAPIToken(token) //check that the hash function is repeatable

	if hashedToken1[:3] == "sk_" {
		t.Error("hashed token should not start with 'sk_' got: " + hashedToken1[:3] + "...")
	}
	if token == hashedToken1 {
		errMsg := fmt.Sprint("hashed token can't be the same as input token! input: ", token, "output:", hashedToken1)
		t.Error(errMsg)
	}
	if hashedToken1 != hashedToken2 {
		errMsg := fmt.Sprintf("same input to hashed token should reproduce same output. input: %v output1: %v output2: %v", token, hashedToken1, hashedToken2)
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
