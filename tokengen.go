package main

import (
	uuid "github.com/google/uuid"
)

func generateToken() string {
	// Generate UUID for archives and use only the first 4 bytes
	newUid := uuid.New().String()[:8]

	// Regenerate new UUIDs as long as there are collisions
	for _, ok := archives[newUid]; ok; {
		newUid = uuid.New().String()[:8]
	}

	return newUid
}
