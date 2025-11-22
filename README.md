# AlertManager Plugin

This plugin is the [AlertManager](https://github.com/prometheus/alertmanager) bot for Mattermost.

Forked and inspired on https://github.com/metalmatze/alertmanager-bot the alertmanager for Telegram. Thanks so much [@metalmatze](https://github.com/metalmatze/)

## Features

- Receive the Alerts via webhook
- **NEW:** Smart alert lifecycle management - firing alerts create posts, resolved alerts update them ✅
- **NEW:** Thread replies with resolution timing information
- **NEW:** Interactive action buttons (Silence/ACK/UNACK) with dynamic button updates 🎯
- **NEW:** Visual alert states - color-coded posts (red=firing, yellow=acknowledged, green=resolved) 🎨
- **NEW:** Severity-based mentions (@team notifications) 📢
- **NEW:** Custom alert templates for firing and resolved alerts 📝
- Can list existing alerts
- Can list existing silences
- Can expire a silence
- **NEW:** Reload channel mappings without plugin restart (`/alertmanager reload`)
- **NEW:** Display current configuration and channel mappings (`/alertmanager config`)
- **NEW:** Auto-reload channel mappings on configuration save
- **NEW:** Enhanced logging for webhook processing and troubleshooting

## Alert Lifecycle Management 🆕

The plugin now intelligently handles alert state transitions:

### Firing Alert
When an alert fires, the plugin:
1. Creates a new post in the configured channel
2. Stores the alert fingerprint → post ID mapping
3. Uses red color (🔥 FIRING 🔥)

### Resolved Alert
When an alert resolves, the plugin:
1. Finds the original firing alert post
2. **Updates** the original post (changes color to green, status to ✅ RESOLVED)
3. Creates a **thread reply** with timing information:
   ```
   ✅ Alert Resolved

   Fired at: Thu, 21 Nov 2024 10:00:00 UTC
   Resolved at: Thu, 21 Nov 2024 10:10:00 UTC
   Duration: 10 minutes
   ```

This approach keeps your channels clean and makes it easy to see alert duration at a glance!

## Interactive Action Buttons 🆕

Enable interactive action buttons on alert posts for quick alert management:

### Available Actions
- **🔕 Silence 1h / 4h** - Silence the alert for 1 or 4 hours in AlertManager
- **👁️ ACK** - Acknowledge the alert (marks it as seen, changes color to yellow/orange)
- **🔄 UNACK** - Unacknowledge the alert (removes acknowledgment, returns to red)

### Configuration
Enable action buttons in the plugin configuration:

```json
{
  "EnableActions": true
}
```

When enabled, each firing alert will include action buttons. Clicking a button will:
1. Perform the action (silence/acknowledge/unacknowledge)
2. Add a thread reply with action details (who, when)
3. **Update the post** - change button states, colors, and status:
   - **ACK**: Button changes from "👁️ ACK" → "🔄 UNACK", color changes from red → yellow/orange, status changes from "🔥 FIRING 🔥" → "👁️ ACKNOWLEDGED 👁️"
   - **UNACK**: Button changes from "🔄 UNACK" → "👁️ ACK", color returns from yellow → red, status returns to "🔥 FIRING 🔥"

### Visual Indicators
- **🔥 FIRING (not acknowledged)**: Red color
- **👁️ ACKNOWLEDGED**: Yellow/orange color
- **✅ RESOLVED**: Green color

Example thread reply for ACK:
```
👁️ Alert Acknowledged

By: @johndoe
At: Thu, 21 Nov 2024 10:05:00 UTC
```

Example thread reply for UNACK:
```
🔄 Alert Unacknowledged

By: @johndoe
At: Thu, 21 Nov 2024 10:15:00 UTC
```

Example thread reply for Silence:
```
🔕 Silenced for 1h

By: @johndoe
Until: Thu, 21 Nov 2024 11:05:00 UTC
```

## Severity-Based Mentions 🆕

Automatically mention teams or users based on alert severity:

### Configuration
Configure mentions in the plugin settings:

```json
{
  "SeverityMentions": {
    "critical": "@devops-oncall @sre-team",
    "warning": "@devops",
    "info": ""
  }
}
```

### Behavior
When an alert fires, the plugin checks the alert's `severity` label and adds the configured mentions to the post message. This ensures critical alerts immediately notify the right people.

Example post:
```
@devops-oncall @sre-team

[Alert attachment with details...]
```

## Custom Alert Templates 🆕

Customize how alerts are displayed using Go templates:

### Configuration
Configure custom templates in the plugin settings:

```json
{
  "FiringTemplate": "🔥 **{{ .Labels.alertname }}**\n\n{{ if .Annotations.summary }}**Summary:** {{ .Annotations.summary }}{{ end }}\n**Severity:** {{ .Labels.severity }}\n**Started at:** {{ formatTime .StartsAt }}",

  "ResolvedTemplate": "✅ **{{ .Labels.alertname }} - RESOLVED**\n\n**Duration:** {{ .EndsAt.Sub .StartsAt }}"
}
```

### Template Functions
Available template functions:
- `formatTime` - Format time as RFC1123 (e.g., "Thu, 21 Nov 2024 10:00:00 UTC")
- `toUpper` - Convert string to uppercase
- Standard Go template functions

### Template Data
Templates have access to the full Prometheus alert object:
- `.Labels` - Alert labels (map[string]string)
- `.Annotations` - Alert annotations (map[string]string)
- `.StartsAt` - Alert start time
- `.EndsAt` - Alert end time (for resolved alerts)
- `.GeneratorURL` - Link to the alert in Prometheus
- `.Fingerprint` - Unique alert identifier

### Behavior
- If custom templates are configured, they replace the default attachment formatting
- If template rendering fails, the plugin falls back to default formatting
- Severity mentions and action buttons still work with custom templates

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
[WEBHOOK] Processing alert: fingerprint=abc123, alert_status=firing
[WEBHOOK] Created post for firing alert: post_id=xyz
[WEBHOOK] Updated post for resolved alert: duration=10m
```

**Error Logging:**
```
[WEBHOOK] No channel mapping found: config_id=0
[HTTP] No matching token found
```

**Supported Mattermost Server Versions: 5.37+**

## Installation

1. Go to the [releases page of this GitHub repository](https://github.com/Kuzyashin/mattermost-plugin-alertmanager/releases) and download the latest release for your Mattermost server.
2. Upload this file in the Mattermost **System Console > Plugins > Management** page to install the plugin, and enable it. To learn more about how to upload a plugin, [see the documentation](https://docs.mattermost.com/administration/plugins.html#plugin-uploads).

Next, to configure the plugin, follow these steps:

3. After you've uploaded the plugin in **System Console > Plugins > Management**, go to the plugin's settings page at **System Console > Plugins > AlertManager**.
4. Specify the team and channel to send messages to. For each, use the URL of the team or channel instead of their respective display names.
5. Specify the AlertManager Server URL.
6. Generate the Token that will be use to validate the requests.
7. Hit **Save** (the plugin will automatically reload channel mappings).
8. Next, copy the **Token** above the **Save** button, which is used to configure the plugin for your AlertManager account.
9. Go to your Alertmanager configuration, paste the following webhook URL and specify the name of the service and the token you copied in step 8.
10. Invite the `@alertmanagerbot` user to your target team and channel.

```
https://SITEURL/plugins/alertmanager/api/webhook?token=TOKEN
```
Sometimes the token has to be quoted.

Example alertmanager config:

```yaml
webhook_configs:
  - send_resolved: true  # IMPORTANT: Set to true to enable resolved alert updates
    url: "https://mattermost.example.org/plugins/alertmanager/api/webhook?token='xxxxxxxxxxxxxxxxxxx-yyyyyyy'"
```

**Important:** Make sure to set `send_resolved: true` in your AlertManager webhook configuration to enable the resolved alert feature!

## Multiple AlertManager Configurations Example

You can configure multiple AlertManager instances or routes to different Mattermost channels:

```yaml
# In Mattermost Plugin Settings, configure:
# Config #0: Team=myteam, Channel=alerts-prod, Token=token-prod-xxx
# Config #1: Team=myteam, Channel=alerts-dev, Token=token-dev-yyy
# Config #2: Team=myteam, Channel=alerts-ops, Token=token-ops-zzz

# In AlertManager configuration:
receivers:
  - name: 'mattermost-prod'
    webhook_configs:
      - send_resolved: true
        url: 'https://mattermost.example.org/plugins/alertmanager/api/webhook?token=token-prod-xxx'
  
  - name: 'mattermost-dev'
    webhook_configs:
      - send_resolved: true
        url: 'https://mattermost.example.org/plugins/alertmanager/api/webhook?token=token-dev-yyy'
  
  - name: 'mattermost-ops'
    webhook_configs:
      - send_resolved: true
        url: 'https://mattermost.example.org/plugins/alertmanager/api/webhook?token=token-ops-zzz'

route:
  receiver: 'mattermost-prod'  # default
  routes:
    - receiver: 'mattermost-dev'
      matchers:
        - k8s_cluster_name = dev
    
    - receiver: 'mattermost-ops'
      matchers:
        - k8s_cluster_name = ops
```

After configuration, verify with `/alertmanager config` command to ensure all mappings are correct.

## Plugin in Action

![alertmanager-bot-1](assets/alertmanager-1.png)
![alertmanager-bot-2](assets/alertmanager-2.png)
![alertmanager-bot-3](assets/alertmanager-3.png)

## Development

To build the plugin:
```bash
make dist
```

The built plugin will be in `dist/` directory.

### Testing with curl

You can test the plugin by sending a test alert:

```bash
# Test firing alert
curl -X POST "https://YOUR-MATTERMOST-URL/plugins/alertmanager/api/webhook?token=YOUR-TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "receiver": "mattermost",
    "status": "firing",
    "alerts": [
      {
        "status": "firing",
        "labels": {
          "alertname": "TestAlert",
          "severity": "critical",
          "instance": "localhost:9090"
        },
        "annotations": {
          "summary": "This is a test alert",
          "description": "Testing the Mattermost AlertManager plugin"
        },
        "startsAt": "2024-11-22T10:00:00Z",
        "endsAt": "0001-01-01T00:00:00Z",
        "generatorURL": "http://localhost:9090/graph",
        "fingerprint": "test-alert-123"
      }
    ],
    "externalURL": "http://localhost:9093"
  }'

