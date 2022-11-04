package dbutil

import (
	"context"
	"fmt"
	"log"

	"github.com/eth-library/dataset-dj/datastructs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//	type metaArchive struct {
//		ID    string `json:"id"`
//		Files set    `json:"files"`
//	}
type MetaArchiveRaw struct {
	ID          string   `json:"id"`
	Files       []string `json:"files"`
	TimeCreated string   `json:"timeCreated"`
	TimeUpdated string   `json:"timeUpdated"`
	Status      string   `json:"status"`
	Source      string   `json:"source"`
}

// MetaArchive is the blueprint for the zip archives that will be created once the user initiates
// the download process. Files is implemented as a set in order to avoid duplicate files within a
// metaArchive
type MetaArchive struct {
	ID          string          `json:"id"`
	Files       datastructs.Set `json:"files"`
	TimeCreated string          `json:"timeCreated"`
	TimeUpdated string          `json:"timeUpdated"`
	Status      string          `json:"status"`
	Source      string          `json:"source"`
}

func (raw MetaArchiveRaw) convert() MetaArchive {
	var a MetaArchive
	a.ID = raw.ID
	a.Files = datastructs.SetFromSlice(raw.Files)
	a.TimeCreated = raw.TimeCreated
	a.TimeUpdated = raw.TimeUpdated
	a.Status = raw.Status
	a.Source = raw.Source

	return a
}

// ToBSON converts the meta archive to binary JSON format
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

func NewMetaArchiveInDB(ctx context.Context, client *mongo.Client, dbName string, id string, files []string) MetaArchive {
	// Create new metaArchive with random UID
	archive := MetaArchive{ID: id, Files: datastructs.SetFromSlice(files)}
	AddArchiveToDB(ctx, client, dbName, archive)
	return archive
}

func AddArchiveToDB(ctx context.Context, client *mongo.Client, dbName string, archive MetaArchive) {
	archiveBSON := archive.ToBSON()
	result, err := InsertOne(ctx, client, dbName, "archives", archiveBSON)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}
}

// FindArchiveInDB retrieves an archive from the MongoDB
func FindArchiveInDB(ctx context.Context, client *mongo.Client, dbName, id string) (MetaArchive, error) {
	var raw MetaArchiveRaw
	var archive MetaArchive
	collection := client.Database(dbName).Collection("archives")
	err := collection.FindOne(ctx, bson.D{{"_id", bson.D{{"$eq", id}}}}).Decode(&raw)
	fmt.Println(err)
	archive = raw.convert()
	archive.ID = id
	return archive, err
}

// UpdateFilesOfArchive accepts client, context, database, collection, filter and update filter
// and update is of type interface this method returns UpdateResult and an error if any.
func UpdateFilesOfArchive(ctx context.Context, client *mongo.Client, dbName string, id string, update interface{}) (*mongo.UpdateResult, error) {

	// select the database and the collection
	collection := client.Database(dbName).Collection("archives")

	// A single document that match with the
	// filter will get updated.
	// update contains the filed which should get updated.
	result, err := collection.UpdateByID(ctx, id, bson.D{{"$set", bson.D{{"files", update}}}})
	return result, err
}

// UpdateArchiveIDs updates the list of archiveIDs in the DB
func UpdateArchiveIDs(ctx context.Context, client *mongo.Client, dbName string, update interface{}) (*mongo.UpdateResult, error) {
	collection := client.Database(dbName).Collection("archiveIDs")

	result, err := collection.UpdateByID(ctx, "id-file", bson.M{"$set": bson.M{"ids": update}})
	return result, err
}

// LoadArchiveIDs retrieves a list of archiveIDs from the database
func LoadArchiveIDs(ctx context.Context, client *mongo.Client, dbName string) (datastructs.Set, error) {
	var idStruct idFileWrapper
	var archiveIDs datastructs.Set

	col := client.Database(dbName).Collection("archiveIDs")
	err := col.FindOne(ctx, bson.D{{"_id", bson.D{{"$eq", "id-file"}}}}).Decode(&idStruct)
	if err != nil {
		if errText := "mongo: no documents in result"; err.Error() == errText {
			emptySlice := make([]string, 0)
			archiveIDs = datastructs.SetFromSlice(emptySlice)
			err = nil
		} else {
			log.Println("LoadArchiveIDs error: ", err)
		}
	}
	archiveIDs = datastructs.SetFromSlice(idStruct.Ids)
	return archiveIDs, err
}

// LoadSourceBuckets retrieves a list of archiveIDs from the db
func LoadSourceBuckets(ctx context.Context, client *mongo.Client, dbName string) ([]SourceBucket, error) {
	var sourceStruct bucketFileWrapper
	col := client.Database(dbName).Collection("sourceBuckets")
	err := col.FindOne(ctx, bson.D{{"_id", bson.D{{"$eq", "bucket-file"}}}}).Decode(&sourceStruct)
	return sourceStruct.buckets, err
}

// UpdateSourceBuckets updates the list of archiveIDs in the DB
func UpdateSourceBuckets(ctx context.Context, client *mongo.Client, dbName string, update interface{}) (*mongo.UpdateResult, error) {
	collection := client.Database(dbName).Collection("sourceBuckets")

	result, err := collection.UpdateByID(ctx, "bucket-file", bson.M{"$set": bson.M{"buckets": update}})
	return result, err
}
