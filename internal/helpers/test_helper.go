package helpers

import (
	"net/http"
	"net/http/httptest"
)

// MakeTestServer creates an api server for testing
func MakeTestServer(responseCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(responseCode)
		_, err := res.Write([]byte(body))
		if err != nil {
			panic(err)
		}
	}))
}

// MakeRequestCapturingTestServer creates an api server that captures the request object
func MakeRequestCapturingTestServer(responseCode int, body string, request *http.Request) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		*request = *req
		res.WriteHeader(responseCode)
		_, err := res.Write([]byte(body))
		if err != nil {
			panic(err)
		}
	}))
}
