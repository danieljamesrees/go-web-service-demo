// +build unit

package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

const (
	testStockSymbol  = "MSFTMORE"
	testNumberOfDays = 7

	seed = 42
)

var (
	testStockPrices = []float32{10.00, 20.00, 30.00, 40.00, 50.00, 60.00, 70.00}

	// Should really compute this, but tricky to populate the struct.
	expectedStockPrices = testStockSymbol + " data=[10.00, 20.00, 30.00, 40.00, 50.00, 60.00, 70.00], average=40.00"
)

func mockStockPrices(w http.ResponseWriter, stockSymbol string, numberOfDays int) string {
	return expectedStockPrices
}

func TestMain(m *testing.M) {
	err := os.Setenv("NDAYS", strconv.Itoa(testNumberOfDays))
	if err != nil {
		panic(err)
	}

	err = os.Setenv("SYMBOL", testStockSymbol)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	os.Exit(code)
}

func TestAllPricesAndMean_ValidStockSymbol_ReturnOKAndExpectedValues(t *testing.T) {
	expectedResp := testStockSymbol + " data=["

	var totalStockPrice float32 = 0.0

	for i := 0; i < testNumberOfDays; i++ {
		stockPrice := testStockPrices[i]
		stockPriceString := fmt.Sprintf("%.2f", stockPrice)

		expectedResp += stockPriceString

		if i < (testNumberOfDays - 1) {
			expectedResp += ", "
		}

		totalStockPrice += stockPrice
	}

	meanStockPriceString := fmt.Sprintf("%.2f", totalStockPrice/testNumberOfDays)

	expectedResp += "], average=" + meanStockPriceString

	t.Logf("Testing allPricesAndMean with valid stock symbol %s and %d days", testStockSymbol, testNumberOfDays)

	w := httptest.NewRecorder()

	allPricesAndMean(w, testStockSymbol, testNumberOfDays, mockStockPrices)

	resp := w.Result()

	t.Logf("Testing allPricesAndMean with valid stock symbol %s and %d days returned response: %v", testStockSymbol, testNumberOfDays, resp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code was %d not %d when getting stock price for %s", resp.StatusCode, http.StatusOK, testStockSymbol)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if string(body) != expectedResp {
		t.Errorf("body was %s not %s when reading response", body, expectedResp)
	}
}

func TestAllPricesAndMean_InvalidStockSymbol_ReturnNotFound(t *testing.T) {
	testStockSymbol := "X"

	t.Logf("Testing allPricesAndMean with invalid stock symbol %s and %d days", testStockSymbol, testNumberOfDays)

	w := httptest.NewRecorder()

	allPricesAndMean(w, testStockSymbol, testNumberOfDays, mockStockPrices)

	resp := w.Result()

	t.Logf("Testing allPricesAndMean with invalid stock symbol %s and %d days returned response: %v", testStockSymbol, testNumberOfDays, resp)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status code was %d not %d when getting stock price for %s", resp.StatusCode, http.StatusNotFound, testStockSymbol)
		return
	}
}
