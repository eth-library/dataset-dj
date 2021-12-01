package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type idWrapper struct {
	_id string   `json:"id"`
	Ids []string `json:"ids"`
}

// This is a user defined method to close resources.
// This method closes mongoDB connection and cancel context.
func closeMDB(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {

	// CancelFunc to cancel to context
	defer cancel()

	// client provides a method to close
	// a mongoDB connection.
	defer func() {

		// client.Disconnect method also has deadline.
		// returns error if any,
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

// This is a user defined method that returns mongo.Client,
// context.Context, context.CancelFunc and error.
// mongo.Client will be used for further database operation.
// context.Context will be used set deadlines for process.
// context.CancelFunc will be used to cancel context and
// resource associtated with it.

func connectToMDB(uri string) (*mongo.Client, context.Context, context.CancelFunc, error) {

	// ctx will be used to set deadline for process, here
	// deadline will of 30 seconds.
	ctx, cancel := context.WithCancel(context.Background())

	// mongo.Connect return mongo.Client method
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, cancel, err
}

// This is a user defined method that accepts
// mongo.Client and context.Context
// This method used to ping the mongoDB, return error if any.
func pingMDB(client *mongo.Client, ctx context.Context) error {

	// mongo.Client has Ping to ping mongoDB, deadline of
	// the Ping method will be determined by cxt
	// Ping method return error if any occurred, then
	// the error can be handled.
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	fmt.Println("connected successfully")
	return nil
}

// insertOne is a user defined method, used to insert
// documents into collection returns result of InsertOne
// and error if any.
func insertOne(client *mongo.Client, ctx context.Context, dataBase, col string,
	doc interface{}) (*mongo.InsertOneResult, error) {

	// select database and collection ith Client.Database method
	// and Database.Collection method
	collection := client.Database(dataBase).Collection(col)

	// InsertOne accept two argument of type Context
	// and of empty interface
	result, err := collection.InsertOne(ctx, doc)
	return result, err
}

// insertMany is a user defined method, used to insert
// documents into collection returns result of
// InsertMany and error if any.
func insertMany(client *mongo.Client, ctx context.Context, dataBase, col string,
	docs []interface{}) (*mongo.InsertManyResult, error) {

	// select database and collection ith Client.Database
	// method and Database.Collection method
	collection := client.Database(dataBase).Collection(col)

	// InsertMany accept two argument of type Context
	// and of empty interface
	result, err := collection.InsertMany(ctx, docs)
	return result, err
}

func newMetaArchiveInDB(files []string) metaArchive {
	// Create new metaArchive with random UID
	archive := metaArchive{ID: generateToken(), Files: files}
	archiveBSON := archive.toBSON()
	result, err := insertOne(mongoClient, mongoCtx, "data-dj-main", "archives", archiveBSON)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}
	return archive
}

// funciton for retrieving an archive from the MongoDB
func findArchiveInDB(id string) (metaArchive, error) {
	var archive metaArchive
	collection := mongoClient.Database("data-dj-main").Collection("archives")
	err := collection.FindOne(mongoCtx, bson.D{{"_id", bson.D{{"$eq", id}}}}).Decode(&archive)
	archive.ID = id
	return archive, err
}

// This methods accepts client, context, database, collection, filter and update filter
// and update is of type interface this method returns UpdateResult and an error if any.
func updateFilesOfArchive(id string, update interface{}) (*mongo.UpdateResult, error) {

	// select the database and the collection
	collection := mongoClient.Database("data-dj-main").Collection("archives")

	// A single document that match with the
	// filter will get updated.
	// update contains the filed which should get updated.
	result, err := collection.UpdateByID(mongoCtx, id, bson.D{{"$set", bson.D{{"files", update}}}})
	return result, err
}

// function that updates the list of archiveIDs
func updateArchiveIDs(update interface{}) (*mongo.UpdateResult, error) {
	collection := mongoClient.Database("data-dj-main").Collection("archiveIDs")

	result, err := collection.UpdateByID(mongoCtx, "id-file", bson.M{"$set": bson.M{"ids": update}})
	return result, err
}

// function for retrieving list of archiveIDs
func loadArchiveIDs() (set, error) {
	var idStruct idWrapper
	col := mongoClient.Database("data-dj-main").Collection("archiveIDs")
	err := col.FindOne(mongoCtx, bson.D{{"_id", bson.D{{"$eq", "id-file"}}}}).Decode(&idStruct)
	return setFromSlice(idStruct.Ids), err
}
