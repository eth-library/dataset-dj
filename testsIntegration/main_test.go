//go:build integration

package integration_tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

var host = "http://0.0.0.0:8765"

func TestPass(t *testing.T) {
	assert.Equal(t, 0, 0, "this should pass")
}

// func TestFail(t *testing.T) {
// 	assert.Equal(t, 0, 1, "this should fail")
// }

func TestPing(t *testing.T) {

	url := host + "/ping"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed ping get request")
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

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

func TestMain(m *testing.M) {
	fmt.Println("start integration testing")
	exitVal := m.Run()
	fmt.Println("finished integration testing")
	// tidy up AFTER the tests
	os.Exit(exitVal)
}
