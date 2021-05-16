package main

import (
	"log"
	"net/http"
	"os"
	"stockpricedemo/internal/app/rest"
	"strconv"
	"strings"
	"sync"

	_ "stockpricedemo/statik"

	"github.com/rakyll/statik/fs"
)

const (
	defaultHttpPort          string = "8008"
	unsupportedStatusMessage string = "Unsupported"
	upPath                   string = "/up"
)

var (
	httpPort    string = strings.TrimSpace(os.Getenv("HTTP_PORT"))
	serviceName string = strings.TrimSpace(os.Getenv("SERVICE_NAME"))

	stockSymbol  string
	numberOfDays int

	servicePath string = "/" + serviceName
	statikFS    http.FileSystem
	wg          = &sync.WaitGroup{}
)

func init() {
	if httpPort == "" {
		httpPort = defaultHttpPort
	}
}

func setEnvironment() {
	var err error
	numberOfDays, err = strconv.Atoi(strings.TrimSpace(os.Getenv("NDAYS")))
	if err != nil {
		log.Fatalf("Error converting the environment variable NDAYS due to %v:", err)
		os.Exit(1)
	}

	stockSymbol = strings.TrimSpace(os.Getenv("SYMBOL"))

	if stockSymbol == "" {
		log.Fatal("Must use the environment variable SYMBOL to specify the stock symbol for which prices are required")
		os.Exit(1)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s received a %s request via %v", serviceName, r.Method, r.URL)

	if r.URL.Path == upPath {
		w.WriteHeader(http.StatusOK)
	} else if r.URL.Path == servicePath {
		handleStockPriceRequests(w, r)
	} else {
		http.Error(w, unsupportedStatusMessage, http.StatusNotFound)
	}
}

func handleStockPriceRequests(w http.ResponseWriter, r *http.Request) {
	setEnvironment()

	switch r.Method {
	case http.MethodGet:
		log.Printf("Got stock price request: %v", w)
		rest.AllPricesAndMean(w, stockSymbol, numberOfDays)
		return
	default:
		http.Error(w, unsupportedStatusMessage, http.StatusBadRequest)
	}
}

func handleSwagger() {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	swaggerHandler := http.StripPrefix("/swaggerui/", staticServer)

	http.Handle("/swaggerui/", swaggerHandler)
}

func listen() {
	go func() {
		log.Print(http.ListenAndServe(":"+httpPort, nil))
	}()

	wg.Add(1)
}

func main() {
	handleSwagger()

	// TODO Replace with net/http.ServeMux (minimal), https://github.com/julienschmidt/httprouter (simple), or https://github.com/go-chi/chi.
	http.HandleFunc("/", handler)

	listen()
	log.Printf("%s REST service started.", serviceName)
	wg.Wait()
}
