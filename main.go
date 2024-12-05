package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/bianavic/go-exchange-rate/client"
	"github.com/bianavic/go-exchange-rate/config"
	"github.com/bianavic/go-exchange-rate/server"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"time"
)

const (
	dbTimeout     = 10 * time.Millisecond // Timeout for the database operation (10ms)
	serverPort    = ":8080"
	clientTimeout = 2 * time.Second // timeout for the client
)

var (
	logger config.Logger
)

func main() {

	logger = *config.GetLogger("main")

	db, err := config.InitDB()
	if err != nil {
		logger.Errorf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	go startServer(db)

	if err := waitForServerReady("http://localhost" + serverPort); err != nil {
		logger.Errorf("Server not ready: %v", err)
		return
	}

	// get the exchange rate from the local server
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	rate, err := client.GetExchangeRate(ctx)
	if err != nil {
		logger.Errorf("error getting exchange rate: %v ", err)
		return
	}

	if err := client.SaveToFile(rate.Bid); err != nil {
		logger.Errorf("error saving to file: %v\n ", err)
		return
	}

	fmt.Println("exchange rate saved to cotacao.txt")
}

func startServer(db *sql.DB) {
	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		server.ExchangeRateHandler(w, r, db)
	})
	fmt.Printf("Server running on http://localhost%s/cotacao\n", serverPort)
	if err := http.ListenAndServe(serverPort, nil); err != nil {
		logger.Errorf("failed to start server: %v\n ", err)
	}
}

// allow the server some time to start before the client makes a request
func waitForServerReady(serverURL string) error {
	timeout := time.After(10 * time.Second)   // wait server max time
	tick := time.Tick(100 * time.Millisecond) // time interval retries

	for {
		select {
		case <-timeout:
			return fmt.Errorf("server did not start within the timeout")
		case <-tick:
			resp, err := http.Get(serverURL + "/cotacao")
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return nil // server ready
				}
			}
		}
	}
}
