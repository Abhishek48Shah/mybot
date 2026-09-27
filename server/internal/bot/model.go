package bot

type RequestInput struct {
	Messages []Message `json:"messages"`
	Language string    `json:"language,omitempty"`
}
type OllamaChatReqest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type OllamaChatResponse struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}
