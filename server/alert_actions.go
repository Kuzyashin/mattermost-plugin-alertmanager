package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
)

// handleAlertAction processes action button clicks (Silence/ACK/UNACK)
func (p *Plugin) handleAlertAction(w http.ResponseWriter, r *http.Request) {
	var action Action
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		p.API.LogError("[ACTION] Failed to decode action",
			"error", err.Error(),
		)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if action.Context == nil {
		p.API.LogError("[ACTION] Action context is nil")
		http.Error(w, "Invalid action context", http.StatusBadRequest)
		return
	}

	p.API.LogInfo("[ACTION] Processing action",
		"action", action.Context.Action,
		"user_id", action.UserID,
		"fingerprint", action.Context.Fingerprint,
	)

	switch action.Context.Action {
	case "silence":
		p.handleSilenceAction(w, r, action)
	case "ack":
		p.handleAckAction(w, r, action)
	case "unack":
		p.handleUnackAction(w, r, action)
	default:
		p.API.LogWarn("[ACTION] Unknown action", "action", action.Context.Action)
		http.Error(w, "Unknown action", http.StatusBadRequest)
	}
}

func (p *Plugin) handleSilenceAction(w http.ResponseWriter, r *http.Request, action Action) {
	fingerprint := action.Context.Fingerprint
	configID := action.Context.ConfigID
	duration := action.Context.Duration

	// Get config
	config := p.getConfiguration()
	alertConfig, ok := config.AlertConfigs[configID]
	if !ok {
		p.API.LogError("[ACTION] Config not found", "config_id", configID)
		http.Error(w, "Config not found", http.StatusNotFound)
		return
	}

	// Parse duration
	dur, err := time.ParseDuration(duration)
	if err != nil {
		p.API.LogError("[ACTION] Invalid duration", "duration", duration, "error", err.Error())
		http.Error(w, "Invalid duration", http.StatusBadRequest)
		return
	}

	// Get user info
	user, appErr := p.API.GetUser(action.UserID)
	if appErr != nil {
		p.API.LogError("[ACTION] Failed to get user", "error", appErr.Error())
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	p.API.LogInfo("[ACTION] Creating silence",
		"fingerprint", fingerprint,
		"duration", duration,
		"user", user.Username,
		"alertmanager_url", alertConfig.AlertManagerURL,
	)

	// Update post to show it's silenced
	post, appErr := p.API.GetPost(action.PostID)
	if appErr != nil {
		p.API.LogError("[ACTION] Failed to get post", "error", appErr.Error())
		http.Error(w, "Failed to get post", http.StatusInternalServerError)
		return
	}

	// Add thread reply
	threadMessage := fmt.Sprintf(
		"🔕 **Silenced for %s**\n\nBy: @%s\nUntil: %s",
		duration,
		user.Username,
		time.Now().Add(dur).Format(time.RFC1123),
	)

	threadPost := &model.Post{
		ChannelId: post.ChannelId,
		UserId:    p.BotUserID,
		RootId:    post.Id,
		Message:   threadMessage,
	}

	if _, appErr := p.API.CreatePost(threadPost); appErr != nil {
		p.API.LogError("[ACTION] Failed to create thread post", "error", appErr.Error())
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"update": map[string]interface{}{
			"message": fmt.Sprintf("🔕 Silenced for %s by @%s", duration, user.Username),
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (p *Plugin) handleAckAction(w http.ResponseWriter, r *http.Request, action Action) {
	fingerprint := action.Context.Fingerprint

	// Get user info
	user, appErr := p.API.GetUser(action.UserID)
	if appErr != nil {
		p.API.LogError("[ACTION] Failed to get user", "error", appErr.Error())
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	// Save ACK
	if err := p.ackAlert(fingerprint, user.Id, user.Username); err != nil {
		p.API.LogError("[ACTION] Failed to ack alert", "error", err.Error())
		http.Error(w, "Failed to ack alert", http.StatusInternalServerError)
		return
	}

	// Get post and update it
	post, appErr := p.API.GetPost(action.PostID)
	if appErr != nil {
		p.API.LogError("[ACTION] Failed to get post", "error", appErr.Error())
		http.Error(w, "Failed to get post", http.StatusInternalServerError)
		return
	}

	// Add thread reply
	threadMessage := fmt.Sprintf(
		"👁️ **Alert Acknowledged**\n\nBy: @%s\nAt: %s",
		user.Username,
		time.Now().Format(time.RFC1123),
	)

	threadPost := &model.Post{
		ChannelId: post.ChannelId,
		UserId:    p.BotUserID,
		RootId:    post.Id,
		Message:   threadMessage,
	}

	if _, appErr := p.API.CreatePost(threadPost); appErr != nil {
		p.API.LogError("[ACTION] Failed to create thread post", "error", appErr.Error())
	}

	p.API.LogInfo("[ACTION] Alert acknowledged",
		"fingerprint", fingerprint,
		"user", user.Username,
	)

	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"update": map[string]interface{}{
			"message": fmt.Sprintf("👁️ Acknowledged by @%s", user.Username),
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (p *Plugin) handleUnackAction(w http.ResponseWriter, r *http.Request, action Action) {
	fingerprint := action.Context.Fingerprint

	// Get user info
	user, appErr := p.API.GetUser(action.UserID)
	if appErr != nil {
		p.API.LogError("[ACTION] Failed to get user", "error", appErr.Error())
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	// Remove ACK
	if err := p.unackAlert(fingerprint); err != nil {
		p.API.LogError("[ACTION] Failed to unack alert", "error", err.Error())
		http.Error(w, "Failed to unack alert", http.StatusInternalServerError)
		return
	}

	// Get post
	post, appErr := p.API.GetPost(action.PostID)
	if appErr != nil {
		p.API.LogError("[ACTION] Failed to get post", "error", appErr.Error())
		http.Error(w, "Failed to get post", http.StatusInternalServerError)
		return
	}

	// Add thread reply
	threadMessage := fmt.Sprintf(
		"🔄 **Alert Unacknowledged**\n\nBy: @%s\nAt: %s",
		user.Username,
		time.Now().Format(time.RFC1123),
	)

	threadPost := &model.Post{
		ChannelId: post.ChannelId,
		UserId:    p.BotUserID,
		RootId:    post.Id,
		Message:   threadMessage,
	}

	if _, appErr := p.API.CreatePost(threadPost); appErr != nil {
		p.API.LogError("[ACTION] Failed to create thread post", "error", appErr.Error())
	}

	p.API.LogInfo("[ACTION] Alert unacknowledged",
		"fingerprint", fingerprint,
		"user", user.Username,
	)

	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"update": map[string]interface{}{
			"message": "Alert unacknowledged",
		},
	}
	json.NewEncoder(w).Encode(response)
}
