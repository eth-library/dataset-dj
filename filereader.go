package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
)

var archBaseName string = "archive"

func getFiles(request archiveRequest) error {

	fmt.Println("creating zip archive...")
	archive, err := os.Create(storage + archBaseName + "_" + request.ArchiveID + ".zip")
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	for i, file := range request.Files {

		fmt.Printf("downloading file %d / %d: %s\n", i+1, len(request.Files), collection+file)
		err := writeToZip(file, zipWriter)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("closing zip archive...")
	zipWriter.Close()
	return nil
}

func writeToZip(fileName string, writer *zip.Writer) error {

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