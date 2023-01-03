package main

import (
	"context"
	"fmt"
	"github.com/eth-library/dataset-dj/mailHandler"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type singleUseLink struct {
	// ID         string    `bson:"_id"`     // will use the _id returned from Mongo to create a link
	Used       bool      `bson:"used"`       // set to True when visited. Then delete this document
	ExpireAt   time.Time `bson:"expireAt"`   //after this time MongoDB will expire this document
	Permission string    `bson:"permission"` //purpose that this link has e.g. create createAPIToken
	OwnerID    int       `bson:"ownerID"`
	Email      string    `bson:"email"`
}

// EmailRequestBody for binding email field in a json body
type EmailRequestBody struct {
	Email string `json:"email"`
}

func startAPILinkEmailTask(url string, recipientEmail string) {

	content := fmt.Sprintf(mailHandler.APILinkContent, url, url)

	emailParts := mailHandler.EmailParts{
		To:         recipientEmail,
		Subject:    "DataDJ - Link to claim API Key",
		BodyType:   "text/html",
		Body:       content,
		Server:     config.ServiceEmailHost,
		Address:    config.ServiceEmailAddress,
		Password:   config.ServiceEmailPassword,
		ErrorMsg:   "an error occurred while sending the API Token email notification: ",
		SuccessMsg: "email with token link sent",
	}

	go mailHandler.SendEmailAsync(emailParts)
}

func createSingleUseLink(ctx context.Context, client *mongo.Client, email string) string {

	var signUpLink singleUseLink

	signUpLink = singleUseLink{
		Used:       false,
		ExpireAt:   time.Now().Add(time.Second * 120), //change this to time.hours in production
		Permission: "createAPIToken",
		Email:      email,
	}
	collection := client.Database(config.DbName).Collection("temporaryLinks")
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
	collection := client.Database(config.DbName).Collection("temporaryLinks")
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
		fmt.Println("link already used")
		return false, nil
	}
	return true, nil
}

// deleteExistingLinks deletes any existing temporaryLinks assosciated with `email`
// prevents possibility of individual users having multiple valid links
func deleteExistingLinks(ctx context.Context, client *mongo.Client, email string) error {

	collection := client.Database(config.DbName).Collection("temporaryLinks")
	result, err := collection.DeleteMany(
		ctx,
		bson.M{"email": email},
	)
	fmt.Printf("deleted %v existing temp links\n", result.DeletedCount)
	return err
}

func expireLink(ctx context.Context, client *mongo.Client, linkID string) error {

	id, _ := primitive.ObjectIDFromHex(linkID)
	collection := client.Database(config.DbName).Collection("temporaryLinks")
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.D{
			{"$set", bson.M{"used": true, "expireAt": time.Now()}},
		},
	)
	fmt.Println(result)
	return err
}
