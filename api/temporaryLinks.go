package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	redisutil "github.com/eth-library/dataset-dj/redisutil"
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
	Email      string    `bson:"email"`
}

// EmailRequestBody for binding email field in a json body
type EmailRequestBody struct {
	Email string `json:"email"`
}

// EmailParts required by the mailHandler to send an email
type EmailParts struct {
	To       string
	Subject  string
	BodyType string // e.g.: text/plain
	Body     string
}

func handleCreateLink(c *gin.Context) {

	var emailRequestBody EmailRequestBody
	if err := c.BindJSON(&emailRequestBody); err != nil {
		c.IndentedJSON(http.StatusBadRequest, "recipient email required")
		return
	}
	email := emailRequestBody.Email
	email, err := emailIsValid(email)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, "email format not valid")
		return
	}

	//to prevent duplicates
	err = deleteExistingLinks(runtime.MongoCtx, runtime.MongoClient, email)
	if err != nil {
		log.Println("error deleting existing links: ", err)
	}
	//create link
	linkID := createSingleUseLink(runtime.MongoCtx, runtime.MongoClient, email)
	url := c.Request.Host + "/key/claim/" + linkID

	//TO DO: send email to recipient instead of return link
	err = publishAPILinkEmailTask(url, email)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "error encountered while sending email")
		return
	}
	c.IndentedJSON(http.StatusCreated, "email with token link sent")
}

func publishAPILinkEmailTask(url string, recipientEmail string) error {

	content := fmt.Sprintf(`
	<h1>Welcome to the Data DJ</h1>

	<p>Thanks for joining the Data DJ!</p>
	
	<p>
	Below is a <em>single-use</em> link that returns an API Key.
	This API Key allows your application to authentic with the Data DJ API.
	<br/>
	<br/>
	The Key can only be viewed once. It should be saved somewhere securely (e.g. password manager), not disclosed to users or client side code, and should not be hardcoded or checked into repositories.</p>
	</p>
	<p>
	   <a href="%v" target="_">%v</a> <br/>
	   (click or copy & paste into your browser)
	</p>
	
	In case of issues, please contact us at contact[at]librarylab.ethz.ch
	`, url, url)

	emailparts := EmailParts{
		To:       recipientEmail,
		Subject:  "DataDJ - Link to claim API Key",
		BodyType: "text/html",
		Body:     content,
	}

	err := redisutil.PublishTask(runtime.RdbClient, emailparts, "emails")
	if err != nil {
		log.Println("ERROR while publishing email task:", err.Error())
		return err
	}
	return nil
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
		// bson.D{
		// 	{"$set", bson.M{"used": true, "expireAt": time.Now()}},
		// },
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
