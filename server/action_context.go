package main

// ActionContext passed from action buttons
type ActionContext struct {
	Labels      map[string]interface{} `json:"labels"` // Alert labels for silence matchers
	SilenceID   string                 `json:"silence_id"`
	UserID      string                 `json:"user_id"`
	Action      string                 `json:"action"`
	Fingerprint string                 `json:"fingerprint"`
	ConfigID    string                 `json:"config_id"`
	Duration    string                 `json:"duration"` // For silence: 1h, 4h, 12h, 24h
	Severity    string                 `json:"severity"` // Alert severity for color mapping
}

// Action type for decoding action buttons
type Action struct {
	Context *ActionContext `json:"context"`
	UserID  string         `json:"user_id"`
	PostID  string         `json:"post_id"`
}
