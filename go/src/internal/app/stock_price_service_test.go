// +build unit

package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()

	os.Exit(code)
}

type statikFunc func()

func mockStatik(t *testing.T, handleStatik statikFunc, expectedStatusCode int, expectedBody string) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		handleStatik()

		statikFile, err := statikFS.Open("/index.html")
		if err != nil {
			t.Fatal(err)
			return
		}

		statikFileStat, err := statikFile.Stat()
		if err != nil {
			t.Fatal(err)
			return
		}

		statikData := make([]byte, statikFileStat.Size())
		bytesRead, err := statikFile.Read(statikData)
		if err != nil {
			t.Fatal(err)
			return
		}
		defer statikFile.Close()

		if bytesRead == 0 {
			t.Fatal("No bytes read")
			return
		}

		statikOutput := string(statikData[:])
		log.Printf("Read %s from index.html", statikOutput)

		io.WriteString(w, statikOutput)
	}

	handler(rr, req)
	resp := rr.Result()

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
		return
	}

	if statusCode := resp.StatusCode; statusCode != expectedStatusCode {
		t.Errorf("handler returned wrong status code: got %v want %v",
			statusCode, expectedStatusCode)
		return
	}

	if !strings.Contains(string(body[:]), expectedBody) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			body, expectedBody)
		return
	}
}

// All looks right, but only test when it works in reality.
func TestSwaggerUi(t *testing.T) {
	t.Log("Testing request to use Swagger")

	//	mockStatik(t, handleSwagger, http.StatusOK, "swagger")
}
