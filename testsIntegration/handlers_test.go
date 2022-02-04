package integrationtests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/eth-library-lab/dataset-dj/dbutil"
	"github.com/stretchr/testify/assert"
)

// test that the request for an archive returns ok status and the file id
// and that the zip file gets written to the correct location
func TestHandleArchive(t *testing.T) {

	host := "http://0.0.0.0:8765"
	url := host + "/archive"
	serviceKey := os.Getenv("SERVICE_KEY")
	fmt.Println("serviceKey: ", serviceKey)

	type filesRequest struct {
		Email string   `json:"email"`
		Files []string `json:"files"`
	}

	fileReq := filesRequest{
		Email: "barry.sunderland@outlook.com",
		Files: []string{
			"local/cmt-001_1917_001_0001.jpg",
			"local/cmt-001_1917_001_0005.jpg",
			"local/cmt-001_1917_001_0010.jpg"},
	}
	fmt.Println(fileReq)

	emailJSON, err := json.Marshal(fileReq)
	assert.Equal(t, nil, err, "error preparing test: could not marshal file request into json")

	emailBody := bytes.NewBuffer(emailJSON)
	req, err := http.NewRequest("POST", url, emailBody)
	assert.Equal(t, nil, err, "error preparing test: could not create new request")
	req.Header = http.Header{
		"Authorization": []string{"Bearer " + serviceKey},
		"content":       []string{"application/json"},
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.Equal(t, nil, err, "request should be error free")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "response should be ok")

	body, err := ioutil.ReadAll(resp.Body)
	var respBody dbutil.MetaArchive
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		fmt.Println("error unmarshalling response")
	}
	fmt.Printf("\nresponse body: \n%+v\n", respBody)
	fmt.Println("respBody.ID: ", respBody.ID)

	// test that zip file is created
	archiveDir := os.Getenv("ARCHIVE_LOCAL_DIR")
	expectedZipPath := archiveDir + "archive_" + respBody.ID + ".zip"
	// wait to make sure zip file has been written
	fmt.Println("wait for zip...")
	time.Sleep(1 * time.Second)
	stats, err := os.Stat(expectedZipPath)
	fmt.Println("stats: ", stats)
	errMsg := fmt.Sprintf("expected zip file to be saved to: %v\n instead got %v", expectedZipPath, err)
	assert.Equal(t, nil, err, errMsg)
	assert.NotEqual(t, nil, stats, "zip file found")

}
