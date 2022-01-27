package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

//test that router is returned and reachable
func TestSetupRouter(t *testing.T) {

	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	var msg string
	msg = fmt.Sprintf("expected %v got %v", 200, w.Code)
	assert.Equal(t, http.StatusOK, w.Code, msg)
	msg = fmt.Sprintf("expected %v got %v", "pong", w.Body.String())
	assert.Equal(t, "pong", w.Body.String(), msg)
}

// func TestAuthTable(t *testing.T) {

// 	router := setupRouter()
// 	testCases := []struct {
// 		name         string
// 		header       http.Header
// 		body         io.Reader
// 		expectedCode int
// 		errMsg       string
// 	}{
// 		{"no auth header",
// 			http.Header{
// 				"Authorization": []string{""},
// 			},
// 			nil,
// 			401,
// 			"should not authenticate if no auth header",
// 		},
// 		{"Authorization header incorrect format",
// 			http.Header{
// 				"Authorization": []string{"sk_215as4as"},
// 			},
// 			nil,
// 			401,
// 			"should not authenticate if Authorization header incorrect format",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		req, _ := http.NewRequest("GET", "/files", nil)
// 		req.Header = tc.header
// 		r := httptest.NewRecorder()
// 		router.ServeHTTP(r, req)
// 		assert.Equal(t, tc.expectedCode, r.Code, tc.errMsg)

// 	}
// }
