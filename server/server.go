package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	apiURL         = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	processTimeout = 3 * time.Second
	clientTimeout  = 2 * time.Second // timeout for the client
	InsertQuery    = `INSERT INTO cotacoes (bid) VALUES (?)`
)

// CurrencyRates structure for API response
type CurrencyRates struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func ExchangeRateHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("starting request")
	defer log.Println("request finalized")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rate, err := fetchExchangeRate()
	if err != nil {
		http.Error(w, "failed to fetch exchange rate", http.StatusInternalServerError)
		fmt.Printf("error fetching exchange rate: %v\n", err)
		return
	}

	if err := saveExchangeRateToDB(db, rate); err != nil {
		http.Error(w, "Failed to save exchange rate", http.StatusInternalServerError)
		fmt.Printf("Error saving exchange rate to database: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"bid": rate})
}

func fetchExchangeRate() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), processTimeout) // function execution timeout return context deadline exceeded
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: clientTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch exchange rate: %w", err)
	}
	defer resp.Body.Close()

	var rates CurrencyRates
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return rates.USDBRL.Bid, nil
}

func saveExchangeRateToDB(db *sql.DB, bid string) error {
	_, err := db.Exec(InsertQuery, bid)
	if err != nil {
		return fmt.Errorf("failed to insert exchange rate: %w", err)
	}
	return nil
}
