//go:build integration

package integration_tests

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestAuthRequiredEndpoints(t *testing.T) {

	client := &http.Client{}

	validAdminKey := os.Getenv("ADMIN_KEY")

	testCases := []struct {
		errMsg       string
		path         string
		method       string
		header       http.Header
		body         io.Reader
		expectedCode int
	}{
		{
			"should not need authentication",
			"/ping",
			"GET",
			http.Header{},
			nil,
			200,
		},
		{
			"should not need authentication, but invalid link id should fail",
			"/key/claim/1234",
			"GET",
			http.Header{},
			nil,
			400,
		},
		{
			"should not authenticate if token is missing",
			"/files",
			"GET",
			http.Header{
				"Authorization": []string{""},
			},
			nil,
			401,
		},
		{
			"should not authenticate if token is invalid",
			"/files",
			"GET",
			http.Header{
				"Authorization": []string{"Bearer foo"},
			},
			nil,
			401,
		},
		{
			"should not authenticate if not a service Key",
			"/files",
			"GET",
			http.Header{
				"Authorization": []string{"Bearer " + validAdminKey},
			},
			nil,
			401,
		},
		{
			"admin endpoint should not authenticate if admin Key not valid",
			"/admin/createKeyLink",
			"POST",
			http.Header{
				"Authorization": []string{"Bearer ak_a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"},
			},
			nil,
			401,
		},
		{
			"admin endpoint should not authenticate if admin Key not valid",
			"/admin/revokeKey",
			"POST",
			http.Header{
				"Authorization": []string{"Bearer ak_a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"},
			},
			nil,
			401,
		},
		{
			"admin endpoint should not authenticate if admin Key not valid",
			"/admin/addSourceBucket",
			"POST",
			http.Header{
				"Authorization": []string{"Bearer ak_a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"},
			},
			nil,
			401,
		},
		{
			"admin endpoint should authenticate if admin Key IS valid but form is missing",
			"/admin/addSourceBucket",
			"POST",
			http.Header{
				"Authorization": []string{"Bearer " + validAdminKey},
			},
			nil,
			400,
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
