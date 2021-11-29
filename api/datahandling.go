package main

import (
	"io/ioutil"
	"log"
)

// "Database" for the metaArchives so far
var archives map[string]metaArchive = make(map[string]metaArchive)

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
type metaArchive struct {
	ID          string `json:"id"`
	Files       set    `json:"files"`
	TimeCreated string `json:"timeCreated"`
	TimeUpdated string `json:"timeUpdated"`
	Status      string `json:"status"`
}

// a set is a struct with one attribute that are its elements contained within a map
type set struct {
	elems map[string]bool
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

// little helper function that pipelines the download of the files contained in a metaArchive,
// the creation of the zip archive and sending a mail with the download link to the user together
func downloadFiles(request archiveRequest) {
	err := getFilesGC(request)
	if err != nil {
		log.Fatal(err)
	}
	err = sendNotification(request)
	if err != nil {
		log.Fatal(err)
	}
}
