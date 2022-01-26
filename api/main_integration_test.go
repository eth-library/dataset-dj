//go:build integration

package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

var host string = "http://0.0.0.0:8765"

func TestMain(m *testing.M) {

	// set up BEFORE the tests
	os.Setenv("DB_NAME", "test")
	os.Setenv("GIN_MODE", "test")
	setupConfig()

	runfig.MongoClient.Database("test").Drop(runfig.MongoCtx)
	fmt.Printf("\ntest setup complete. using mode %v\n\n", config.Mode)
	// run the tests
	exitVal := m.Run()
	// tidy up AFTER the tests
	os.Exit(exitVal)
}

func TestPing(t *testing.T) {

	url := host + "/ping"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed ping get request")
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// expectedCode := 200
	// if resp.StatusCode != 200 {
	// 	t.Fatalf("expected status code %v, got %v", expectedCode, resp.StatusCode)
	// }

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response body")
	}
	expectedResp := "pong"
	gotResp := string(body)
	if gotResp != expectedResp {
		t.Fatalf("expected %v, got %v", expectedResp, gotResp)
	}
}

func TestAuth(t *testing.T) {

	client := &http.Client{}
	// var req http.NewRequest{}

	testCases := []struct {
		errMsg       string
		path         string
		header       http.Header
		body         io.Reader
		expectedCode int
	}{
		{
			"should not need authentication",
			"/ping",
			http.Header{},
			nil,
			200,
		},
		{
			"should not need authentication, but invalid link id should fail",
			"/key/claim/1234",
			http.Header{},
			nil,
			400,
		},
		{
			"should not authenticate if token is missing",
			"/files",
			http.Header{
				"Authorization": []string{""},
			},
			nil,
			401,
		},
		{
			"should not authenticate if token is invalid",
			"/files",
			http.Header{
				"Authorization": []string{"Bearer foo"},
			},
			nil,
			401,
		},
		{
			"should not authenticate if not service Key",
			"/files",
			http.Header{
				"Authorization": []string{"Bearer ak_a1d9b19fb064d0cbbbf55bdcb07331e4"},
			},
			nil,
			401,
		},
	}

	for _, tc := range testCases {
		url := host + tc.path
		req, err := http.NewRequest("GET", url, tc.body)
		req.Header = tc.header
		resp, err := client.Do(req)

		assert.Equal(t, nil, err, "should be error free")
		assert.Equal(t, tc.expectedCode, resp.StatusCode, tc.errMsg)
	}

}
