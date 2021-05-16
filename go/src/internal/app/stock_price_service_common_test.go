package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
)

type pathMethodBodyAndExpectedResp struct {
	path               string
	method             string
	body               string
	expectedStatusCode int
	expectedBody       string
}

const (
	testStockSymbol  = "MSFT"
	testNumberOfDays = 5

	// Should really try and devise a deterministic response.
	expectedStockPrices = testStockSymbol + " data=["
)

var (
	testPort string = strings.TrimSpace(os.Getenv("TEST_PORT"))

	pathsMethodsBodiesAndExpectedResps []pathMethodBodyAndExpectedResp
)

func init() {
	err := os.Setenv("NDAYS", strconv.Itoa(testNumberOfDays))
	if err != nil {
		panic(err)
	}

	err = os.Setenv("SYMBOL", testStockSymbol)
	if err != nil {
		panic(err)
	}

	pathsMethodsBodiesAndExpectedResps = []pathMethodBodyAndExpectedResp{
		{"/stockpricedemo", http.MethodGet, "", http.StatusOK, expectedStockPrices},
	}
}
