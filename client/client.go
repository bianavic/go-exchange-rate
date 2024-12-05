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
)

const (
	serverURL = "http://localhost:8080/cotacao"
)

var (
	logger *config.Logger
)

// local server response
type BidResponse struct {
	Bid string `json:"bid"`
}

// GetExchangeRate fetches the exchange rate from the local server
func GetExchangeRate(ctx context.Context) (*BidResponse, error) {

	logger = config.GetLogger("client")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		logger.Errorf("failed to create request: %v", err)
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
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
		return nil, err
	}

	var bidResp BidResponse
	if err := json.NewDecoder(resp.Body).Decode(&bidResp); err != nil {
		logger.Errorf("failed to fetch decode response: %v", err)
		return nil, err
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
