package main

import (
	uuid "github.com/google/uuid"
)

var archiveIDs set

func generateToken() string {
	// Generate UUID for archives and use only the first 4 bytes
	newUid := uuid.New().String()[:8]

	// Regenerate new UUIDs as long as there are collisions
	for ok := archiveIDs.Check(newUid); ok; {
		newUid = uuid.New().String()[:8]
	}

	archiveIDs.Add(newUid)
	updateArchiveIDs(archiveIDs.toSlice())

	return newUid
}
