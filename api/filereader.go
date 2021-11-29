package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

var archBaseName string = "archive"

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

	fmt.Println("creating local zip archive...")
	archive, err := os.Create(config.archiveLocalDir + archBaseName + "_" + request.ArchiveID + ".zip")
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	for i, file := range request.Files {

		fmt.Printf("downloading file %d / %d: %s\n", i+1, len(request.Files), config.sourceLocalDir+file)
		err := WriteToZipLocal(file, zipWriter)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("closing zip archive...")
	zipWriter.Close()
	return nil
}

//WriteToZipLocal is a helper function for locally zipping files
func WriteToZipLocal(fileName string, writer *zip.Writer) error {

	f, err := os.Open(config.sourceLocalDir + fileName)
	if err != nil {
		return fmt.Errorf("could not find file: %s", config.sourceLocalDir+fileName)
	}
	defer f.Close()

	fmt.Printf("writing file to archive: %s\n", fileName)
	w, err := writer.Create(fileName)
	if err != nil {
		return fmt.Errorf("could not create file in archive: %s", fileName)
	}

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("could not write to file in archive: %s", fileName)
	}
	return nil
}
