package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// configuration captures the plugin's external configuration as exposed in the Mattermost server
// configuration, as well as values computed from the configuration. Any public fields will be
// deserialized from the Mattermost server configuration in OnConfigurationChange.
//
// As plugins are inherently concurrent (hooks being called asynchronously), and the plugin
// configuration can change at any time, access to the configuration must be synchronized. The
// strategy used in this plugin is to guard a pointer to the configuration, and clone the entire
// struct whenever it changes. You may replace this with whatever strategy you choose.
//
// If you add non-reference types to your configuration struct, be sure to rewrite Clone as a deep
// copy appropriate for your types.
type configuration struct {
	AlertConfigs map[string]alertConfig
}

type alertConfig struct {
	EnableActions    bool // Enable Silence/ACK/UNACK buttons
	ID               string
	Token            string
	Channel          string
	Team             string
	AlertManagerURL  string
	FiringTemplate   string              // Custom template for firing alerts
	ResolvedTemplate string              // Custom template for resolved alerts
	SeverityMentions SeverityMentionsMap // e.g. {"critical": "@devops-oncall", "warning": "@devops"}
}

// SeverityMentionsMap is a custom type that handles both string (JSON) and map unmarshaling
type SeverityMentionsMap map[string]string

// UnmarshalJSON implements custom unmarshaling to handle both string and map
func (s *SeverityMentionsMap) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a map first
	var m map[string]string
	if err := json.Unmarshal(data, &m); err == nil {
		*s = SeverityMentionsMap(m)
		return nil
	}

	// If that fails, try to unmarshal as a string (JSON string)
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	// Parse the JSON string
	if str == "" {
		*s = make(SeverityMentionsMap)
		return nil
	}

	var m2 map[string]string
	if err := json.Unmarshal([]byte(str), &m2); err != nil {
		return err
	}

	*s = SeverityMentionsMap(m2)
	return nil
}

func (ac *alertConfig) IsValid() error {
	if ac.Team == "" {
		return errors.New("must set a Team")
	}

	if ac.Channel == "" {
		return errors.New("must set a Channel")
	}

	if ac.Token == "" {
		return errors.New("must set a Token")
	}

	if ac.AlertManagerURL == "" {
		return errors.New("must set the AlertManager URL")
	}

	return nil
}

// Clone shallow copies the configuration. Your implementation may require a deep copy if
// your configuration has reference types.
func (c *configuration) Clone() *configuration {
	var clone configuration
	for k, v := range c.AlertConfigs {
		clone.AlertConfigs[k] = v
	}
	return &clone
}

// getConfiguration retrieves the active configuration under lock, making it safe to use
// concurrently. The active configuration may change underneath the client of this method, but
// the struct returned by this API call is considered immutable.
func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{
			AlertConfigs: make(map[string]alertConfig),
		}
	}

	return p.configuration
}

// setConfiguration replaces the active configuration under lock.
//
// Do not call setConfiguration while holding the configurationLock, as sync.Mutex is not
// reentrant. In particular, avoid using the plugin API entirely, as this may in turn trigger a
// hook back into the plugin. If that hook attempts to acquire this lock, a deadlock may occur.
//
// This method panics if setConfiguration is called with the existing configuration. This almost
// certainly means that the configuration was modified without being cloned and may result in
// an unsafe access.
func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		// Ignore assignment if the configuration struct is empty. Go will optimize the
		// allocation for same to point at the same memory address, breaking the check
		// above.
		if reflect.ValueOf(*configuration).NumField() == 0 {
			return
		}

		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = configuration
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	var configurationInstance = configuration{
		AlertConfigs: make(map[string]alertConfig),
	}

	// Load the public configuration fields from the Mattermost server configuration.
	if err := p.API.LoadPluginConfiguration(&configurationInstance); err != nil {
		return fmt.Errorf("failed to load plugin configuration: %w", err)
	}

	for id, alertConfigInstance := range configurationInstance.AlertConfigs {
		alertConfigInstance.ID = id
		alertConfigInstance.AlertManagerURL = strings.TrimRight(alertConfigInstance.AlertManagerURL, `/`)
		configurationInstance.AlertConfigs[id] = alertConfigInstance
	}

	p.setConfiguration(&configurationInstance)

	// Reload channel mappings after configuration change
	p.API.LogInfo("Configuration changed, reloading channel mappings")
	if err := p.reloadChannelMappings(); err != nil {
		p.API.LogError("Failed to reload channel mappings after configuration change", "error", err.Error())
	}

	return nil
}
