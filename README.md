# AlertManager Plugin

This plugin is the [AlertManager](https://github.com/prometheus/alertmanager) bot for Mattermost.

Forked and inspired on https://github.com/metalmatze/alertmanager-bot the alertmanager for Telegram. Thanks so much [@metalmatze](https://github.com/metalmatze/)

## Features

- Receive the Alerts via webhook
- Can list existing alerts
- Can list existing silences
- Can expire a silence
- **NEW:** Reload channel mappings without plugin restart (`/alertmanager reload`)
- **NEW:** Display current configuration and channel mappings (`/alertmanager config`)
- **NEW:** Auto-reload channel mappings on configuration save
- **NEW:** Enhanced logging for webhook processing and troubleshooting

## Known Issues and Solutions

### Channel Routing Problem

**Problem:** Alerts from different AlertManager configurations (e.g., dev, ops, prod) may be sent to incorrect Mattermost channels even when webhook URLs and tokens are properly configured.

**Root Cause:** The plugin builds a mapping between configuration IDs and channel IDs during initialization. If this mapping becomes stale after configuration changes, alerts may be routed to wrong channels.

**Solutions:**

1. **Automatic Reload (Recommended):** The plugin now automatically reloads channel mappings when you save configuration changes in the Mattermost UI.

2. **Manual Reload:** Use `/alertmanager reload` to manually refresh channel mappings without restarting the plugin.

3. **Verify Configuration:** Use `/alertmanager config` to display current mappings and verify correctness.

**Debugging:** Enhanced logging tracks the complete webhook flow:
```
[HTTP] Incoming request → Token matching → Config identification
[WEBHOOK] Message received → Alert processing → Channel lookup → Post creation
```

## Commands

### `/alertmanager reload` 🆕
Reloads channel configuration mappings without restarting the plugin.

### `/alertmanager config` 🆕
Displays current AlertManager configurations with channel mappings, IDs, and token prefixes.

### Other commands
- `/alertmanager alerts` - List existing alerts
- `/alertmanager silences` - List existing silences
- `/alertmanager expire_silence [Config ID] [Silence ID]` - Expire a silence
- `/alertmanager status` - Show AlertManager version and uptime
- `/alertmanager help` - Show all commands
- `/alertmanager about` - Show build information

## Enhanced Logging 🆕

Detailed logging for troubleshooting:

**HTTP Logging:**
```
[HTTP] Incoming request: method=POST, path=/api/webhook
[HTTP] Token matched config: config_id=0
```

**Webhook Logging:**
```
[WEBHOOK] Received notification: config_id=0, config_channel=alerts-dev
[WEBHOOK] Sending alert to channel: channel_id=xxx
[WEBHOOK] Successfully posted alert
```

**Error Logging:**
```
[WEBHOOK] No channel mapping found: config_id=0
[HTTP] No matching token found
```

## Installation

[... остальная часть README остается без изменений ...]

## Multiple AlertManager Configurations Example
```yaml
# Mattermost Plugin: 3 configs with different tokens/channels
# Config #0: alerts-prod, token-prod-xxx
# Config #1: alerts-dev, token-dev-yyy  
# Config #2: alerts-ops, token-ops-zzz

# AlertManager config:
receivers:
  - name: 'mattermost-prod'
    webhook_configs:
      - url: 'https://chat.example.com/plugins/alertmanager/api/webhook?token=token-prod-xxx'
  - name: 'mattermost-dev'
    webhook_configs:
      - url: 'https://chat.example.com/plugins/alertmanager/api/webhook?token=token-dev-yyy'

route:
  receiver: 'mattermost-prod'
  routes:
    - receiver: 'mattermost-dev'
      matchers:
        - k8s_cluster_name = dev
```

Verify with `/alertmanager config` after setup.

## Troubleshooting

### Alerts going to wrong channels

1. Run `/alertmanager config` to verify mappings
2. Check logs for `[WEBHOOK]` entries
3. Run `/alertmanager reload` to refresh
4. Verify channel names match exactly

### Plugin not receiving webhooks

1. Check logs for `[HTTP]` entries
2. Verify webhook URL is accessible
3. Check token matches plugin config
4. Ensure bot is invited to channels