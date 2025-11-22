package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"

	root "github.com/cpanato/mattermost-plugin-alertmanager"
)

var (
	Manifest model.Manifest = root.Manifest
)

type Plugin struct {
	plugin.MattermostPlugin
	client *pluginapi.Client

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	// key - alert config id, value - existing or created channel id received from api
	AlertConfigIDChannelID map[string]string
	BotUserID              string

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex
}

// Helper functions for alert fingerprint -> post ID mapping
func (p *Plugin) getAlertPostKey(fingerprint string) string {
	return fmt.Sprintf("alert_post_%s", fingerprint)
}

func (p *Plugin) saveAlertPost(fingerprint, postID string) error {
	key := p.getAlertPostKey(fingerprint)
	appErr := p.API.KVSet(key, []byte(postID))
	if appErr != nil {
		return appErr
	}
	return nil
}

func (p *Plugin) getAlertPost(fingerprint string) (string, error) {
	key := p.getAlertPostKey(fingerprint)
	data, appErr := p.API.KVGet(key)
	if appErr != nil {
		return "", appErr
	}
	if data == nil {
		return "", nil
	}
	return string(data), nil
}

func (p *Plugin) deleteAlertPost(fingerprint string) error {
	key := p.getAlertPostKey(fingerprint)
	appErr := p.API.KVDelete(key)
	if appErr != nil {
		return appErr
	}
	return nil
}

// Helper functions for alert acknowledgment
func (p *Plugin) getAlertAckKey(fingerprint string) string {
	return fmt.Sprintf("alert_ack_%s", fingerprint)
}

type AlertAck struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Timestamp int64  `json:"timestamp"`
}

func (p *Plugin) ackAlert(fingerprint, userID, username string) error {
	ack := AlertAck{
		UserID:    userID,
		Username:  username,
		Timestamp: model.GetMillis(),
	}
	data, err := json.Marshal(ack)
	if err != nil {
		return err
	}
	key := p.getAlertAckKey(fingerprint)
	appErr := p.API.KVSet(key, data)
	if appErr != nil {
		return appErr
	}
	return nil
}

func (p *Plugin) unackAlert(fingerprint string) error {
	key := p.getAlertAckKey(fingerprint)
	appErr := p.API.KVDelete(key)
	if appErr != nil {
		return appErr
	}
	return nil
}

func (p *Plugin) getAlertAck(fingerprint string) (*AlertAck, error) {
	key := p.getAlertAckKey(fingerprint)
	data, appErr := p.API.KVGet(key)
	if appErr != nil {
		return nil, appErr
	}
	if data == nil {
		return nil, nil
	}
	var ack AlertAck
	if err := json.Unmarshal(data, &ack); err != nil {
		return nil, err
	}
	return &ack, nil
}

func (p *Plugin) OnDeactivate() error {
	return nil
}

func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API, p.Driver)
	botID, err := p.client.Bot.EnsureBot(&model.Bot{
		Username:    "alertmanagerbot",
		DisplayName: "AlertManager Bot",
		Description: "Created by the AlertManager plugin.",
	}, pluginapi.ProfileImagePath(filepath.Join("assets", "alertmanager-logo.png")))
	if err != nil {
		return fmt.Errorf("failed to ensure bot account: %w", err)
	}
	p.BotUserID = botID

	configuration := p.getConfiguration()
	p.AlertConfigIDChannelID = make(map[string]string)
	for k, alertConfig := range configuration.AlertConfigs {
		var channelID string
		channelID, err = p.ensureAlertChannelExists(alertConfig)
		if err != nil {
			p.API.LogWarn(fmt.Sprintf("Failed to ensure alert channel %v", k), "error", err.Error())
		} else {
			p.AlertConfigIDChannelID[alertConfig.ID] = channelID
		}
	}

	command, err := p.getCommand()
	if err != nil {
		return fmt.Errorf("failed to get command: %w", err)
	}

	err = p.API.RegisterCommand(command)
	if err != nil {
		return fmt.Errorf("failed to register command: %w", err)
	}

	return nil
}