# Test resolved alert (same fingerprint)
curl -X POST "https://YOUR-MATTERMOST-URL/plugins/alertmanager/api/webhook?token=YOUR-TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "receiver": "mattermost",
    "status": "resolved",
    "alerts": [
      {
        "status": "resolved",
        "labels": {
          "alertname": "TestAlert",
          "severity": "critical",
          "instance": "localhost:9090"
        },
        "annotations": {
          "summary": "This is a test alert",
          "description": "Testing the Mattermost AlertManager plugin"
        },
        "startsAt": "2024-11-22T10:00:00Z",
        "endsAt": "2024-11-22T10:10:00Z",
        "generatorURL": "http://localhost:9090/graph",
        "fingerprint": "test-alert-123"
      }
    ],
    "externalURL": "http://localhost:9093"
  }'
```

Replace:
- `YOUR-MATTERMOST-URL` with your Mattermost server URL
- `YOUR-TOKEN` with the token from your plugin configuration

The first curl will create a firing alert (red), the second will resolve it (green) and add timing information to the thread.

## Troubleshooting

### Alerts going to wrong channels

1. Run `/alertmanager config` to verify channel mappings
2. Check Mattermost logs for `[WEBHOOK]` entries to see which config received the alert
3. Verify AlertManager routing configuration matches plugin token configuration
4. Run `/alertmanager reload` to refresh mappings
5. Check that channel names in plugin config exactly match Mattermost channel names

### Plugin not receiving webhooks

1. Check Mattermost logs for `[HTTP]` entries
2. Verify the webhook URL is accessible from AlertManager
3. Check token in URL matches plugin configuration
4. Ensure `@alertmanagerbot` is invited to target channels

### Resolved alerts not updating posts

1. Verify `send_resolved: true` is set in AlertManager webhook configuration
2. Check logs for `[WEBHOOK] Updated post for resolved alert` messages
3. Ensure the alert has the same fingerprint when firing and resolving
4. Check that the original post wasn't manually deleted

### Need more detailed logs

Set Mattermost log level to DEBUG in System Console to see detailed webhook processing logs including alert fingerprints and post IDs.
