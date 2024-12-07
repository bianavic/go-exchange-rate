package main

import (
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
	serverPort = ":8080"
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

	for {
		select {
		case <-time.After(5 * time.Second):

			rate, err := client.GetExchangeRate()
			if err != nil {
				logger.Errorf("error getting exchange rate: %v ", err)
				return
			}

			if err := client.SaveToFile(rate.Bid); err != nil {
				logger.Errorf("error saving to file: %v\n ", err)
				return
			}

			fmt.Println("exchange rate saved to cotacao.txt")
			return
		}
	}
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
