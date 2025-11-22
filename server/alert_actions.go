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

	// Update post buttons - replace ACK with UNACK
	updatedAttachments := p.updateActionButtons(post, fingerprint, "ack_to_unack")

	// Update the post and return the update response
	if updatedAttachments != nil {
		// Save original message
		originalMessage := post.Message

		// Clear the post and re-parse with updated attachments
		post.Message = originalMessage
		post.Props = make(model.StringInterface)

		// Use ParseSlackAttachment like webhook does
		model.ParseSlackAttachment(post, updatedAttachments)

		// Restore message in case ParseSlackAttachment cleared it
		if originalMessage != "" && post.Message == "" {
			post.Message = originalMessage
		}

		// Update via API to persist changes
		if _, appErr := p.API.UpdatePost(post); appErr != nil {
			p.API.LogError("[ACTION] Failed to update post", "error", appErr.Error())
		}

		// Return update in response for immediate UI update
		w.WriteHeader(http.StatusOK)
		response := model.PostActionIntegrationResponse{
			Update: post,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
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

	// Update post buttons - replace UNACK with ACK
	updatedAttachments := p.updateActionButtons(post, fingerprint, "unack_to_ack")

	// Update the post and return the update response
	if updatedAttachments != nil {
		// Save original message
		originalMessage := post.Message

		// Clear the post and re-parse with updated attachments
		post.Message = originalMessage
		post.Props = make(model.StringInterface)

		// Use ParseSlackAttachment like webhook does
		model.ParseSlackAttachment(post, updatedAttachments)

		// Restore message in case ParseSlackAttachment cleared it
		if originalMessage != "" && post.Message == "" {
			post.Message = originalMessage
		}

		// Update via API to persist changes
		if _, appErr := p.API.UpdatePost(post); appErr != nil {
			p.API.LogError("[ACTION] Failed to update post", "error", appErr.Error())
		}

		// Return update in response for immediate UI update
		w.WriteHeader(http.StatusOK)
		response := model.PostActionIntegrationResponse{
			Update: post,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

// updateActionButtons updates the action buttons in a post
// mode: "ack_to_unack" or "unack_to_ack"
func (p *Plugin) updateActionButtons(post *model.Post, fingerprint, mode string) []*model.SlackAttachment {
	// Get current attachments - they might be stored as []interface{}
	attachmentsProp := post.GetProp("attachments")
	if attachmentsProp == nil {
		p.API.LogWarn("[ACTION] No attachments prop found in post", "post_id", post.Id)
		return nil
	}

	// Debug: log what type we got
	p.API.LogDebug("[ACTION] Attachments prop type", "type", fmt.Sprintf("%T", attachmentsProp))

	// Try to convert from []interface{} to []*model.SlackAttachment
	var attachments []*model.SlackAttachment

	// First try direct cast
	if att, ok := attachmentsProp.([]*model.SlackAttachment); ok {
		attachments = att
	} else if attSlice, ok := attachmentsProp.([]interface{}); ok {
		// Convert from []interface{}
		for _, item := range attSlice {
			if slackAtt, ok := item.(*model.SlackAttachment); ok {
				attachments = append(attachments, slackAtt)
			} else if attMap, ok := item.(map[string]interface{}); ok {
				// Convert from map - extract what we need
				slackAtt := &model.SlackAttachment{}

				// Extract actions
				if actions, ok := attMap["actions"].([]interface{}); ok {
					for _, act := range actions {
						if actMap, ok := act.(map[string]interface{}); ok {
							action := &model.PostAction{}
							if name, ok := actMap["name"].(string); ok {
								action.Name = name
							}
							if typ, ok := actMap["type"].(string); ok {
								action.Type = typ
							}
							if integration, ok := actMap["integration"].(map[string]interface{}); ok {
								action.Integration = &model.PostActionIntegration{}
								if url, ok := integration["url"].(string); ok {
									action.Integration.URL = url
								}
								if ctx, ok := integration["context"].(map[string]interface{}); ok {
									action.Integration.Context = ctx
								}
							}
							slackAtt.Actions = append(slackAtt.Actions, action)
						}
					}
				}

				// Extract fields
				if fields, ok := attMap["fields"].([]interface{}); ok {
					for _, fld := range fields {
						if fldMap, ok := fld.(map[string]interface{}); ok {
							field := &model.SlackAttachmentField{}
							if title, ok := fldMap["title"].(string); ok {
								field.Title = title
							}
							if value, ok := fldMap["value"].(string); ok {
								field.Value = value
							}
							if short, ok := fldMap["short"].(bool); ok {
								field.Short = model.SlackCompatibleBool(short)
							}
							slackAtt.Fields = append(slackAtt.Fields, field)
						}
					}
				}

				// Extract color
				if color, ok := attMap["color"].(string); ok {
					slackAtt.Color = color
				}

				// Extract other common properties
				if title, ok := attMap["title"].(string); ok {
					slackAtt.Title = title
				}
				if text, ok := attMap["text"].(string); ok {
					slackAtt.Text = text
				}
				if fallback, ok := attMap["fallback"].(string); ok {
					slackAtt.Fallback = fallback
				}
				if pretext, ok := attMap["pretext"].(string); ok {
					slackAtt.Pretext = pretext
				}

				attachments = append(attachments, slackAtt)
			}
		}
	}

	if len(attachments) == 0 {
		p.API.LogWarn("[ACTION] No attachments found in post after conversion", "post_id", post.Id)
		return nil
	}

	// Get SiteURL for action buttons
	config := p.API.GetConfig()
	if config == nil || config.ServiceSettings.SiteURL == nil || *config.ServiceSettings.SiteURL == "" {
		p.API.LogWarn("[ACTION] SiteURL not configured")
		return attachments
	}
	siteURL := *config.ServiceSettings.SiteURL

	// Update the first attachment's actions
	if len(attachments) > 0 && attachments[0] != nil {
		var newActions []*model.PostAction

		for _, action := range attachments[0].Actions {
			if action == nil {
				continue
			}

			// Keep Silence buttons as-is
			if action.Integration != nil {
				if ctx, ok := action.Integration.Context["action"].(string); ok && ctx == "silence" {
					newActions = append(newActions, action)
					continue
				}
			}

			// Replace ACK/UNACK button based on mode
			if mode == "ack_to_unack" {
				// Skip the old ACK button, add UNACK instead
				if action.Integration != nil {
					if ctx, ok := action.Integration.Context["action"].(string); ok && ctx == "ack" {
						// Replace with UNACK button
						newActions = append(newActions, &model.PostAction{
							Name: "🔄 UNACK",
							Type: model.PostActionTypeButton,
							Integration: &model.PostActionIntegration{
								URL: fmt.Sprintf("%s/plugins/%s/api/action", siteURL, Manifest.Id),
								Context: map[string]interface{}{
									"action":      "unack",
									"fingerprint": fingerprint,
								},
							},
						})
						continue
					}
				}
			} else if mode == "unack_to_ack" {
				// Skip the old UNACK button, add ACK instead
				if action.Integration != nil {
					if ctx, ok := action.Integration.Context["action"].(string); ok && ctx == "unack" {
						// Replace with ACK button
						newActions = append(newActions, &model.PostAction{
							Name: "👁️ ACK",
							Type: model.PostActionTypeButton,
							Integration: &model.PostActionIntegration{
								URL: fmt.Sprintf("%s/plugins/%s/api/action", siteURL, Manifest.Id),
								Context: map[string]interface{}{
									"action":      "ack",
									"fingerprint": fingerprint,
								},
							},
						})
						continue
					}
				}
			}

			// Keep other buttons
			newActions = append(newActions, action)
		}

		// Update the attachment with new actions
		attachments[0].Actions = newActions

		// Update color and title when ACKed
		if mode == "ack_to_unack" {
			attachments[0].Color = "#FFAA00" // Yellow/Orange for acknowledged
			// Update title in first field if it exists
			if len(attachments[0].Fields) > 0 && attachments[0].Fields[0] != nil {
				attachments[0].Fields[0].Title = "👁️ ACKNOWLEDGED 👁️"
			}
		} else if mode == "unack_to_ack" {
			attachments[0].Color = colorFiring // Red for firing
			// Restore firing title in first field if it exists
			if len(attachments[0].Fields) > 0 && attachments[0].Fields[0] != nil {
				attachments[0].Fields[0].Title = "🔥 FIRING 🔥"
			}
		}
	}

	return attachments
}
