package dbutil

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eth-library/dataset-dj/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// CloseMDB is a user defined method to close resources.
// This method closes mongoDB connection and cancel context.
func CloseMDB(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {

	fmt.Println("closing DB")
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

// ConnectToMDB is a user defined method that returns mongo.Client,
// context.Context, context.CancelFunc and error.
// mongo.Client will be used for further database operation.
// context.Context will be used set deadlines for process.
// context.CancelFunc will be used to cancel context and
// resource associtated with it.
func ConnectToMDB(uri string) (*mongo.Client, context.Context, context.CancelFunc, error) {

	// ctx will be used to set deadline for process, here
	// deadline will of 30 seconds.
	ctx, cancel := context.WithCancel(context.Background())

	// mongo.Connect return mongo.Client method
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, cancel, err
}

// PingMDB is a user defined method that accepts
// mongo.Client and context.Context
// This method used to ping the mongoDB, return error if any.
func PingMDB(ctx context.Context, client *mongo.Client) error {

	// mongo.Client has Ping to ping mongoDB, deadline of
	// the Ping method will be determined by cxt
	// Ping method return error if any occurred, then
	// the error can be handled.
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

	return nil
}

// InsertOne is a user defined method, used to insert
// documents into collection returns result of InsertOne
// and error if any.
func InsertOne(ctx context.Context, client *mongo.Client, dbName string, col string,
	doc interface{}) (*mongo.InsertOneResult, error) {

	// select database and collection ith Client.Database method
	// and Database.Collection method
	collection := client.Database(dbName).Collection(col)

	// InsertOne accept two argument of type Context
	// and of empty interface
	result, err := collection.InsertOne(ctx, doc)
	return result, err
}

// InsertMany is a user defined method, used to insert
// documents into collection returns result of
// InsertMany and error if any.
func InsertMany(ctx context.Context, client *mongo.Client, dbName string, col string,
	docs []interface{}) (*mongo.InsertManyResult, error) {

	// select database and collection ith Client.Database
	// method and Database.Collection method
	collection := client.Database(dbName).Collection(col)

	// InsertMany accept two argument of type Context
	// and of empty interface
	result, err := collection.InsertMany(ctx, docs)
	return result, err
}

func AddArchiveToDB(ctx context.Context, client *mongo.Client, dbName string, archive MetaArchive) {
	archiveBSON := archive.ToBSON()
	result, err := InsertOne(ctx, client, dbName, "archives", archiveBSON)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}
	sourceObj := bson.D{{"_id", archive.ID}, {"sources", archive.Sources}}
	result, err = InsertOne(ctx, client, dbName, "sources", sourceObj)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}
}

// FindArchiveInDB retrieves an archive from the MongoDB
func FindArchiveInDB(ctx context.Context, client *mongo.Client, dbName, id string) (MetaArchive, error) {
	var raw MetaArchiveDB
	var archive MetaArchive
	collection := client.Database(dbName).Collection("archives")
	err := collection.FindOne(ctx, bson.D{{"_id", bson.D{{"$eq", id}}}}).Decode(&raw)
	fmt.Println(err)
	archive = raw.Convert()
	archive.ID = id
	return archive, err
}

// UpdateArchiveContent accepts client, context, database, collection, filter and update filter
// and update is of type interface this method returns UpdateResult and an error if any.
func UpdateArchiveContent(ctx context.Context, client *mongo.Client, dbName string,
	id string, contentUpdate interface{}, sourceUpdate interface{}) (*mongo.UpdateResult, error) {

	// Update sources of the archive in "sources" collection used for orders
	sourceCol := client.Database(dbName).Collection("sources")
	_, err := sourceCol.UpdateByID(ctx, id, bson.D{{"$set",
		bson.D{{"sources", sourceUpdate}}}})
	if err != nil {
		log.Println(err)
	}

	// select the database and the collection
	collection := client.Database(dbName).Collection("archives")

	// A single document that match with the
	// filter will get updated.
	// update contains the filed which should get updated.
	result, err := collection.UpdateByID(ctx, id, bson.D{{"$set",
		bson.D{{"content", contentUpdate}}}})
	if err != nil {
		log.Println(err)
		return result, err
	}
	_, err = collection.UpdateByID(ctx, id, bson.D{{"$set",
		bson.D{{"timeUpdated", time.Now().String()}}}})
	if err != nil {
		log.Println(err)
	}
	result, err = collection.UpdateByID(ctx, id, bson.D{{"$set",
		bson.D{{"sources", sourceUpdate}}}})
	return result, err
}

// UpdateArchiveIDs updates the list of archiveIDs in the DB
func UpdateArchiveIDs(ctx context.Context, client *mongo.Client, dbName string, update interface{}) (*mongo.UpdateResult, error) {
	collection := client.Database(dbName).Collection("archiveIDs")

	result, err := collection.UpdateByID(ctx, "id-file", bson.M{"$set": bson.M{"ids": update}})
	return result, err
}

// LoadArchiveIDs retrieves a list of archiveIDs from the database
func LoadArchiveIDs(ctx context.Context, client *mongo.Client, dbName string) (util.Set, error) {
	var idStruct idFileWrapper
	var archiveIDs util.Set

	col := client.Database(dbName).Collection("archiveIDs")
	err := col.FindOne(ctx, bson.D{{"_id", bson.D{{"$eq", "id-file"}}}}).Decode(&idStruct)
	if err != nil {
		if errText := "mongo: no documents in result"; err.Error() == errText {
			emptySlice := make([]string, 0)
			archiveIDs = util.SetFromSlice(emptySlice)
			err = nil
		} else {
			log.Println("LoadArchiveIDs error: ", err)
		}
	}
	archiveIDs = util.SetFromSlice(idStruct.Ids)
	return archiveIDs, err
}

func LoadSourcesByID(ctx context.Context, client *mongo.Client, dbName string, id string) ([]Source, error) {
	var sources []Source
	col := client.Database(dbName).Collection("sources")
	err := col.FindOne(ctx, bson.D{{"$eq", id}}).Decode(&sources)
	return sources, err
}

func LoadOrders(ctx context.Context, client *mongo.Client, dbName string) ([]Order, error) {
	var orders []Order
	var results bson.M
	col := client.Database(dbName).Collection("orders")
	cursor, err := col.Find(ctx, bson.D{})
	if err != nil {
		log.Println("LoadOrders error: ", err)
	}
	if err = cursor.All(ctx, &results); err != nil {
		log.Println("LoadOrders error: ", err)
	}
	for _, res := range results {
		orders = append(orders, res.(Order))
	}
	return orders, err
}
