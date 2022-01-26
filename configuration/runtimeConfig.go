package configuration

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
	"github.com/eth-library-lab/dataset-dj/datastructs"
	dbutil "github.com/eth-library-lab/dataset-dj/dbutil"
	"github.com/eth-library-lab/dataset-dj/redisutil"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
)

//RuntimeConfig holds pointers to storage clients and some in memory lists
type RuntimeConfig struct {
	StorageClient    *storage.Client // client used to connect to the storage in order to read and write files
	RdbClient        *redis.Client
	MongoClient      *mongo.Client
	MongoCtx         context.Context
	CtxCancel        context.CancelFunc
	ArchiveIDs       datastructs.Set
	SourceBucketList []dbutil.SourceBucket
	SourceBuckets    map[string]dbutil.SourceBucket
}

func InitRuntimeConfig(sc *ServerConfig) *RuntimeConfig {
	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer storageClient.Close()

	// connect to redis instance
	rdbClient := redisutil.InitRedisConnection(sc.RdbAddr)

	// Get Client, Context, CalcelFunc and
	// err from connect method.
	// "mongodb+srv://data-dj:LibLab123@archive-cluster.jzmhu.mongodb.net/data-dj-main?retryWrites=true&w=majority"
	mongoClient, mongoCtx, cancel, err := dbutil.ConnectToMDB(sc.DbConnector)

	if err != nil {
		panic(err)
	}

	// Ping mongoDB with Ping method
	err = dbutil.PingMDB(ctx, mongoClient)
	if err != nil {
		fmt.Println("error PingMDB: ", err)
	}

	// Load the list of already used archiveIDs when redeploying
	archiveIDs, err := dbutil.LoadArchiveIDs(mongoCtx, mongoClient, sc.DbName)
	if err != nil {
		log.Fatal(err)
	}

	sourceBucketList, err := dbutil.LoadSourceBuckets(mongoCtx, mongoClient, sc.DbName)
	sourceBuckets := dbutil.BucketMapfromSlice(sourceBucketList)
	if err != nil {
		if sc.Mode != "test" {
			log.Println("WARNING: no sourceBucketList found: ", err)
		}
		// init as empty slices and maps to avoid nil pointer errors
		sourceBucketList = []dbutil.SourceBucket{}
		sourceBuckets = map[string]dbutil.SourceBucket{}
	}

	rc := RuntimeConfig{
		StorageClient:    storageClient,
		RdbClient:        rdbClient,
		MongoClient:      mongoClient,
		MongoCtx:         mongoCtx,
		CtxCancel:        cancel,
		ArchiveIDs:       archiveIDs,
		SourceBucketList: sourceBucketList,
		SourceBuckets:    sourceBuckets}

	return &rc

}
