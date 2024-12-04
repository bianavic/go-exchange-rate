package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	serverURL  = "http://localhost:8080/cotacao"
	serverPort = ":8080"
)

// local server response
type BidResponse struct {
	Bid string `json:"bid"`
}

// GetExchangeRate fetches the exchange rate from the local server
func GetExchangeRate(ctx context.Context) (*BidResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("context deadline exceeded")
		}
		return nil, fmt.Errorf("failed to fetch exchange rate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
	}

	var bidResp BidResponse
	if err := json.NewDecoder(resp.Body).Decode(&bidResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &bidResp, nil
}

func SaveToFile(rate string) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dolar: %s", rate))
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
