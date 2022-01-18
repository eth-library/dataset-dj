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

type archiveRequest struct {
	Email     string   `json:"email"`
	ArchiveID string   `json:"archiveID"`
	Files     []string `json:"files"`
}

var archBaseName string = "archive" // prefix to include at start of archive filename

// handleArchiveMessage handles a message received in the 'archives' channel
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

// zipFiles is a wrapper function that decides if zipFilesLocal or zipFilesGC ('zipFilesGoogleCloud') should be called
func zipFiles(archRequest archiveRequest) error {

	split := splitFiles(archRequest.Files)

	if config.ArchiveLocalDir != "" {
		// fmt.Println("creating local zip archive...")
		archiveFilePath := config.ArchiveLocalDir + archBaseName + "_" + archRequest.ArchiveID + ".zip"
		archive, err := os.Create(archiveFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer archive.Close()
		zipWriter := zip.NewWriter(archive)

		for i, file := range split.localFiles {

			err := WriteLocalToZip(file, zipWriter)
			if err != nil {
				fmt.Printf("\r zipping file %d / %d: %s\n", i+1, len(archRequest.Files), config.SourceLocalDir+file)
				log.Fatal(err)
			}
		}

		ctx := context.Background()
		bkt := runfig.StorageClient.Bucket(config.SourceBucketName) // get bucket handle

		for _, file := range split.cloudFiles {
			obj := bkt.Object(file)
			storageReader, err := obj.NewReader(ctx) // file reader
			if err != nil {
				return err
			}
			defer storageReader.Close()

			zipFile, err := zipWriter.Create(file) // create the file inside the zip archive
			if err != nil {
				return fmt.Errorf("could not create file in local archive: %s", file)
			}
			_, err = io.Copy(zipFile, storageReader) // copy content of the file into the new zipped version
			if err != nil {
				return fmt.Errorf("could not write to file in local archive: %s", file)
			}
		}

		for _, file := range split.apiFiles {
			fmt.Println("API to local: " + file)
		}

		err = zipWriter.Close()
		if err != nil {
			return err
		}
		fmt.Println("local zip archive written to: ", archiveFilePath)
	}

	if config.ArchiveBucketName != "" && config.ArchiveBucketPrefix != "" {
		ctx := context.Background()
		srcbkt := runfig.StorageClient.Bucket(config.SourceBucketName)
		archbkt := runfig.StorageClient.Bucket(config.ArchiveBucketName)                                            // get bucket handle
		archive := archbkt.Object(config.ArchiveBucketPrefix + archBaseName + "_" + archRequest.ArchiveID + ".zip") // create zip archive
		storageWriter := archive.NewWriter(ctx)                                                                     // create writer that writes to the bucket
		defer storageWriter.Close()
		zipWriter := zip.NewWriter(storageWriter) // create zip writer that writes to the bucket writer

		for _, file := range split.cloudFiles {
			obj := srcbkt.Object(file)
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

		for i, file := range split.localFiles {

			err := WriteLocalToZip(file, zipWriter)
			if err != nil {
				fmt.Printf("\r zipping file %d / %d: %s\n", i+1, len(archRequest.Files), config.SourceLocalDir+file)
				log.Fatal(err)
			}
		}

		for _, file := range split.apiFiles {
			fmt.Println("API to cloud: " + file)
		}

		err := zipWriter.Close()
		if err != nil {
			return err
		}
		fmt.Println("cloud zip archive written to: ", archive)
	}

	return nil
}

//WriteToZipLocal is a helper function for writing an individual local file to zip.Writer object
func WriteLocalToZip(fileName string, writer *zip.Writer) error {

	f, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("could not find file: %s", fileName)
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

// ----------------------------------------- Legacy Code ------------------------------------------------------

// zipFilesGC retrieves files listed in the metaArchive request and fetches them from the cloud storage
// immediately rewriting the to the storage as a zip archive
func ZipFilesGC(request archiveRequest) error {
	ctx := context.Background()
	bkt := runfig.StorageClient.Bucket(config.ArchiveBucketName)                                        // get bucket handle
	archive := bkt.Object(config.ArchiveBucketPrefix + archBaseName + "_" + request.ArchiveID + ".zip") // create zip archive
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
func ZipFilesLocal(request archiveRequest) error {

	// fmt.Println("creating local zip archive...")
	archiveFilePath := config.ArchiveLocalDir + archBaseName + "_" + request.ArchiveID + ".zip"
	archive, err := os.Create(archiveFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	for i, file := range request.Files {

		err := WriteLocalToZip(file, zipWriter)
		if err != nil {
			fmt.Printf("\r zipping file %d / %d: %s\n", i+1, len(request.Files), config.SourceLocalDir+file)
			log.Fatal(err)
		}
	}

	zipWriter.Close()
	fmt.Println("zip archive written to: ", archiveFilePath)
	return nil
}
