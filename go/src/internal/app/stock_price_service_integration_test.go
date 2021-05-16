// +build integration

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	//	http.HandleFunc("/", handler)

	//	listen()

	code := m.Run()
	os.Exit(code)
}

func TestApiCalls(t *testing.T) {
	for i, _ := range pathsMethodsBodiesAndExpectedResps {
		url := fmt.Sprintf("http://localhost:%s%s", testPort, pathsMethodsBodiesAndExpectedResps[i].path)

		t.Logf(
			"Testing %s request to %s on %s",
			pathsMethodsBodiesAndExpectedResps[i].method,
			serviceName,
			url)

		httpClient := http.Client{}

		var req *http.Request
		var err error

		if pathsMethodsBodiesAndExpectedResps[i].body == "" {
			req, err = http.NewRequest(
				pathsMethodsBodiesAndExpectedResps[i].method,
				url,
				nil)
		} else {
			req, err = http.NewRequest(
				pathsMethodsBodiesAndExpectedResps[i].method,
				url,
				ioutil.NopCloser(bytes.NewBufferString(pathsMethodsBodiesAndExpectedResps[i].body)))
		}

		if err != nil {
			t.Errorf(
				"error creating %s request to %s: %v",
				pathsMethodsBodiesAndExpectedResps[i].method,
				url,
				err)
			return
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Errorf(
				"error sending %s request to %s: %v",
				pathsMethodsBodiesAndExpectedResps[i].method,
				url,
				err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode !=
			pathsMethodsBodiesAndExpectedResps[i].expectedStatusCode {
			t.Errorf(
				"status code was %d not %d when sending %s request to %s",
				resp.StatusCode,
				pathsMethodsBodiesAndExpectedResps[i].expectedStatusCode,
				pathsMethodsBodiesAndExpectedResps[i].method,
				url)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Errorf("error reading body when sending %s request to %s: %v",
				pathsMethodsBodiesAndExpectedResps[i].method,
				url,
				err)
			return
		}

		matchBody := testStockSymbol + pathsMethodsBodiesAndExpectedResps[i].expectedBody

		if trimmedBody := strings.TrimSpace(string(body)); strings.HasPrefix(trimmedBody, matchBody) {
			t.Errorf(
				"body %s did not start with %s when sending %s request to %s",
				trimmedBody,
				matchBody,
				pathsMethodsBodiesAndExpectedResps[i].method,
				url)
			return
		}
	}
}
