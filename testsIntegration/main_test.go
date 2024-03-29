package integrationtests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var host = "http://0.0.0.0:8765"

func TestMain(m *testing.M) {
	// set up before testing
	fmt.Println("start integration testing")
	// run tests
	exitVal := m.Run()
	fmt.Println("finished integration testing")
	// tidy up AFTER the tests
	os.Exit(exitVal)
}

func TestThisShouldPass(t *testing.T) {
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
