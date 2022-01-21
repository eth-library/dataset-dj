package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type singleUseLink struct {
	// ID         string    `bson:"_id"`        // will use the _id returned from Mongo to create a link
	Used       bool      `bson:"used"`       // set to True when visited. Then delete this document
	ExpireAt   time.Time `bson:"expireAt"`   //after this time MongoDB will expire this document
	Permission string    `bson:"permission"` //purpose that this link has e.g. create createAPIToken
	OwnerID    int       `bson:"ownerID"`
}

type tokenResponse struct {
	APIKey  string
	Message string
}

func handleCreateLink(c *gin.Context) {
	id := createSingleUseLink(runfig.MongoCtx, runfig.MongoClient)
	url := c.Request.Host + "/key/claim/" + id
	c.IndentedJSON(http.StatusCreated, url)
}

func createSingleUseLink(ctx context.Context, client *mongo.Client) string {

	var signUpLink singleUseLink

	signUpLink = singleUseLink{
		Used:       false,
		ExpireAt:   time.Now().Add(time.Second * 120), //change this to time.hours in production
		Permission: "createAPIToken",
		OwnerID:    1,
	}
	collection := client.Database("data-dj-main").Collection("temporaryLinks")
	result, err := collection.InsertOne(ctx, signUpLink)
	if err != nil {
		fmt.Println(err)
	}

	oid, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		return oid.Hex()
	}
	return ""
}

func validateTokenLink(ctx context.Context, client *mongo.Client, linkID string) (bool, error) {
	id, _ := primitive.ObjectIDFromHex(linkID)
	var link singleUseLink
	collection := client.Database("data-dj-main").Collection("temporaryLinks")
	result := collection.FindOne(
		ctx,
		bson.M{"_id": id},
	)
	result.Decode(&link)

	if result.Err() != nil {
		fmt.Println(result.Err().Error())
		if result.Err().Error() == "mongo: no documents in result" {
			return false, nil
		}
		return false, result.Err()
	}
	if link.Used == true {
		return false, fmt.Errorf("link already used")
	}
	return true, nil
}



func expireLink(ctx context.Context, client *mongo.Client, linkID string) error {

	id, _ := primitive.ObjectIDFromHex(linkID)
	collection := client.Database("data-dj-main").Collection("temporaryLinks")
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.D{
			{"$set", bson.M{"used": true, "expireAt": time.Now().Add(10)}},
		},
	)
	fmt.Println(result)
	return err
}
