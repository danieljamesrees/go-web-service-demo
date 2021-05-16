// Implements the non-HATEOS REST API.
package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type StockPrices struct {
	TimeSeriesData TimeSeries `json:"Time Series (Daily)"`
}

type TimeSeries map[string]PriceInfo

type PriceInfo struct {
	ClosingPrice string `json:"4. close"`
}

type getStockPricesFunction func(w http.ResponseWriter, stockSymbol string, numberOfDays int) string

const (
	remoteApiUrl               = "https://www.alphavantage.co"
	minimumLengthOfStockSymbol = 2
	maximumNumberOfDays        = 100

	//emptyStockPrices =
	//        StockPrices{
	//                TimeSeries: {
	//                        },
	//		},
)

var (
	apiKey = strings.TrimSpace(os.Getenv("APIKEY"))

	fullRemoteApiUrlPrefix string
)

func init() {
	if apiKey == "" {
		fmt.Println("Must specify an API key in the environment variable, APIKEY")
		os.Exit(1)
	}

	// datatype=csv might be simpler/more concise.
	// output=compact returns a maximum of 100 data points.
	fullRemoteApiUrlPrefix = fmt.Sprintf("%s/query?apikey=%s&function=TIME_SERIES_DAILY_ADJUSTED&symbol=", remoteApiUrl, apiKey)
}

func AllPricesAndMean(w http.ResponseWriter, stockSymbol string, numberOfDays int) {
	allPricesAndMean(w, stockSymbol, numberOfDays, getStockPrices)
}

func allPricesAndMean(w http.ResponseWriter, stockSymbol string, numberOfDays int, getStockPrices getStockPricesFunction) {
	log.Printf("Stock symbol is %s and number of days to get is %d\n", stockSymbol, numberOfDays)

	if numberOfDays > maximumNumberOfDays {
		log.Printf("Number of days must be less than the maximum %d", maximumNumberOfDays)
		http.Error(w, "Number of days greater than the maximum", http.StatusBadRequest)
		return
	}

	err := validate(stockSymbol)
	if err != nil {
		log.Printf("Error validating stock symbol %s due to: %v", stockSymbol, err)
		http.Error(w, "Error validating stock symbol", http.StatusNotFound)
		return
	}

	// Probably should return an error type too.
	stockPrices := getStockPrices(w, stockSymbol, numberOfDays)

	if stockPrices != "" {
		setHeaders(w)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(stockPrices))
	}
}

func validate(stockSymbol string) error {
	if len([]rune(stockSymbol)) < minimumLengthOfStockSymbol {
		return fmt.Errorf("stock symbol %s is invalid", stockSymbol)
	}

	return nil
}

// Invokes the free alphavantage stock price service.
func getStockPrices(w http.ResponseWriter, stockSymbol string, numberOfDays int) string {
	fullRemoteApiUrl := fmt.Sprintf("%s%s", fullRemoteApiUrlPrefix, stockSymbol)

	// Don't log as it exposes the APIKEY.
	//log.Printf("Sending GET request to remote API %s for stock symbol %s", fullRemoteApiUrl, stockSymbol)
	log.Printf("Sending GET request to remote API for stock symbol %s", stockSymbol)

	resp, err := http.Get(fullRemoteApiUrl)

	if err != nil {
		log.Printf("Error getting stock price from remote API for stock symbol %s due to: %v", stockSymbol, err)
		http.Error(w, "Error getting stock price from remote API", http.StatusInternalServerError)
		return ""
	}

	respData, err := ioutil.ReadAll(resp.Body)

	//log.Printf("GET request to remote API for stock symbol %s returned: %s", stockSymbol, respData)

	if err != nil {
		log.Printf("Error reading response from remote API for stock symbol %s due to: %v", stockSymbol, err)
		http.Error(w, "Error reading response from remote API", resp.StatusCode)
		return ""
	}

	var jsonStockPrices StockPrices
	err = json.Unmarshal(respData, &jsonStockPrices)
	if err != nil {
		log.Printf("Error unmarshalling response for stock symbol %s due to: %v", stockSymbol, err)
		http.Error(w, "Error unmarshalling response", http.StatusInternalServerError)
		return ""
	}

	var serviceResp string
	serviceResp, err = buildServiceResp(stockSymbol, numberOfDays, jsonStockPrices)
	if err != nil {
		log.Printf("Error building service response for stock symbol %s due to: %v", stockSymbol, err)
		http.Error(w, "Error building service response", http.StatusInternalServerError)
		return ""
	}

	return serviceResp
}

func buildServiceResp(stockSymbol string, numberOfDays int, jsonStockPrices StockPrices) (string, error) {
	serviceResp := stockSymbol + " data=["

	var totalStockPrice float32 = 0.0

	count := 1

	// The data seems to be returned in date order, but the map loses this ordering. TODO Almost certainly hugely inefficient, and likely only necessary due to an incorrect data type.
	// Thankfully dates are in a sane date format.
	for _, date := range reverseSortedKeys(jsonStockPrices.TimeSeriesData) {
		price := jsonStockPrices.TimeSeriesData[date].ClosingPrice

		log.Printf("GET request to remote API for stock symbol %s returned price %s for date %s", stockSymbol, price, date)

		closingPrice, err := strconv.ParseFloat(price, 32)

		if err != nil {
			return "", fmt.Errorf("closing price %s cannot be converted to float due to %v", price, err)
		}

		serviceResp += price
		totalStockPrice += float32(closingPrice)

		if count < numberOfDays {
			serviceResp += ", "
		} else {
			break
		}

		count++
	}

	averageStockPriceString := fmt.Sprintf("%.2f", totalStockPrice/float32(numberOfDays))

	serviceResp += "], average=" + averageStockPriceString

	return serviceResp, nil
}

func reverseSortedKeys(m map[string]PriceInfo) []string {
	keys := make([]string, len(m))
	i := 0

	for k := range m {
		keys[i] = k
		i++
	}

	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	return keys
}

func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
}
