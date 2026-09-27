package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrEmbedRequest  = errors.New("embedding: failed to build embed request")
	ErrEmbedCall     = errors.New("embedding: failed to call embedding server")
	ErrEmbedStatus   = errors.New("embedding: embedding server returned non-200")
	ErrEmbedDecode   = errors.New("embedding: failed to decode embedding response")
	ErrEmbedEmptyVec = errors.New("embedding: embedding server returned empty vector")
)

type Embedding struct {
	Host  string
	Port  int
	Model string

	client *http.Client
}

func NewEmbedding(host string, port int, model string) *Embedding {
	return &Embedding{
		Host:   host,
		Port:   port,
		Model:  model,
		client: &http.Client{},
	}
}

type embedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embedResponse struct {
	Embedding []float32 `json:"embedding"`
}

func (e *Embedding) Embed(ctx context.Context, text string) ([]float32, error) {
	url := fmt.Sprintf("http://%s:%d/api/embeddings", e.Host, e.Port)

	body, err := json.Marshal(embedRequest{Model: e.Model, Prompt: text})
	if err != nil {
		return nil, errors.Join(ErrEmbedRequest, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, errors.Join(ErrEmbedRequest, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, errors.Join(ErrEmbedCall, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrEmbedStatus, resp.StatusCode)
	}

	var out embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, errors.Join(ErrEmbedDecode, err)
	}

	if len(out.Embedding) == 0 {
		return nil, ErrEmbedEmptyVec
	}

	return out.Embedding, nil
}
