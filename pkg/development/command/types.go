package command


type MarketSystemUpdate struct {
	Timestamp  int64             `json:"timestamp"`
	User       string            `json:"user"`
	NotifyType string            `json:"notify_type"`
	Point      string            `json:"point"`
	Extensions map[string]string `json:"extensions,omitempty"` // Additional extension information
}
