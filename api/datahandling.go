package main

import (
	"net/mail"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// "Database" for the metaArchives so far
// var archives map[string]metaArchive = make(map[string]metaArchive)

// File represents metadata about a file, not used so far
type File struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Size     int32  `json:"size"`
}

// metaArchives are the blueprints for the zip archives that will be created once the user initiates
// the download process. Files is implemented as a set in order to avoid duplicate files within a
// metaArchive
// type metaArchive struct {
// 	ID    string `json:"id"`
// 	Files set    `json:"files"`
// }
type metaArchive struct {
	ID          string `json:"id"`
	Files       set    `json:"files"`
	TimeCreated string `json:"timeCreated"`
	TimeUpdated string `json:"timeUpdated"`
	Status      string `json:"status"`
	Source      string `json:"source"`
}

// a set is a struct with one attribute that are its elements contained within a map
type set struct {
	elems map[string]bool `json:"elements"`
}

func (a metaArchive) toBSON() bson.D {
	var files bson.A
	for _, v := range a.Files.toSlice() {
		files = append(files, v)
	}
	return bson.D{primitive.E{Key: "_id", Value: a.ID}, primitive.E{Key: "files", Value: files}}
}

func emailIsValid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// checks if elem is contained within set s
func (s set) Check(elem string) bool {
	return s.elems[elem]
}

// adds elem to set s
func (s set) Add(elem string) {
	s.elems[elem] = true
}

// deletes elem from set s
func (s set) Del(elem string) {
	delete(s.elems, elem)
}

// replace the elements of a set by the contents of a slice
func (s set) SetElemsFromSlice(slice []string) {
	s.elems = map[string]bool{}
	for _, e := range slice {
		s.elems[e] = true
	}
}

// return the elements of a set as a slice
func (s set) toSlice() []string {
	slice := []string{}
	for e := range s.elems {
		slice = append(slice, e)
	}
	return slice
}

// create a set from a slice
func setFromSlice(slice []string) set {
	s := set{elems: map[string]bool{}}
	for _, e := range slice {
		s.elems[e] = true
	}
	return s
}

// return a new set as the union of two sets
func setUnion(s1, s2 set) set {
	newSet := set{elems: map[string]bool{}}
	for e := range s1.elems {
		newSet.elems[e] = true
	}
	for e := range s2.elems {
		newSet.elems[e] = true
	}
	return newSet
}

func retrieveAllFiles() ([]string, error) {
	var allAvailableFiles []string
	localFiles, err := retrieveFilesLocal(config.sourceLocalDir)
	if err != nil {
		return nil, err
	}
	allAvailableFiles = append(allAvailableFiles, localFiles...)

	cloudFiles, err := retrieveFilesCloud(storageClient, config)
	if err != nil {
		return allAvailableFiles, err
	}
	allAvailableFiles = append(allAvailableFiles, cloudFiles...)

	apiFiles, err := retriveFilesAPI()
	if err != nil {
		return allAvailableFiles, err
	}
	allAvailableFiles = append(allAvailableFiles, apiFiles...)
	return allAvailableFiles, nil
}
