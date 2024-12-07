package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/bianavic/go-exchange-rate/config"
	"net/http"
	"time"
)

const (
	apiURL            = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	maxAPICallTimeout = 200 * time.Millisecond
	dbTimeout         = 10 * time.Millisecond
	InsertQuery       = `INSERT INTO cotacoes (bid) VALUES (?)`
)

var (
	logger *config.Logger
)

// CurrencyRates structure for API response
type CurrencyRates struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func ExchangeRateHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	logger = config.GetLogger("server")
	logger.Info("starting request")
	defer logger.Info("request finalized")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rate, err := fetchExchangeRate()
	if err != nil {
		http.Error(w, "failed to fetch exchange rate", http.StatusInternalServerError)
		logger.Errorf("error fetching exchange rate: %v\n", err)
		return
	}

	if err := saveExchangeRateToDB(db, rate); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Failed to save exchange rate", http.StatusInternalServerError)
			logger.Errorf("error saving exchange rate to database: %v\n", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"bid": rate})
}

// fetch from API
func fetchExchangeRate() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxAPICallTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Error("timeout: context deadline exceeded")
		} else {
			logger.Errorf("failed to create request: %v\n", err)
		}
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf("failed to fetch exchange rate: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	var rates CurrencyRates
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		logger.Errorf("failed to decode response: %v\n", err)
		return "", err
	}

	return rates.USDBRL.Bid, nil
}

func saveExchangeRateToDB(db *sql.DB, bid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, InsertQuery, bid)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Error("timeout: context deadline exceeded")
		} else {
			logger.Errorf("failed to insert exchange ratee: %v\n", err)
		}
		return err
	}
	logger.Infof("successfully stored exchange rate: %s", bid)
	return nil
}
