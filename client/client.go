package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bianavic/go-exchange-rate/config"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	serverURL             = "http://localhost:8080/cotacao"
	serverResponseTimeout = 300 * time.Millisecond
)

var (
	logger *config.Logger
)

// BidResponse local server response
type BidResponse struct {
	Bid string `json:"bid"`
}

// GetExchangeRate fetches the exchange rate from server
func GetExchangeRate() (*BidResponse, error) {
	logger = config.GetLogger("client")

	ctx, cancel := context.WithTimeout(context.Background(), serverResponseTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		logger.Errorf("failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logger.Errorf("context deadline exceeded")
			return nil, err
		}
		logger.Errorf("failed to fetch exchange rate: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var bidResp BidResponse
	if err := json.NewDecoder(resp.Body).Decode(&bidResp); err != nil {
		logger.Errorf("failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &bidResp, nil
}

func SaveToFile(rate string) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		logger.Errorf("failed to create file: %w", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dolar: %s", rate))
	if err != nil {
		logger.Errorf("failed to write to file: %w", err)
		return err
	}

	return nil
}
