package bot

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"

	"github.com/Abhishek48Shah/bot/internal/util"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

const maxBodySize = 10 << 10

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}
func (h *Handler) HandleChatStream(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		util.WriteError(w, r, util.UnsupportedMediaTypeErr("unsupported media type", fmt.Errorf("parse media type: %w", err)), h.logger)
		return
	}
	if mediaType != "application/json" {
		util.WriteError(w, r, util.UnsupportedMediaTypeErr("unsupported media type", fmt.Errorf("expected application/json, got: %q", mediaType)), h.logger)
		return
	}
	var message RequestInput
	if err := util.ReadJSON(w, r, &message, maxBodySize); err != nil {
		util.WriteError(w, r, util.BadRequestErr(err.Error(), fmt.Errorf("read json: %w", err)), h.logger)
		return
	}
	response, err := h.service.ChatStream(r.Context(), &message)
	if err != nil {
		if errors.Is(err, ErrNoUserMessage) {
			util.WriteError(w, r, util.BadRequestErr("at least one non-empty user message is required", err), h.logger)
			return
		}
		util.WriteError(w, r, util.InternalServerErr("internal server error", fmt.Errorf("chat stream: %w", err)), h.logger)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		var ollamaErr struct {
			Error string `json:"error"`
		}
		body, _ := io.ReadAll(response.Body)
		_ = json.Unmarshal(body, &ollamaErr)
		util.WriteError(w, r, util.BadGateWayErr("the upstream server returned an invalid or unparsable response", fmt.Errorf("ollama returned status: %d: %s", response.StatusCode, ollamaErr.Error)), h.logger)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		util.WriteError(w, r, util.InternalServerErr("internal server error", fmt.Errorf("type assertion: %w", err)), h.logger)
		return
	}
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		var ollamaChatResponse OllamaChatResponse
		if err := json.Unmarshal(scanner.Bytes(), &ollamaChatResponse); err != nil {
			continue
		}
		if ollamaChatResponse.Done {
			fmt.Fprint(w, "event: done\ndata: {}\n\n")
			flusher.Flush()
			break
		}
		payload, err := json.Marshal(map[string]string{"content": ollamaChatResponse.Message.Content})
		if err != nil {
			continue
		}
		fmt.Fprintf(w, "event: token\ndata: %s\n\n", payload)
		flusher.Flush()
	}
	if err := scanner.Err(); err != nil {
		h.logger.Error("stream interrupted", "error", err)
		fmt.Fprintf(w, "event: error\ndata: {\"message\":\"stream interrupted\"}\n\n")
		flusher.Flush()
		return
	}

}
