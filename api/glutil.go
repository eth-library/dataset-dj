package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func retrieveFilesLocal(localSourceDir string) ([]string, error) {
	return listFileDir(localSourceDir)
}

func retrieveFilesCloud(client *storage.Client, config *serverConfig) ([]string, error) {
	ctx := context.Background()
	var cloudFiles []string

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// get bucket handler and obtain an iterator over all objects returned by query

	bucket := client.Bucket(config.sourceBucketName)

	it := bucket.Objects(ctx, &storage.Query{
		Prefix: config.sourceBucketPrefix,
	})

	// Loop over all objects returned by the query
	for {
		attrs, err := it.Next()
		if err != nil {
			return nil, fmt.Errorf("An error occured while retrieving a file from the cloud storage")
		}

		if err == iterator.Done {
			break
		}

		if attrs.Name == config.sourceBucketPrefix { // make sure the directory is not listed as available file
			continue
		}
		cloudFiles = append(cloudFiles, attrs.Name)
	}
	return cloudFiles, nil
}

func retriveFilesAPI() ([]string, error) {
	return []string{}, nil
}

// list names of files in the given directory
func listFileDir(dirPath string) ([]string, error) {

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var filenames []string

	for _, file := range files {
		filenames = append(filenames, file.Name())
		//print filename and if its a direcory
		// fmt.Println(file.Name(), file.IsDir())
	}

	return filenames, nil
}
