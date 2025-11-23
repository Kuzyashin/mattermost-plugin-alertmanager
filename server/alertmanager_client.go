package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Silence represents an AlertManager silence
type Silence struct {
	StartsAt  time.Time        `json:"startsAt"`
	EndsAt    time.Time        `json:"endsAt"`
	ID        string           `json:"id,omitempty"`
	CreatedBy string           `json:"createdBy"`
	Comment   string           `json:"comment"`
	Matchers  []SilenceMatcher `json:"matchers"`
}

// SilenceMatcher represents a matcher for a silence
type SilenceMatcher struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	IsRegex bool   `json:"isRegex"`
	IsEqual bool   `json:"isEqual"`
}

// SilenceResponse represents the response from AlertManager
type SilenceResponse struct {
	SilenceID string `json:"silenceID"`
}

// createSilence creates a silence in AlertManager
func (p *Plugin) createSilence(alertManagerURL string, labels map[string]interface{}, duration time.Duration, createdBy, comment string) (string, error) {
	// Build matchers from alert labels
	var matchers []SilenceMatcher
	for name, value := range labels {
		if strValue, ok := value.(string); ok {
			matchers = append(matchers, SilenceMatcher{
				Name:    name,
				Value:   strValue,
				IsRegex: false,
				IsEqual: true,
			})
		}
	}

	if len(matchers) == 0 {
		return "", fmt.Errorf("no valid matchers found in alert labels")
	}

	// Create silence object
	now := time.Now()
	silence := Silence{
		Matchers:  matchers,
		StartsAt:  now,
		EndsAt:    now.Add(duration),
		CreatedBy: createdBy,
		Comment:   comment,
	}

	// Marshal to JSON
	payload, err := json.Marshal(silence)
	if err != nil {
		return "", fmt.Errorf("failed to marshal silence: %w", err)
	}

	p.API.LogDebug("[SILENCE] Creating silence in AlertManager",
		"url", alertManagerURL,
		"matchers", fmt.Sprintf("%+v", matchers),
		"duration", duration.String(),
	)

	// Send request to AlertManager
	url := fmt.Sprintf("%s/api/v2/silences", alertManagerURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to AlertManager: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("AlertManager returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var silenceResp SilenceResponse
	if err := json.Unmarshal(body, &silenceResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	p.API.LogInfo("[SILENCE] Successfully created silence in AlertManager",
		"silence_id", silenceResp.SilenceID,
		"duration", duration.String(),
	)

	return silenceResp.SilenceID, nil
}
