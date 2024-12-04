package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/bianavic/go-exchange-rate/client"
	"github.com/bianavic/go-exchange-rate/server"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

const (
	dbTimeout     = 10 * time.Millisecond // Timeout for the database operation (10ms)
	serverPort    = ":8080"
	clientTimeout = 2 * time.Second // timeout for the client
	dbPath        = "cotacoes.db"
	createQuery   = `
	CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
)

func main() {

	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// start the server
	go startServer(db)

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

func startServer(db *sql.DB) {
	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		server.ExchangeRateHandler(w, r, db)
	})
	fmt.Printf("Server running on http://localhost%s/cotacao\n", serverPort)
	if err := http.ListenAndServe(serverPort, nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec(createQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return db, nil
}
