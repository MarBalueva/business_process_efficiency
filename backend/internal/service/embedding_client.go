package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type EmbeddingClient struct {
	enabled  bool
	baseURL  string
	http     *http.Client
}

func NewEmbeddingClient(enabled bool, baseURL string, timeout time.Duration) *EmbeddingClient {
	return &EmbeddingClient{
		enabled:  enabled,
		baseURL:  strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *EmbeddingClient) Enabled() bool {
	return c != nil && c.enabled
}

type embeddingResponse struct {
	Vectors [][]float64 `json:"vectors"`
	Dim     int         `json:"dim"`
}

func (c *EmbeddingClient) EmbedTexts(ctx context.Context, texts []string) ([][]float64, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("embedding client is disabled")
	}
	if len(texts) == 0 {
		return [][]float64{}, nil
	}

	body, err := json.Marshal(map[string][]string{
		"texts": texts,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	log.Printf(
		"embedding request: url=%s inputs=%d firstPreview=%q",
		c.baseURL+"/embed",
		len(texts),
		truncateForLog(texts[0], 120),
	)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding service returned status %d", resp.StatusCode)
	}
	if len(out.Vectors) != len(texts) {
		return nil, fmt.Errorf("embedding vectors count mismatch: got %d, want %d", len(out.Vectors), len(texts))
	}

	for i := range out.Vectors {
		if len(out.Vectors[i]) == 0 {
			return nil, fmt.Errorf("empty embedding at index %d", i)
		}
	}

	return out.Vectors, nil
}

func truncateForLog(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max] + "..."
}
