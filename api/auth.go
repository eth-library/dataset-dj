package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"github.com/eth-library-lab/dataset-dj/dbutil"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

//APIKey is the format of the mongodb document that stores the keys
type APIKey struct {
	HashedToken string `bson:"hashedToken,omitempty"`
	OwnerID     int32  `bson:"ownerId,omitempty"`
}

func generateAPIToken(length int) string {

	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func hashAPIToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return string(h.Sum(nil))
}

// handleTokenRequest generates and returns a new api token
func handleTokenRequest(c *gin.Context) {
	token := "sk_" + generateAPIToken(16)
	hashedToken := hashAPIToken(token)

	newAPIKey := APIKey{
		HashedToken: hashedToken,
		OwnerID:     1,
	}

	result, err := dbutil.InsertOne(runfig.MongoCtx, runfig.MongoClient, "data-dj-main", "apiKeys", newAPIKey)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(result)
	}

	c.IndentedJSON(http.StatusOK, token)
}

//validateAPIToken hashes the token and checks if it exists in the database
func validateAPIToken(ctx context.Context, client *mongo.Client, token string) bool {

	var resultKey APIKey

	hashedToken := hashAPIToken(token)
	collection := client.Database("data-dj-main").Collection("apiKeys")
	err := collection.FindOne(ctx, bson.M{"hashedToken": hashedToken}).Decode(&resultKey)

	noDocs := "ErrNoDocuments"
	if err != nil {
		if err.Error() == noDocs {
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}
