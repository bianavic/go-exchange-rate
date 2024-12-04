package main

import (
	"context"
	"fmt"
	"github.com/bianavic/go-exchange-rate/client"
	"github.com/bianavic/go-exchange-rate/server"
	"net/http"
	"time"
)

const (
	dbTimeout     = 10 * time.Millisecond // Timeout for the database operation (10ms)
	serverPort    = ":8080"
	clientTimeout = 2 * time.Second // timeout for the client
)

func main() {

	// start the server
	go startServer()

	// allow the server some time to start before the client makes a request
	time.Sleep(10 * time.Second)

	// get the exchange rate from the local server
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	rate, err := client.GetExchangeRate(ctx)
	if err != nil {
		fmt.Printf("error getting exchange rate: %v\n", err)
		return
	}

	// save the rate to a file
	if err := client.SaveToFile(rate.Bid); err != nil {
		fmt.Printf("error saving to file: %v\n", err)
		return
	}

	fmt.Println("exchange rate saved to cotacao.txt")
}

func startServer() {
	http.HandleFunc("/cotacao", server.ExchangeRateHandler)
	fmt.Printf("Server running on http://localhost%s/cotacao\n", serverPort)
	if err := http.ListenAndServe(serverPort, nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
