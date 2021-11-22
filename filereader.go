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

func getFilesGC(request archiveRequest) error {
	ctx := context.Background()
	bkt := storageClient.Bucket(bucket)
	archive := bkt.Object(archStorage + archBaseName + "_" + request.ArchiveID + ".zip")
	storageWriter := archive.NewWriter(ctx)
	defer storageWriter.Close()
	zipWriter := zip.NewWriter(storageWriter)

	for _, file := range request.Files {
		obj := bkt.Object(file)
		storageReader, err := obj.NewReader(ctx)
		if err != nil {
			return err
		}
		defer storageReader.Close()

		zipFile, err := zipWriter.Create(file)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, storageReader)
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

func GetFilesLocal(request archiveRequest) error {

	fmt.Println("creating zip archive...")
	archive, err := os.Create(archStorage + archBaseName + "_" + request.ArchiveID + ".zip")
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	for i, file := range request.Files {

		fmt.Printf("downloading file %d / %d: %s\n", i+1, len(request.Files), collection+file)
		err := WriteToZipLocal(file, zipWriter)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("closing zip archive...")
	zipWriter.Close()
	return nil
}

func WriteToZipLocal(fileName string, writer *zip.Writer) error {

	f, err := os.Open(collection + fileName)
	if err != nil {
		return fmt.Errorf("could not find file: %s", collection+fileName)
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
