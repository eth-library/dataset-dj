package main

import (
	"github.com/eth-library-lab/dataset-dj/dbutil"
	"github.com/google/uuid"
)

func generateToken() string {
	// Generate UUID for archives and use only the first 4 bytes
	newUID := uuid.New().String()[:8]

	// Regenerate new UUIDs as long as there are collisions
	for ok := runfig.ArchiveIDs.Check(newUID); ok; {
		newUID = uuid.New().String()[:8]
	}

	runfig.ArchiveIDs.Add(newUID)
	dbutil.UpdateArchiveIDs(runfig.MongoCtx, runfig.MongoClient, config.DbName, runfig.ArchiveIDs.ToSlice())

	return newUID
}
