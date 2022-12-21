package configuration

import (
	"context"
	"fmt"
	"github.com/eth-library/dataset-dj/constants"
	"log"

	"cloud.google.com/go/storage"
	"github.com/eth-library/dataset-dj/dbutil"
	"github.com/eth-library/dataset-dj/util"
	"go.mongodb.org/mongo-driver/mongo"
)

// ApiRuntime holds pointers to storage clients and some in memory lists
type ApiRuntime struct {
	MongoClient *mongo.Client
	MongoCtx    context.Context
	CtxCancel   context.CancelFunc
	ArchiveIDs  util.Set
	SourceIDs   util.Set
	OrderIDs    util.Set
}

func InitApiRuntime(sc *ApiConfig) *ApiRuntime {
	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer storageClient.Close()

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
	archiveIDs, err := dbutil.LoadIDs(mongoCtx, mongoClient, sc.DbName, constants.ArchiveIDs)
	if err != nil {
		log.Fatal(err)
	}
	// Load the list of already used sourceIDs when redeploying
	sourceIDs, err := dbutil.LoadIDs(mongoCtx, mongoClient, sc.DbName, constants.SourceIDs)
	if err != nil {
		log.Fatal(err)
	}
	// Load the list of already used orderIDs when redeploying
	orderIDs, err := dbutil.LoadIDs(mongoCtx, mongoClient, sc.DbName, constants.OrderIDs)
	if err != nil {
		log.Fatal(err)
	}

	rc := ApiRuntime{
		MongoClient: mongoClient,
		MongoCtx:    mongoCtx,
		CtxCancel:   cancel,
		ArchiveIDs:  archiveIDs,
		SourceIDs:   sourceIDs,
		OrderIDs:    orderIDs,
	}

	return &rc

}
