package dbutil

import (
	"context"
	"fmt"

	"github.com/eth-library-lab/dataset-dj/datastructs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// metaArchives are the blueprints for the zip archives that will be created once the user initiates
// the download process. Files is implemented as a set in order to avoid duplicate files within a
// metaArchive
// type metaArchive struct {
// 	ID    string `json:"id"`
// 	Files set    `json:"files"`
// }
type MetaArchive struct {
	ID          string          `json:"id"`
	Files       datastructs.Set `json:"files"`
	TimeCreated string          `json:"timeCreated"`
	TimeUpdated string          `json:"timeUpdated"`
	Status      string          `json:"status"`
	Source      string          `json:"source"`
}

// converts the meta archive to binary JSON format
func (a MetaArchive) ToBSON() bson.D {
	var files bson.A
	for _, v := range a.Files.ToSlice() {
		files = append(files, v)
	}
	return bson.D{primitive.E{Key: "_id", Value: a.ID}, primitive.E{Key: "files", Value: files}}
}

type SourceBucket struct {
	BucketID       string `json:"ID"`
	BucketURL      string
	BucketName     string
	BucketPrefixes []string
	BucketOrigin   string
	Description    string
	Owner          string
}

func (sb SourceBucket) ToBSON() bson.D {
	var prefixes bson.A
	for _, v := range sb.BucketPrefixes {
		prefixes = append(prefixes, v)
	}
	res := bson.D{primitive.E{Key: "_id", Value: sb.BucketID},
		primitive.E{Key: "URL", Value: sb.BucketURL},
		primitive.E{Key: "Name", Value: sb.BucketName},
		primitive.E{Key: "Prefixes", Value: prefixes},
		primitive.E{Key: "Origin", Value: sb.BucketOrigin},
		primitive.E{Key: "Description", Value: sb.Description},
		primitive.E{Key: "Owner", Value: sb.Owner}}

	return res
}

func BucketMapfromSlice(slice []SourceBucket) map[string]SourceBucket {
	bucketMap := make(map[string]SourceBucket)
	for _, b := range slice {
		bucketMap[b.BucketURL+b.BucketName] = b
	}
	return bucketMap
}

type bucketFileWrapper struct {
	_id     string         `json:"id"`
	buckets []SourceBucket `json:"buckets"`
}

type idFileWrapper struct {
	_id string   `json:"id"`
	Ids []string `json:"ids"`
}

// This is a user defined method to close resources.
// This method closes mongoDB connection and cancel context.
func CloseMDB(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {

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

func ConnectToMDB(uri string) (*mongo.Client, context.Context, context.CancelFunc, error) {

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
func PingMDB(client *mongo.Client, ctx context.Context) error {

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
func InsertOne(client *mongo.Client, ctx context.Context, dataBase, col string,
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
func InsertMany(client *mongo.Client, ctx context.Context, dataBase, col string,
	docs []interface{}) (*mongo.InsertManyResult, error) {

	// select database and collection ith Client.Database
	// method and Database.Collection method
	collection := client.Database(dataBase).Collection(col)

	// InsertMany accept two argument of type Context
	// and of empty interface
	result, err := collection.InsertMany(ctx, docs)
	return result, err
}

func NewMetaArchiveInDB(client *mongo.Client, ctx context.Context, id string, files []string) MetaArchive {
	// Create new metaArchive with random UID
	archive := MetaArchive{ID: id, Files: datastructs.SetFromSlice(files)}
	AddArchiveToDB(client, ctx, archive)
	return archive
}

func AddArchiveToDB(client *mongo.Client, ctx context.Context, archive MetaArchive) {
	archiveBSON := archive.ToBSON()
	result, err := InsertOne(client, ctx, "data-dj-main", "archives", archiveBSON)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}
}

// funciton for retrieving an archive from the MongoDB
func FindArchiveInDB(client *mongo.Client, ctx context.Context, id string) (MetaArchive, error) {
	var archive MetaArchive
	collection := client.Database("data-dj-main").Collection("archives")
	err := collection.FindOne(ctx, bson.D{{"_id", bson.D{{"$eq", id}}}}).Decode(&archive)
	archive.ID = id
	return archive, err
}

// This methods accepts client, context, database, collection, filter and update filter
// and update is of type interface this method returns UpdateResult and an error if any.
func UpdateFilesOfArchive(client *mongo.Client, ctx context.Context, id string, update interface{}) (*mongo.UpdateResult, error) {

	// select the database and the collection
	collection := client.Database("data-dj-main").Collection("archives")

	// A single document that match with the
	// filter will get updated.
	// update contains the filed which should get updated.
	result, err := collection.UpdateByID(ctx, id, bson.D{{"$set", bson.D{{"files", update}}}})
	return result, err
}

// function that updates the list of archiveIDs
func UpdateArchiveIDs(client *mongo.Client, ctx context.Context, update interface{}) (*mongo.UpdateResult, error) {
	collection := client.Database("data-dj-main").Collection("archiveIDs")

	result, err := collection.UpdateByID(ctx, "id-file", bson.M{"$set": bson.M{"ids": update}})
	return result, err
}

// function for retrieving list of archiveIDs
func LoadArchiveIDs(client *mongo.Client, ctx context.Context) (datastructs.Set, error) {
	var idStruct idFileWrapper
	col := client.Database("data-dj-main").Collection("archiveIDs")
	err := col.FindOne(ctx, bson.D{{"_id", bson.D{{"$eq", "id-file"}}}}).Decode(&idStruct)
	return datastructs.SetFromSlice(idStruct.Ids), err
}

// function for retrieving list of archiveIDs
func LoadSourceBuckets(client *mongo.Client, ctx context.Context) ([]SourceBucket, error) {
	var sourceStruct bucketFileWrapper
	col := client.Database("data-dj-main").Collection("sourceBuckets")
	err := col.FindOne(ctx, bson.D{{"_id", bson.D{{"$eq", "bucket-file"}}}}).Decode(&sourceStruct)
	return sourceStruct.buckets, err
}

// function that updates the list of archiveIDs
func UpdateSourceBuckets(client *mongo.Client, ctx context.Context, update interface{}) (*mongo.UpdateResult, error) {
	collection := client.Database("data-dj-main").Collection("sourceBuckets")

	result, err := collection.UpdateByID(ctx, "bucket-file", bson.M{"$set": bson.M{"buckets": update}})
	return result, err
}
