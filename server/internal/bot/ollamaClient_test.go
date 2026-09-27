package bot

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestChatIdentifiesLatestUserMessageForLanguage(t *testing.T) {
	var request OllamaChatReqest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "{}")
	}))
	defer server.Close()

	host, portText, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		t.Fatalf("parse test server address: %v", err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatalf("parse test server port: %v", err)
	}
	client := NewOllamaClient(port, host, "test-model")

	messages := []Message{
		{Role: "user", Content: "I am planning to visit Bharatpur"},
		{Role: "assistant", Content: "Sounds good!"},
		{Role: "user", Content: "म भरतपुर घुम्न जाने योजना बनाउँदै छु"},
	}
	response, err := client.Chat(context.Background(), messages, "", "")
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	_ = response.Body.Close()

	if len(request.Messages) == 0 || request.Messages[0].Role != "system" {
		t.Fatal("request is missing its system prompt")
	}
	prompt := request.Messages[0].Content
	if !strings.Contains(prompt, messages[2].Content) {
		t.Fatal("system prompt does not identify the latest user message")
	}
	if !strings.Contains(prompt, "latest user message shown") {
		t.Fatal("system prompt does not prioritize the latest user message")
	}
}
