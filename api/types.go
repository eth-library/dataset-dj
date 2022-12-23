package main

import (
	db "github.com/eth-library/dataset-dj/dbutil"
)

// archiveRequest is the main data structure that is being received by the API when information or
// modifications about archives are requested. Email simply is an email as string, ArchiveID is the UID of
// a metaArchive as string and Files is a list of fileNames as strings. Possible combinations:
// 1. Email and ArchiveID set, Files empty -> Retrieve metaArchive with ArchiveID, create the zip archive
// 	  and send download link to Email
// 2. ArchiveID and Files set, Email empty -> Add fileNames in Files to metaArchive with ArchiveID
// 3. Email and Files set, ArchiveID empty -> Create new metaArchive containing the fileNames from Files,
//	  immediatly retrieve the files from the collection and create the zip archive and send the download
//    link to Email
// 4. Files set, Email and ArchiveID empty -> Create new metaArchive from the fileNames in Files
//
// The function handleArchive() implements the logic to identify the different cases and to act accordingly

type archiveRequest struct {
	Email     string           `json:"email"`
	ArchiveID string           `json:"archiveID"`
	Content   []db.FileGroupDB `json:"content"`
	Meta      string           `json:"meta"`
}

type orderRequest struct {
	Sources []string `json:"sources"`
}

type orderStatusRequest struct {
	NewStatus string `json:"newStatus"`
}

type sourceRequest struct {
	Name         string
	Organisation string
}
