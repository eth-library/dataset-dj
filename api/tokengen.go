package main

import (
	"github.com/eth-library/dataset-dj/dbutil"
	"github.com/google/uuid"
)

func generateToken() string {
	// Generate UUID for archives and use only the first 4 bytes
	newUID := uuid.New().String()[:8]

	// Regenerate new UUIDs as long as there are collisions
	for ok := runtime.ArchiveIDs.Check(newUID); ok; {
		newUID = uuid.New().String()[:8]
	}

	runtime.ArchiveIDs.Add(newUID)
	dbutil.UpdateArchiveIDs(runtime.MongoCtx, runtime.MongoClient, config.DbName, runtime.ArchiveIDs.ToSlice())

	return newUID
}