func (p *Plugin) ensureAlertChannelExists(alertConfig alertConfig) (string, error) {
	if err := alertConfig.IsValid(); err != nil {
		return "", fmt.Errorf("alert Configuration is invalid: %w", err)
	}

	team, appErr := p.API.GetTeamByName(alertConfig.Team)
	if appErr != nil {
		return "", fmt.Errorf("failed to get team: %w", appErr)
	}

	channel, appErr := p.API.GetChannelByName(team.Id, alertConfig.Channel, false)
	if appErr != nil {
		if appErr.StatusCode == http.StatusNotFound {
			channelToCreate := &model.Channel{
				Name:        alertConfig.Channel,
				DisplayName: alertConfig.Channel,
				Type:        model.ChannelTypeOpen,
				TeamId:      team.Id,
				CreatorId:   p.BotUserID,
			}

			newChannel, errChannel := p.API.CreateChannel(channelToCreate)
			if errChannel != nil {
				return "", fmt.Errorf("failed to create alert channel: %w", errChannel)
			}

			return newChannel.Id, nil
		}
		return "", fmt.Errorf("failed to get existing alert channel: %w", appErr)
	}

	return channel.Id, nil
}

func (p *Plugin) reloadChannelMappings() error {
	p.API.LogInfo("Starting channel mappings reload")

	configuration := p.getConfiguration()
	newMapping := make(map[string]string)

	for k, alertConfig := range configuration.AlertConfigs {
		var channelID string
		var err error

		channelID, err = p.ensureAlertChannelExists(alertConfig)
		if err != nil {
			p.API.LogWarn(fmt.Sprintf("Failed to ensure alert channel %v during reload", k), "error", err.Error())
			continue
		}

		newMapping[alertConfig.ID] = channelID
		p.API.LogInfo(fmt.Sprintf("Mapped config %s to channel %s", alertConfig.ID, channelID))
	}

	p.AlertConfigIDChannelID = newMapping
	p.API.LogInfo("Channel mappings reload completed", "mappings", fmt.Sprintf("%+v", newMapping))

	return nil
}

func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.API.LogInfo("[HTTP] Incoming request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
	)

	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Mattermost AlertManager Plugin"))
		return
	}

	// Handle action buttons (no token required)
	if r.URL.Path == "/api/action" {
		p.handleAlertAction(w, r)
		return
	}

	invalidOrMissingTokenErr := "Invalid or missing token"
	token := r.URL.Query().Get("token")
	if token == "" {
		p.API.LogWarn("[HTTP] Request without token",
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)
		http.Error(w, invalidOrMissingTokenErr, http.StatusBadRequest)
		return
	}

	p.API.LogDebug("[HTTP] Token received",
		"token_prefix", token[:8],
		"path", r.URL.Path,
	)

	configuration := p.getConfiguration()
	for _, alertConfig := range configuration.AlertConfigs {
		if subtle.ConstantTimeCompare([]byte(token), []byte(alertConfig.Token)) == 1 {
			p.API.LogInfo("[HTTP] Token matched config",
				"config_id", alertConfig.ID,
				"path", r.URL.Path,
			)
			switch r.URL.Path {
			case "/api/webhook":
				p.handleWebhook(w, r, alertConfig)
			case "/api/expire":
				p.handleExpireAction(w, r, alertConfig)
			default:
				p.API.LogWarn("[HTTP] Unknown path",
					"path", r.URL.Path,
					"config_id", alertConfig.ID,
				)
				http.NotFound(w, r)
			}
			return
		}
	}

	p.API.LogWarn("[HTTP] No matching token found",
		"token_prefix", token[:8],
		"path", r.URL.Path,
	)
	http.Error(w, invalidOrMissingTokenErr, http.StatusBadRequest)
}
