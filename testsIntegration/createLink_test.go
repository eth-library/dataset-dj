package integrationtests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateKeyLink(t *testing.T) {

	client := &http.Client{}

	validAdminKey := os.Getenv("ADMIN_KEY")
	emailJSON, err := json.Marshal(map[string]string{
		"email": "barry.sunderland@outlook.com",
	})
	assert.Equal(t, err, nil, "error creating io.reader for request body")
	emailBody := bytes.NewBuffer(emailJSON)

	// assert.Equal(t, err, nil, "error creating io.reader for request body")

	testCases := []struct {
		errMsg       string
		path         string
		method       string
		header       http.Header
		body         io.Reader
		expectedCode int
	}{
		{
			"create new link should succeed",
			"/admin/createKeyLink",
			"POST",
			http.Header{
				"Authorization": []string{"Bearer " + validAdminKey},
				"content":       []string{"application/json"},
			},
			emailBody,
			201,
		},
	}

	for _, tc := range testCases {
		url := host + tc.path
		req, err := http.NewRequest(tc.method, url, tc.body)
		req.Header = tc.header
		resp, err := client.Do(req)

		assert.Equal(t, nil, err, "should be error free")
		assert.Equal(t, tc.expectedCode, resp.StatusCode, tc.path+" : "+tc.errMsg)
	}

}
