package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/eth-library-lab/dataset-dj/dbutil"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type authHeader struct {
	Authorization string `header:"Authorization"`
}

//APIKey is the format of the mongodb document that stores the keys
type APIKey struct {
	HashedToken string `bson:"hashedToken,omitempty"`
	OwnerID     int32  `bson:"ownerId,omitempty"`
	CreatedDate string `bson:"createdDate,omitempty"`
}

//AuthMiddleware validates the bearer token before
//allowing the handler to be a called
func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		token, err := getTokenFromHeader(c)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}
		if token == "" {
			c.IndentedJSON(http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		res := validateAPIToken(runfig.MongoCtx, runfig.MongoClient, token)
		if res == false {
			c.IndentedJSON(http.StatusUnauthorized, "invalid Bearer Token")
			c.Abort()
			return
		}
		fmt.Println("Bearer Token validated successfully")
		c.Next()
	}
}

func setTokenToExpire(ctx context.Context, client *mongo.Client, token string) {

	existingHash := hashAPIToken(token)

	collection := client.Database("data-dj-main").Collection("apiKeys")
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"hashedToken": existingHash},
		bson.D{
			{"$set", bson.D{{"expiryRequestedDate", time.Now()}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
}

func createTokenHandler(c *gin.Context) {
	newToken, err := createToken(runfig.MongoCtx, runfig.MongoClient)
	if err != nil {
		log.Println("error creating token: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, newToken)
}

func createToken(ctx context.Context, client *mongo.Client) (string, error) {

	newToken := generateAPIToken()
	hashedToken := hashAPIToken(newToken)

	newAPIKey := APIKey{
		HashedToken: hashedToken,
		OwnerID:     1,
		CreatedDate: time.Now().String(),
	}

	result, err := dbutil.InsertOne(ctx, client, "data-dj-main", "apiKeys", newAPIKey)

	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("inserted new apiKey ", result.InsertedID)
	}
	return newToken, err
}

// replaceToken saves & returns a new api token
// the token used in the Authorization is scheduled to be deleted
func replaceToken(c *gin.Context) {
	newToken, err := createToken(runfig.MongoCtx, runfig.MongoClient)
	if err != nil {
		log.Println("error creating token: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	oldToken, err := getTokenFromHeader(c)
	setTokenToExpire(runfig.MongoCtx, runfig.MongoClient, oldToken)

	c.IndentedJSON(http.StatusOK, newToken)
}

//getTokenFromHeader extracts the token part of the
//Authorization field in the Header field
func getTokenFromHeader(c *gin.Context) (string, error) {
	h := authHeader{}
	if err := c.ShouldBindHeader(&h); err != nil {
		return "", errors.New("Must include Authorization header with format `Bearer {token}`")
	}
	parts := strings.Split(h.Authorization, "Bearer ")
	if len(parts) != 2 {
		return "", errors.New("Must include Authorization header with format `Bearer {token}`")
	}
	return strings.TrimSpace(parts[1]), nil
}

//handleValidateAPIToken provides a way to check if an Api Key is valid
func handleValidateAPIToken(c *gin.Context) {

	token, err := getTokenFromHeader(c)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	if token == "" {
		c.IndentedJSON(http.StatusUnauthorized, err.Error())
		return
	}

	res := validateAPIToken(runfig.MongoCtx, runfig.MongoClient, token)
	if res == false {
		c.IndentedJSON(http.StatusUnauthorized, "invalid Bearer Token")
	} else {
		c.IndentedJSON(http.StatusOK, "Authorization Bearer Token validated successfully")
	}
}

func generateAPIToken() string {
	length := 16
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return "sk_" + hex.EncodeToString(b)
}

func hashAPIToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return string(h.Sum(nil))
}

func findToken(ctx context.Context, client *mongo.Client, token string) (APIKey, error) {

	var resultKey APIKey

	hashedToken := hashAPIToken(token)
	collection := client.Database("data-dj-main").Collection("apiKeys")
	err := collection.FindOne(ctx, bson.M{"hashedToken": hashedToken}).Decode(&resultKey)

	return resultKey, err
}

//validateAPIToken hashes the token and checks if it exists in the database
func validateAPIToken(ctx context.Context, client *mongo.Client, token string) bool {

	hashedToken := hashAPIToken(token)
	collection := client.Database("data-dj-main").Collection("apiKeys")
	result := collection.FindOne(ctx, bson.M{"hashedToken": hashedToken})
	err := result.Err()

	noDocs := "ErrNoDocuments"
	if err != nil {
		fmt.Println(noDocs)
		if err.Error() == noDocs {
			return false
		}
		fmt.Println(err.Error())
		return false
	}
	//otherwise token is valid
	return true
}
