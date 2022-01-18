package main

import (
	"encoding/json"
	"fmt"

	"github.com/eth-library-lab/dataset-dj/dbutil"
	"go.mongodb.org/mongo-driver/bson"
)

// handleArchiveMessage handles a message received in the 'sourceBuckets' channel
func handleSourceBucketMessage(messagePayload string) {
	var bucket dbutil.SourceBucket

	// convert json string into struct
	json.Unmarshal([]byte(messagePayload), &bucket)

	fmt.Println("handling sourceBucket: ", bucket)

	runfig.SourceBucketList = append(runfig.SourceBucketList, bucket)
	var sourceBucketListBSON bson.A = bson.A{}
	for _, b := range runfig.SourceBucketList {
		sourceBucketListBSON = append(sourceBucketListBSON, b.ToBSON())
	}
	runfig.SourceBuckets[bucket.BucketURL+bucket.BucketName] = bucket

	dbutil.UpdateSourceBuckets(runfig.MongoClient, runfig.MongoCtx, sourceBucketListBSON)
}
