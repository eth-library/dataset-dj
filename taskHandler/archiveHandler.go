package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

var archBaseName string = "archive" // prefix to include at start of archive filename

//handleArchiveMessage handles a message received in the 'archives' channel
func handleArchiveMessage(messagePayload string) {

	var archRequest archiveRequest

	// convert json string into struct
	json.Unmarshal([]byte(messagePayload), &archRequest)

	fmt.Println("handling archRequest: ", archRequest)
	err := zipFiles(archRequest)
	if err == nil && archRequest.Email != "" {
		publishEmailTask(archRequest)
	} else {
		fmt.Println("err: ", err)
	}
}

//zipFiles is a wrapper function that decides if zipFilesLocal or zipFilesGC ('zipFilesGoogleCloud') should be called
func zipFiles(archRequest archiveRequest) error {

	var err error
	if archRequest.Source == "local" {
		err = zipFilesLocal(archRequest)
		if err != nil {
			fmt.Println("error returned while zipping local files", err)
		}
	} else { // assume from cloud bucket
		err = zipFilesGC(archRequest)
		if err != nil {
			fmt.Println("error returned while zipping local files", err)
		}

	}
	return err
}

// zipFilesGC retrieves files listed in the metaArchive request and fetches them from the cloud storage
// immediately rewriting the to the storage as a zip archive
func zipFilesGC(request archiveRequest) error {
	ctx := context.Background()
	bkt := storageClient.Bucket(config.archiveBucketName)                                               // get bucket handle
	archive := bkt.Object(config.archiveBucketPrefix + archBaseName + "_" + request.ArchiveID + ".zip") // create zip archive
	storageWriter := archive.NewWriter(ctx)                                                             // create writer that writes to the bucket
	defer storageWriter.Close()
	zipWriter := zip.NewWriter(storageWriter) // create zip writer that writes to the bucket writer

	for _, file := range request.Files {
		obj := bkt.Object(file)
		storageReader, err := obj.NewReader(ctx) // file reader
		if err != nil {
			return err
		}
		defer storageReader.Close()

		zipFile, err := zipWriter.Create(file) // create the file inside the zip archive
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, storageReader) // copy content of the file into the new zipped version
		if err != nil {
			return err
		}
	}
	err := zipWriter.Close()
	if err != nil {
		return err
	}
	return nil
}

// zipFilesLocal copies files accesible by local filepaths into a newly created zip archive
func zipFilesLocal(request archiveRequest) error {

	// fmt.Println("creating local zip archive...")
	archiveFilePath := config.archiveLocalDir + archBaseName + "_" + request.ArchiveID + ".zip"
	archive, err := os.Create(archiveFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	for i, file := range request.Files {

		err := WriteToZipLocal(file, zipWriter)
		if err != nil {
			fmt.Printf("\r zipping file %d / %d: %s\n", i+1, len(request.Files), config.sourceLocalDir+file)
			log.Fatal(err)
		}
	}

	zipWriter.Close()
	fmt.Println("zip archive written to: ", archiveFilePath)
	return nil
}

//WriteToZipLocal is a helper function for writing an individual local file to zip.Writer object
func WriteToZipLocal(fileName string, writer *zip.Writer) error {

	f, err := os.Open(config.sourceLocalDir + fileName)
	if err != nil {
		return fmt.Errorf("could not find file: %s", config.sourceLocalDir+fileName)
	}
	defer f.Close()

	w, err := writer.Create(fileName)
	if err != nil {
		return fmt.Errorf("could not create file in archive: %s", fileName)
	}

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("could not write to file in archive: %s", fileName)
	}
	return nil
}
