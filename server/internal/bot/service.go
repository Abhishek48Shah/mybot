package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

var (
	ErrEmbedMessage  = errors.New("bot: failed to embed user message")
	ErrRetrieve      = errors.New("bot: failed to retrieve context")
	ErrNoUserMessage = errors.New("bot: no user message provided")
)

const (
	topK                = 5
	similarityThreshold = 0.5
)

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

type Service struct {
	storage  *Storage
	embedder Embedder
	ollama   *OllamaClient
}

func NewService(storage *Storage, embedder Embedder, ollama *OllamaClient) *Service {
	return &Service{storage: storage, embedder: embedder, ollama: ollama}
}

func (s *Service) ChatStream(ctx context.Context, input *RequestInput) (*http.Response, error) {
	latestUserMessage := latestUserMessage(input.Messages)
	if latestUserMessage == "" {
		return nil, ErrNoUserMessage
	}

	vec, err := s.embedder.Embed(ctx, latestUserMessage)
	if err != nil {
		return nil, errors.Join(ErrEmbedMessage, err)
	}

	chunks, err := s.storage.RetrieveChunks(ctx, vec, topK)
	if err != nil {
		return nil, errors.Join(ErrRetrieve, err)
	}

	retrievedContext := buildContext(chunks)

	return s.ollama.Chat(ctx, input.Messages, retrievedContext, input.Language)
}

func latestUserMessage(messages []Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" && strings.TrimSpace(messages[i].Content) != "" {
			return messages[i].Content
		}
	}
	return ""
}

func buildContext(chunks []RetrievedChunk) string {
	if len(chunks) == 0 {
		return ""
	}

	var b strings.Builder
	for _, c := range chunks {
		if c.Similarity < similarityThreshold {
			continue
		}
		b.WriteString(c.Text)
		b.WriteString("\n\n")
	}

	return b.String()
}
