package main

import "log"

var pathPrefix string = "/Users/magnuswuttke/coding/go/datadj/"
var collection string = pathPrefix + "collection/"
var storage string = pathPrefix + "storage/"
var namePrefix string = "cmt-001_1917_001_00"
var suffix string = ".jpg"
var archiveName = "archive.zip"

func main() {
	numbers := []string{"46", "55", "69", "71", "72"}
	fileNames := make([]string, len(numbers))
	for i, n := range numbers {
		fileNames[i] = namePrefix + n + suffix
	}
	err := getFiles(fileNames)
	if err != nil {
		log.Fatal(err)
	}
	err = sendNotification(fileNames)
	if err != nil {
		log.Fatal(err)
	}
}
