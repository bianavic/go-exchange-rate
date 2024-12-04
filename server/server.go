package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	apiURL     = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	apiTimeout = 200 * time.Millisecond
)

// CurrencyRates structure for API response
type CurrencyRates struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func ExchangeRateHandler(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"bid": rate})
}

func fetchExchangeRate() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 3 * time.Minute, // timeout for the client
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
