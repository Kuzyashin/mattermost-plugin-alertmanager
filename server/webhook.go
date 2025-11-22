package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/hako/durafmt"
	"github.com/prometheus/alertmanager/notify/webhook"
	"github.com/prometheus/alertmanager/template"

	"github.com/mattermost/mattermost-server/v6/model"
)

const (
	alertStatusResolved = "resolved"
)

func (p *Plugin) handleWebhook(w http.ResponseWriter, r *http.Request, alertConfig alertConfig) {
	p.API.LogInfo("[WEBHOOK] Received alertmanager notification",
		"config_id", alertConfig.ID,
		"config_team", alertConfig.Team,
		"config_channel", alertConfig.Channel,
		"remote_addr", r.RemoteAddr,
	)

	var message webhook.Message
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		p.API.LogError("[WEBHOOK] Failed to decode webhook message",
			"config_id", alertConfig.ID,
			"error", err.Error(),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p.API.LogInfo("[WEBHOOK] Decoded webhook message",
		"config_id", alertConfig.ID,
		"status", message.Status,
		"receiver", message.Receiver,
		"num_alerts", len(message.Alerts),
	)

	if message == (webhook.Message{}) {
		p.API.LogWarn("[WEBHOOK] Received empty webhook message",
			"config_id", alertConfig.ID,
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Determine target channel
	channelID := p.AlertConfigIDChannelID[alertConfig.ID]
	if channelID == "" {
		p.API.LogError("[WEBHOOK] No channel mapping found for config",
			"config_id", alertConfig.ID,
			"config_channel", alertConfig.Channel,
			"available_mappings", fmt.Sprintf("%+v", p.AlertConfigIDChannelID),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Process each alert separately
	for i, alert := range message.Alerts {
		fingerprint := alert.Fingerprint
		p.API.LogDebug("[WEBHOOK] Processing alert",
			"config_id", alertConfig.ID,
			"alert_index", i,
			"alert_status", alert.Status,
			"fingerprint", fingerprint,
			"alert_labels", fmt.Sprintf("%+v", alert.Labels),
		)

		if alert.Status == alertStatusResolved {
			// Handle resolved alert - update existing post
			p.handleResolvedAlert(alertConfig, alert, message.ExternalURL, message.Receiver, channelID)
		} else {
			// Handle firing alert - create new post
			p.handleFiringAlert(alertConfig, alert, message.ExternalURL, message.Receiver, channelID)
		}
	}

	p.API.LogInfo("[WEBHOOK] Successfully processed all alerts",
		"config_id", alertConfig.ID,
		"channel_id", channelID,
		"num_alerts", len(message.Alerts),
	)
	w.WriteHeader(http.StatusOK)
}

func (p *Plugin) handleFiringAlert(alertConfig alertConfig, alert template.Alert, externalURL, receiver, channelID string) {
	fingerprint := alert.Fingerprint

	// Check if we already have a post for this alert
	existingPostID, err := p.getAlertPost(fingerprint)
	if err != nil {
		p.API.LogError("[WEBHOOK] Failed to check existing alert post",
			"fingerprint", fingerprint,
			"error", err.Error(),
		)
	}

	if existingPostID != "" {
		p.API.LogDebug("[WEBHOOK] Alert already has a post, skipping",
			"fingerprint", fingerprint,
			"post_id", existingPostID,
		)
		return
	}

	post := &model.Post{
		ChannelId: channelID,
		UserId:    p.BotUserID,
	}

	// Add severity mentions if configured
	severity := alert.Labels["severity"]
	if severity != "" && alertConfig.SeverityMentions != nil {
		if mentions, ok := alertConfig.SeverityMentions[severity]; ok && mentions != "" {
			post.Message = mentions
		}
	}

	// Create attachment
	var attachment *model.SlackAttachment

	// Use custom template if configured
	if alertConfig.FiringTemplate != "" {
		customMsg, err := renderAlertTemplate(alertConfig.FiringTemplate, alert)
		if err != nil {
			p.API.LogError("[WEBHOOK] Failed to render custom template",
				"error", err.Error(),
				"fingerprint", fingerprint,
			)
			// Fall back to default formatting
			fields := ConvertAlertToFields(alertConfig, alert, externalURL, receiver)
			attachment = &model.SlackAttachment{
				Fields: fields,
				Color:  colorFiring,
			}
		} else {
			// Append custom template to message
			if post.Message != "" {
				post.Message += "\n\n" + customMsg
			} else {
				post.Message = customMsg
			}
			// Create minimal attachment for action buttons
			attachment = &model.SlackAttachment{
				Color: colorFiring,
			}
		}
	} else {
		// Use default attachment formatting
		fields := ConvertAlertToFields(alertConfig, alert, externalURL, receiver)
		attachment = &model.SlackAttachment{
			Fields: fields,
			Color:  colorFiring,
		}
	}

	// Add action buttons if enabled
	if alertConfig.EnableActions {
		config := p.API.GetConfig()
		if config == nil || config.ServiceSettings.SiteURL == nil || *config.ServiceSettings.SiteURL == "" {
			p.API.LogWarn("[WEBHOOK] SiteURL is not configured, action buttons will not work",
				"config_nil", config == nil,
			)
		} else {
			siteURL := *config.ServiceSettings.SiteURL
			actions := []*model.PostAction{
				{
					Name: "🔕 Silence 1h",
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						URL: fmt.Sprintf("%s/plugins/%s/api/action", siteURL, Manifest.Id),
						Context: map[string]interface{}{
							"action":      "silence",
							"fingerprint": fingerprint,
							"config_id":   alertConfig.ID,
							"duration":    "1h",
						},
					},
				},
				{
					Name: "🔕 Silence 4h",
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						URL: fmt.Sprintf("%s/plugins/%s/api/action", siteURL, Manifest.Id),
						Context: map[string]interface{}{
							"action":      "silence",
							"fingerprint": fingerprint,
							"config_id":   alertConfig.ID,
							"duration":    "4h",
						},
					},
				},
				{
					Name: "👁️ ACK",
					Type: model.PostActionTypeButton,
					Integration: &model.PostActionIntegration{
						URL: fmt.Sprintf("%s/plugins/%s/api/action", siteURL, Manifest.Id),
						Context: map[string]interface{}{
							"action":      "ack",
							"fingerprint": fingerprint,
						},
					},
				},
			}
			attachment.Actions = actions
		}
	}

	model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})
	createdPost, appErr := p.API.CreatePost(post)
	if appErr != nil {
		p.API.LogError("[WEBHOOK] Failed to create post for firing alert",
			"channel_id", channelID,
			"fingerprint", fingerprint,
			"error", appErr.Error(),
		)
		return
	}

	if createdPost == nil {
		p.API.LogError("[WEBHOOK] CreatePost returned nil post without error",
			"channel_id", channelID,
			"fingerprint", fingerprint,
		)
		return
	}

	// Save the mapping
	if err := p.saveAlertPost(fingerprint, createdPost.Id); err != nil {
		p.API.LogError("[WEBHOOK] Failed to save alert post mapping",
			"fingerprint", fingerprint,
			"post_id", createdPost.Id,
			"error", err.Error(),
		)
	}

	p.API.LogInfo("[WEBHOOK] Created post for firing alert",
		"fingerprint", fingerprint,
		"post_id", createdPost.Id,
		"channel_id", channelID,
	)
}

func (p *Plugin) handleResolvedAlert(alertConfig alertConfig, alert template.Alert, externalURL, receiver, channelID string) {
	fingerprint := alert.Fingerprint

	// Find the original post
	originalPostID, err := p.getAlertPost(fingerprint)
	if err != nil {
		p.API.LogError("[WEBHOOK] Failed to get original alert post",
			"fingerprint", fingerprint,
			"error", err.Error(),
		)
		return
	}

	if originalPostID == "" {
		p.API.LogWarn("[WEBHOOK] No original post found for resolved alert, creating new one",
			"fingerprint", fingerprint,
		)
		// Create a new post for resolved alert if original not found
		p.handleFiringAlert(alertConfig, alert, externalURL, receiver, channelID)
		return
	}

	// Get the original post
	originalPost, appErr := p.API.GetPost(originalPostID)
	if appErr != nil {
		p.API.LogError("[WEBHOOK] Failed to retrieve original post",
			"post_id", originalPostID,
			"fingerprint", fingerprint,
			"error", appErr.Error(),
		)
		return
	}

	// Update the original post with resolved status
	originalPost.Message = ""
	originalPost.Props = make(model.StringInterface)

	var attachment *model.SlackAttachment

	// Use custom template if configured
	if alertConfig.ResolvedTemplate != "" {
		customMsg, err := renderAlertTemplate(alertConfig.ResolvedTemplate, alert)
		if err != nil {
			p.API.LogError("[WEBHOOK] Failed to render custom resolved template",
				"error", err.Error(),
				"fingerprint", fingerprint,
			)
			// Fall back to default formatting
			fields := ConvertAlertToFieldsResolved(alertConfig, alert, externalURL, receiver)
			attachment = &model.SlackAttachment{
				Fields: fields,
				Color:  colorResolved,
			}
		} else {
			originalPost.Message = customMsg
			// Create minimal attachment for color
			attachment = &model.SlackAttachment{
				Color: colorResolved,
			}
		}
	} else {
		// Use default attachment formatting
		fields := ConvertAlertToFieldsResolved(alertConfig, alert, externalURL, receiver)
		attachment = &model.SlackAttachment{
			Fields: fields,
			Color:  colorResolved,
		}
	}

	model.ParseSlackAttachment(originalPost, []*model.SlackAttachment{attachment})

	if _, appErr := p.API.UpdatePost(originalPost); appErr != nil {
		p.API.LogError("[WEBHOOK] Failed to update post for resolved alert",
			"post_id", originalPostID,
			"fingerprint", fingerprint,
			"error", appErr.Error(),
		)
		return
	}

	// Create a thread reply with timing information
	duration := alert.EndsAt.Sub(alert.StartsAt)
	threadMessage := fmt.Sprintf(
		"✅ **Alert Resolved**\n\n"+
			"**Fired at:** %s\n"+
			"**Resolved at:** %s\n"+
			"**Duration:** %s",
		alert.StartsAt.Format(time.RFC1123),
		alert.EndsAt.Format(time.RFC1123),
		durafmt.Parse(duration).LimitFirstN(2).String(),
	)

	threadPost := &model.Post{
		ChannelId: channelID,
		UserId:    p.BotUserID,
		RootId:    originalPostID,
		Message:   threadMessage,
	}

	if _, appErr := p.API.CreatePost(threadPost); appErr != nil {
		p.API.LogError("[WEBHOOK] Failed to create thread post for resolved alert",
			"post_id", originalPostID,
			"fingerprint", fingerprint,
			"error", appErr.Error(),
		)
		return
	}

	// Delete the mapping as alert is resolved
	if err := p.deleteAlertPost(fingerprint); err != nil {
		p.API.LogWarn("[WEBHOOK] Failed to delete alert post mapping",
			"fingerprint", fingerprint,
			"error", err.Error(),
		)
	}

	p.API.LogInfo("[WEBHOOK] Updated post for resolved alert",
		"fingerprint", fingerprint,
		"post_id", originalPostID,
		"duration", duration.String(),
	)
}

func addFields(fields []*model.SlackAttachmentField, title, msg string, short bool) []*model.SlackAttachmentField {
	return append(fields, &model.SlackAttachmentField{
		Title: title,
		Value: msg,
		Short: model.SlackCompatibleBool(short),
	})
}

func setColor(impact string) string {
	mapImpactColor := map[string]string{
		"firing":   colorFiring,
		"resolved": colorResolved,
	}

	if val, ok := mapImpactColor[impact]; ok {
		return val
	}

	return colorExpired
}

func ConvertAlertToFields(config alertConfig, alert template.Alert, externalURL, receiver string) []*model.SlackAttachmentField {
	var fields []*model.SlackAttachmentField

	statusMsg := strings.ToUpper(alert.Status)
	if alert.Status == "firing" {
		statusMsg = fmt.Sprintf(":fire: %s :fire:", strings.ToUpper(alert.Status))
	}

	/* first field: Annotations, Start/End, Source */
	var msg string
	annotations := make([]string, 0, len(alert.Annotations))
	for k := range alert.Annotations {
		annotations = append(annotations, k)
	}
	sort.Strings(annotations)
	for _, k := range annotations {
		msg = fmt.Sprintf("%s**%s:** %s\n", msg, cases.Title(language.Und, cases.NoLower).String(k), alert.Annotations[k])
	}
	msg = fmt.Sprintf("%s \n", msg)
	msg = fmt.Sprintf("%s**Started at:** %s (%s ago)\n", msg,
		(alert.StartsAt).Format(time.RFC1123),
		durafmt.Parse(time.Since(alert.StartsAt)).LimitFirstN(2).String(),
	)
	if alert.Status == alertStatusResolved {
		msg = fmt.Sprintf("%s**Ended at:** %s (%s ago)\n", msg,
			(alert.EndsAt).Format(time.RFC1123),
			durafmt.Parse(time.Since(alert.EndsAt)).LimitFirstN(2).String(),
		)
	}
	msg = fmt.Sprintf("%s \n", msg)
	msg = fmt.Sprintf("%sGenerated by a [Prometheus Alert](%s) and sent to the [Alertmanager](%s) '%s' receiver.", msg, alert.GeneratorURL, externalURL, receiver)
	fields = addFields(fields, statusMsg, msg, true)

	/* second field: Labels only */
	msg = ""
	alert.Labels["AlertManager Config ID"] = config.ID
	labels := make([]string, 0, len(alert.Labels))
	for k := range alert.Labels {
		labels = append(labels, k)
	}
	sort.Strings(labels)
	for _, k := range labels {
		msg = fmt.Sprintf("%s**%s:** %s\n", msg, cases.Title(language.Und, cases.NoLower).String(k), alert.Labels[k])
	}

	fields = addFields(fields, "", msg, true)

	return fields
}

func ConvertAlertToFieldsResolved(config alertConfig, alert template.Alert, externalURL, receiver string) []*model.SlackAttachmentField {
	var fields []*model.SlackAttachmentField

	statusMsg := "✅ RESOLVED ✅"

	/* first field: Annotations, Start/End, Source */
	var msg string
	annotations := make([]string, 0, len(alert.Annotations))
	for k := range alert.Annotations {
		annotations = append(annotations, k)
	}
	sort.Strings(annotations)
	for _, k := range annotations {
		msg = fmt.Sprintf("%s**%s:** %s\n", msg, cases.Title(language.Und, cases.NoLower).String(k), alert.Annotations[k])
	}
	msg = fmt.Sprintf("%s \n", msg)
	msg = fmt.Sprintf("%s**Started at:** %s\n", msg, alert.StartsAt.Format(time.RFC1123))
	msg = fmt.Sprintf("%s**Ended at:** %s\n", msg, alert.EndsAt.Format(time.RFC1123))
	duration := alert.EndsAt.Sub(alert.StartsAt)
	msg = fmt.Sprintf("%s**Duration:** %s\n", msg, durafmt.Parse(duration).LimitFirstN(2).String())
	msg = fmt.Sprintf("%s \n", msg)
	msg = fmt.Sprintf("%sGenerated by a [Prometheus Alert](%s) and sent to the [Alertmanager](%s) '%s' receiver.", msg, alert.GeneratorURL, externalURL, receiver)
	fields = addFields(fields, statusMsg, msg, true)

	/* second field: Labels only */
	msg = ""
	alert.Labels["AlertManager Config ID"] = config.ID
	labels := make([]string, 0, len(alert.Labels))
	for k := range alert.Labels {
		labels = append(labels, k)
	}
	sort.Strings(labels)
	for _, k := range labels {
		msg = fmt.Sprintf("%s**%s:** %s\n", msg, cases.Title(language.Und, cases.NoLower).String(k), alert.Labels[k])
	}

	fields = addFields(fields, "", msg, true)

	return fields
}
