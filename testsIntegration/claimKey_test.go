package integration_tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type APIKeyResponse struct {
	APIKey  string `json: "APIKey"`
	Message string `json: "Message"`
}

func TestClaimAPIKey(t *testing.T) {

	client := &http.Client{}
	testCases := []struct {
		errMsg             string
		linkID             string
		expectedCode       int
		expectMarshalError bool
	}{
		{
			"valid link ID should return apiKey",
			"61f4231861914a390f009893",
			http.StatusOK,
			false,
		},
		{
			"already used link ID should return bad request",
			"61f4231861914a390f009893",
			http.StatusBadRequest,
			true,
		},
		{
			"invalid link should return bad request",
			"a1a1a1a1a1a1a1a1a1a1a1a1",
			http.StatusBadRequest,
			true,
		},
	}

	var keyResp APIKeyResponse

	for i, tc := range testCases {
		url := host + "/key/claim/" + tc.linkID
		req, err := http.NewRequest("GET", url, nil)
		resp, err := client.Do(req)

		msgPrefix := fmt.Sprintf("testcase %v: ", i)
		assert.Equal(t, nil, err, msgPrefix+"request should be error free")
		assert.Equal(t, tc.expectedCode, resp.StatusCode, msgPrefix+tc.errMsg)

		body, err := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &keyResp)
		hadMarshalError := err != nil
		assert.Equal(t, tc.expectMarshalError, hadMarshalError, msgPrefix+"resp body not as expected")

		// test that the returned apiKey works
		if tc.expectMarshalError == false {

			url = host + "/key/validate/"
			req, _ := http.NewRequest("GET", url, nil)
			req.Header = http.Header{
				"Authorization": []string{"Bearer " + keyResp.APIKey},
			}

			resp, _ := client.Do(req)
			assert.Equal(t, resp.StatusCode, 200, msgPrefix+"returned key should validate successfully")

		}
	}

}
