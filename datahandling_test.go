// package main

// import (
// 	"fmt"
// 	"os"
// 	"testing"
// )

// func Test_addNumbers(t *testing.T) {
// 	result := 2 + 3
// 	expectedResult := 5
// 	if result != expectedResult {
// 		errMessage := fmt.Sprintf("should not be equal: expected %v, got %v \n", expectedResult, result)
// 		t.Error(errMessage)
// 	}
// }

// // createFile is a helper function called from multiple tests
// func createFile(t *testing.T) (string, error) {
// 	f, err := os.Create("tempFile")
// 	if err != nil {
// 		return "", err
// 	}
// 	// write some data to f
// 	t.Cleanup(func() {
// 		os.Remove(f.Name())
// 	})
// 	return f.Name(), nil
// }

// func TestFileProcessing(t *testing.T) {
// 	fName, err := createFile(t)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Println(fName)
// 	// do testing, don't worry about cleanup
// }	