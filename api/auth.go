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
	"os"
	"strings"
	"time"

	"github.com/eth-library/dataset-dj/dbutil"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type authHeader struct {
	Authorization string `header:"Authorization"`
}

// APIKey is the format of the mongodb document that stores the keys
type APIKey struct {
	HashedToken string `bson:"hashedToken,omitempty"`
	CreatedDate string `bson:"createdDate,omitempty"`
	Permission  string `bson:"permission,omitempty"` //service, handler, user or admin
}

type tokenResponse struct {
	APIKey  string
	Message string
}

type tokenRequest struct {
	APIKey string `json:"apiKey"`
}

type dbCollection interface {
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
}

// colHandle returns a reference to a mongodb collection handle
func getColHandle(collectionName string) *mongo.Collection {
	db := runtime.MongoClient.Database(config.DbName)
	collection := db.Collection(collectionName)
	return collection
}

// AuthMiddleware validates the bearer token before
// allowing the handler to be a called
func AuthMiddleware(requiredPermission string) gin.HandlerFunc {

	return func(c *gin.Context) {
		token, err := getTokenFromHeader(c)
		if err != nil {
			c.IndentedJSON(http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}
		if token == "" {
			c.IndentedJSON(http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		collection := getColHandle("apiKeys")
		res, tokenPermission := validateAPIToken(collection, token)

		if res == false {
			c.IndentedJSON(http.StatusUnauthorized, "invalid Bearer Token")
			c.Abort()
			return
		}
		if requiredPermission == tokenPermission {
			log.Println("Bearer Token validated successfully")
			c.Next()
			return
		}
		c.IndentedJSON(http.StatusUnauthorized, "Insufficient Token Permission for Request")
		c.Abort()
		return

	}
}

func setupAPIToken(c *gin.Context, tokenPermission string) {
	token, err := CreateToken(runtime.MongoCtx, runtime.MongoClient, tokenPermission)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "unable to create APIKey for Taskhandler")
		return
	}

	resp := tokenResponse{
		APIKey:  token,
		Message: "store this API Key securely. It cannot be retrieved again. Do not disclose it to anyone. Use the /key/replace endpoint to replace this key periodically or if it is compromised",
	}
	c.IndentedJSON(http.StatusOK, resp)
}

func deleteToken(col dbCollection, token string) error {

	existingHash := hashAPIToken(token)

	result, err := col.DeleteMany(
		runtime.MongoCtx,
		bson.M{"hashedToken": existingHash},
	)
	if err != nil {
		log.Fatal(err)
		return fmt.Errorf("encountered error while deleting apiKey")
	}

	if result.DeletedCount == 0 {
		if err == nil {
			return fmt.Errorf("apiKey not found")
		}
		return fmt.Errorf("encountered error while deleting apiKey")
	}
	log.Printf("Deleted %v apiKeys\n", result.DeletedCount)

	return nil
}

func setTokenToExpire(ctx context.Context, client *mongo.Client, token string) {

	existingHash := hashAPIToken(token)

	collection := client.Database(config.DbName).Collection("apiKeys")
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

// initAdminToken generates a random api key and inserts it in the database
func initAdminToken(ctx context.Context, client *mongo.Client) error {

	// check if already existing
	col := client.Database(config.DbName).Collection("apiKeys")
	result := col.FindOne(ctx, bson.M{"permission": "admin"})
	if result.Err() == nil {
		if config.Mode != "test" {
			log.Println("using existing admin token")
		}
		return nil
	}
	if result.Err() != nil {
		// no existing admin token
		// save one if not
		token := os.Getenv("ADMIN_KEY")
		hashedToken := hashAPIToken(token)

		newAPIKey := APIKey{
			HashedToken: hashedToken,
			Permission:  "admin",
			CreatedDate: time.Now().String(),
		}
		_, err := dbutil.InsertOne(ctx, client, config.DbName, "apiKeys", newAPIKey)
		if err != nil {
			log.Println(err)
		}
		if config.Mode != "test" {
			log.Println("inserted admin token from environment")
		}
		return nil
	}
	return fmt.Errorf("could not find or save admin token")
}

// CreateToken generates a random api key and inserts it in the database
func CreateToken(ctx context.Context, client *mongo.Client, permission string) (string, error) {

	newToken := generateAPIToken(permission)
	hashedToken := hashAPIToken(newToken)

	newAPIKey := APIKey{
		HashedToken: hashedToken,
		Permission:  permission,
		CreatedDate: time.Now().String(),
	}

	_, err := dbutil.InsertOne(ctx, client, config.DbName, "apiKeys", newAPIKey)

	if err != nil {
		log.Println(err)
	}
	return newToken, err
}

// replaceToken saves & returns a new api token
// the token used in the Authorization is scheduled to be deleted
func replaceToken(c *gin.Context) {
	tokenPermission := "service"
	newToken, err := CreateToken(runtime.MongoCtx, runtime.MongoClient, tokenPermission)
	if err != nil {
		log.Println("error creating token: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	oldToken, err := getTokenFromHeader(c)

	collection := getColHandle("apiKeys")
	err = deleteToken(collection, oldToken)
	if err != nil {
		log.Println("error deleting token: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	resp := tokenResponse{
		APIKey:  newToken,
		Message: "store this new API Key securely. It cannot be retrieved again. Do not disclose it to anyone. Use the /key/replace endpoint to replace this key periodically or if it is compromised",
	}
	c.IndentedJSON(http.StatusOK, resp)
}

// revokeToken revokes an apiKey by deleting the key from the database. It will be revoked immediately.
func revokeToken(c *gin.Context) {

	var token tokenRequest

	err := c.ShouldBindJSON(&token)
	if err != nil {
		log.Println("revokeToken request error: ", err)
		c.JSON(http.StatusBadRequest, "must include key to revoke in request body e.d. {'apiKey':'sk_123456'}")
		return
	}

	// setTokenToExpire(runtime.MongoCtx, runtime.MongoClient, oldToken)
	col := getColHandle("apiKeys")
	err = deleteToken(col, token.APIKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	//make sure that the token is no longer valid
	res, permission := validateAPIToken(col, token.APIKey)
	if (res != false) || (permission != "") {
		log.Println("error deleting token: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	resp := fmt.Sprintf("revoked key: %v", token.APIKey)
	c.IndentedJSON(http.StatusOK, resp)
}

// getTokenFromHeader extracts the token part of the
// Authorization field in the Header field
func getTokenFromHeader(c *gin.Context) (string, error) {
	h := authHeader{}
	if err := c.ShouldBindHeader(&h); err != nil {
		return "", errors.New("must include Authorization header with format `Bearer {token}`")
	}
	parts := strings.Split(h.Authorization, "Bearer ")
	if len(parts) != 2 {
		return "", errors.New("must include Authorization header with format `Bearer {token}`")
	}
	return strings.TrimSpace(parts[1]), nil
}

func getTokenPrefix(permission string) (string, error) {

	if permission == "admin" {
		return "ak_", nil
	}
	if permission == "service" {
		return "sk_", nil
	}
	if permission == "handler" {
		return "hk_", nil
	}
	return "", fmt.Errorf("permission must be one of [admin, service, handler]")

}

func generateAPIToken(permission string) string {
	length := 16
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	prefix, err := getTokenPrefix(permission)
	if err != nil {
		log.Println(err)
		return ""
	}
	return prefix + hex.EncodeToString(b)
}

func hashAPIToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

func findToken(ctx context.Context, client *mongo.Client, token string) (APIKey, error) {

	var resultKey APIKey

	hashedToken := hashAPIToken(token)
	collection := client.Database(config.DbName).Collection("apiKeys")
	err := collection.FindOne(ctx, bson.M{"hashedToken": hashedToken}).Decode(&resultKey)

	return resultKey, err
}

// validateAPIToken hashes the token, checks if it exists in the database and returns the token's permission tag
func validateAPIToken(collection dbCollection, token string) (bool, string) {

	hashedToken := hashAPIToken(token)

	result := collection.FindOne(runtime.MongoCtx, bson.M{"hashedToken": hashedToken})
	err := result.Err()

	noDocs := "ErrNoDocuments"
	if err != nil {
		fmt.Println(noDocs)
		if err.Error() == noDocs {
			return false, ""
		}
		log.Println("validateAPIToken failed:", err.Error())
		return false, ""
	}

	//otherwise token is valid
	var key APIKey
	result.Decode(&key)

	return true, key.Permission
}
