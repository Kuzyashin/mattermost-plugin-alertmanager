# AlertManager Plugin

This plugin is the [AlertManager](https://github.com/prometheus/alertmanager) bot for Mattermost.

## Fork Information

This is a fork of [cpanato/mattermost-plugin-alertmanager](https://github.com/cpanato/mattermost-plugin-alertmanager) with significant improvements and bug fixes:

### Key Improvements in This Fork

#### 🎨 Visual & UX Enhancements
- ✅ **Visual Color Picker UI** - Point-and-click color customization in System Console
- ✅ **ACK/UNACK functionality** - Interactive buttons with real-time updates
- ✅ **Visual feedback** - Posts visually reflect alert state (🔥 FIRING 🔥 ↔ 👁️ ACKNOWLEDGED 👁️)
- ✅ **Flexible color mappings** - Customize colors for alert states (firing/acked/resolved) and severity levels (critical/error/warning/info/debug)
- ✅ **Priority-based coloring** - Severity colors override state colors for maximum flexibility

#### 🔧 Functional Enhancements
- ✅ **Real Silence API integration** - Creates actual silences in AlertManager (not just visual)
- ✅ **Thread tracking** - All actions (ACK/UNACK/Silence) create detailed thread replies
- ✅ **Persistent state** - Alert acknowledgment state stored in KV Store
- ✅ **Multiple duration options** - Silence buttons for 1h/4h/12h/24h

#### 🐛 Bug Fixes & Code Quality
- ✅ **Fixed all golangci-lint errors** - errcheck, goconst, revive, unused, gocritic, govet
- ✅ **Memory optimizations** - Struct field alignment optimizations
- ✅ **Fixed syntax errors** - Corrected string literals and comments
- ✅ **Improved error handling** - All JSON encode/decode operations properly checked
- ✅ **Code consistency** - String constants extracted, consistent naming conventions

### Original Inspiration

Originally forked and inspired by [@metalmatze](https://github.com/metalmatze/)'s [alertmanager-bot](https://github.com/metalmatze/alertmanager-bot) for Telegram

## Features

### Core Functionality
- ✅ Receive alerts via webhook from Prometheus AlertManager
- ✅ Smart alert lifecycle management - firing alerts create posts, resolved alerts update them
- ✅ Multiple AlertManager configurations support with independent routing
- ✅ Automatic channel creation and bot user management

### Interactive Actions 🎯
- ✅ **Fully functional ACK/UNACK buttons** - Real-time post updates with button state changes
- ✅ **Silence buttons** (1h/4h/12h/24h) - Real API integration with AlertManager to create silences
- ✅ **Dynamic visual updates** - Colors and status titles change instantly on action
  - 🔥 FIRING 🔥 (red) ↔ 👁️ ACKNOWLEDGED 👁️ (yellow/orange)
- ✅ **Thread replies** - Every action creates a thread post with user and timestamp
- ✅ **Persistent state** - ACK state stored in KV Store and survives restarts

### Alert Presentation
- ✅ Color-coded posts:
  - 🔴 Red = Firing (not acknowledged)
  - 🟡 Yellow/Orange = Acknowledged
  - 🟢 Green = Resolved
- ✅ Thread replies with resolution timing information
- ✅ Severity-based mentions (@team notifications for critical alerts)
- ✅ Custom Go templates for firing and resolved alerts

### Configuration & Management
- ✅ `/alertmanager reload` - Reload channel mappings without plugin restart
- ✅ `/alertmanager config` - Display current configuration and channel mappings
- ✅ `/alertmanager alerts` - List existing alerts
- ✅ `/alertmanager silences` - List active silences
- ✅ `/alertmanager expire_silence` - Expire a silence
- ✅ Auto-reload channel mappings on configuration save
- ✅ Enhanced logging for webhook processing and troubleshooting

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
- **🔕 Silence 1h / 4h / 12h / 24h** - Create a real silence in AlertManager for the specified duration
  - Uses AlertManager API (`POST /api/v2/silences`)
  - Automatically creates matchers based on alert labels
  - Returns silence ID for tracking
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
   - **ACK**: Button changes from "👁️ ACK" → "🔄 UNACK", color changes based on severity/state mapping, status changes from "🔥 FIRING 🔥" → "👁️ ACKNOWLEDGED 👁️"
   - **UNACK**: Button changes from "🔄 UNACK" → "👁️ ACK", color returns based on severity/state mapping, status returns to "🔥 FIRING 🔥"

### Visual Indicators and Color Mapping 🎨

The plugin supports **flexible color customization** based on alert state and severity with a priority system.

#### Default Colors

**Alert States:**
- **🔥 FIRING**: `#FF0000` (Red)
- **👁️ ACKNOWLEDGED**: `#9013FE` (Purple)
- **✅ RESOLVED**: `#008000` (Green)

**Severity Levels** (applied to FIRING alerts by default):
- **critical**: `#FF0000` (Red)
- **error**: `#F5A623` (Golden-Orange)
- **warning**: `#F8E71C` (Bright Yellow)
- **info**: `#0080FF` (Blue)
- **debug**: `#87CEEB` (Light Blue)

#### Custom Color Configuration

**✨ Visual Color Picker UI Available!**

The plugin includes a **visual color picker** in the System Console for easy customization:

1. Navigate to **System Console** → **Plugins** → **AlertManager**
2. For each Alert Config, you'll see:
   - **Alert State Colors** section with color swatches for Firing/Acknowledged/Resolved
   - **Severity Level Colors** section with color swatches for Critical/Error/Warning/Info/Debug
3. Click any color swatch to open a visual color picker
4. Select your desired color and it updates in real-time
5. Click "Reset" button to restore default colors

**Programmatic Configuration (optional):**

You can also configure colors via JSON if needed:

```json
{
  "AlertConfigs": {
    "my-config": {
      "StateColors": {
        "firing": "#FF0000",
        "acked": "#FFAA00",
        "resolved": "#00FF00"
      },
      "SeverityColors": {
        "critical": "#8B0000",
        "warning": "#FFD700",
        "info": "#4169E1",
        "debug": "#B0E0E6"
      }
    }
  }
}
```

#### Color Priority System

Colors are applied in this priority order:
1. **Severity color** (highest priority) - if alert has `severity` label and custom mapping exists
2. **State color** - if custom state mapping exists
3. **Default severity color** - built-in severity colors for firing alerts
4. **Default state color** (lowest priority) - built-in state colors

**Examples:**
- Critical firing alert → Uses severity color (red)
- Info alert that's been ACKed → Uses ACK state color (yellow/orange)
- Warning alert with custom severity color → Uses custom warning color
- Resolved critical alert → Uses resolved state color (green)

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
Silence ID: `a1b2c3d4-e5f6-7890-abcd-ef1234567890`
```

The Silence ID can be used to manually expire or modify the silence in AlertManager if needed.

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

## AlertManager API Integration

### Silence Creation

The plugin integrates directly with AlertManager API to create real silences:

**API Endpoint**: `POST /api/v2/silences`

**How it works**:
1. User clicks a Silence button (1h/4h/12h/24h) on an alert post
2. Plugin extracts alert labels from the button context
3. Creates silence matchers based on all alert labels
4. Sends POST request to AlertManager with:
   - **Matchers**: Exact match on all alert labels (e.g., `alertname=HighCPU`, `instance=server1`)
   - **StartsAt**: Current time
   - **EndsAt**: Current time + duration
   - **CreatedBy**: Mattermost username
   - **Comment**: "Silenced from Mattermost by {username}"
5. Returns Silence ID in thread reply for tracking

**Benefits**:
- ✅ Real silences in AlertManager - alerts stop firing
- ✅ Automatic matcher creation from alert labels
- ✅ Audit trail with creator username
- ✅ Silence ID for manual expiration if needed
- ✅ Works with any AlertManager v2 API compatible instance

**Example API Request**:
```json
{
  "matchers": [
    {"name": "alertname", "value": "HighCPU", "isRegex": false, "isEqual": true},
    {"name": "instance", "value": "server1", "isRegex": false, "isEqual": true},
    {"name": "severity", "value": "critical", "isRegex": false, "isEqual": true}
  ],
  "startsAt": "2024-11-23T10:00:00Z",
  "endsAt": "2024-11-23-11:00:00Z",
  "createdBy": "johndoe",
  "comment": "Silenced from Mattermost by johndoe"
}
```

**Error Handling**:
- Connection errors to AlertManager are logged and returned to user
- Invalid durations are rejected
- Missing alert labels result in error

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

## Technical Details & Improvements

### Color Customization Architecture

The plugin implements a sophisticated **priority-based color system**:

**Priority Order (highest to lowest):**
1. **Custom Severity Color** - User-defined color for specific severity level
2. **Custom State Color** - User-defined color for alert state
3. **Default Severity Color** - Built-in color based on severity label
4. **Default State Color** - Built-in color based on alert state

**Implementation:**
- `getAlertColor(alertConfig, severity, state)` function in `/server/colors.go`
- Color resolution happens at render time for webhooks and button actions
- Frontend: React component with `react-color` SketchPicker integration
- Backend: `StateColorMap` and `SeverityColorMap` types with JSON unmarshaling

**Default Color Palette:**
```go
// State colors
colorFiring       = "#FF0000" // red
colorAcknowledged = "#9013FE" // purple
colorResolved     = "#008000" // green

// Severity colors
colorCritical = "#FF0000" // red
colorError    = "#F5A623" // golden-orange
colorWarning  = "#F8E71C" // bright yellow
colorInfo     = "#0080FF" // blue
colorDebug    = "#87CEEB" // light blue
```

### Code Quality Improvements

**Linter Fixes:**
- **errcheck**: Added error handling for all `json.Encoder.Encode()` calls
- **goconst**: Extracted repeated strings to constants (`actionSilence`, `actionAck`, etc.)
- **revive**: Removed unused parameters (`_` prefix for intentionally unused)
- **unused**: Deleted unused `getAlertAck()` function
- **gocritic/unlambda**: Simplified lambda expressions (e.g., `strings.ToUpper` instead of wrapper)
- **govet/fieldalignment**: Optimized struct memory layout (reduced from 128 to 104 pointer bytes)
- **govet/shadow**: Renamed shadowed variables (`alertConfig` → `alertCfg`)

**Memory Optimizations:**
```go
// Before (128 pointer bytes)
type alertConfig struct {
    ID               string
    Token            string
    // ... strings first
    SeverityMentions SeverityMentionsMap
}

// After (104 pointer bytes) - 18.75% reduction
type alertConfig struct {
    SeverityMentions SeverityMentionsMap // maps first
    StateColors      StateColorMap
    SeverityColors   SeverityColorMap
    ID               string              // strings next
    Token            string
    // ... other fields
}
```

**Error Handling Pattern:**
```go
// Before
json.NewEncoder(w).Encode(response)

// After
if err := json.NewEncoder(w).Encode(response); err != nil {
    p.API.LogError("[ACTION] Failed to encode response", "error", err.Error())
}
```

### Frontend Architecture

**Components:**
- **ColorMapEditor.jsx** - Reusable color picker component with presets
- **AMAttribute.jsx** - Alert config editor with integrated color pickers
- **CustomAttributeSettings.jsx** - Multi-config management

**Features:**
- Visual color swatches showing current color
- Click-to-open SketchPicker with HEX input
- Reset buttons to restore defaults
- Real-time preview and save
- Responsive CSS Grid layout

**Dependencies:**
```json
{
  "react-color": "^2.19.3"
}
```

### Silence API Integration

**Flow:**
1. User clicks Silence button (1h/4h/12h/24h)
2. Button context includes alert labels + severity
3. Plugin creates matchers from all labels (exact match, non-regex)
4. POST to AlertManager `/api/v2/silences`
5. Silence ID returned and displayed in thread

**Matcher Generation:**
```go
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
```

**API Client:**
- 10 second timeout
- Full error logging with AlertManager URL
- Supports all AlertManager v2 compatible instances

### ACK/UNACK State Management

**Storage:**
- KV Store key: `alert_ack_{fingerprint}`
- Value: JSON with `{userID, username, timestamp}`
- Persists across plugin restarts

**Button Updates:**
- Post attachment modification (color, title, actions)
- Thread reply with audit trail
- Real-time UI update via `PostActionIntegrationResponse`

### Alert Lifecycle Tracking

**Fingerprint Mapping:**
- KV Store key: `alert_post_{fingerprint}`
- Value: Post ID
- Enables resolved alert updates

**Resolved Alert Flow:**
1. Find original post via fingerprint
2. Update post color → green
3. Remove action buttons
4. Create thread with timing (`startsAt`, `endsAt`, `duration`)

### Build & Dependencies

**Backend (Go):**
- Go 1.24+
- Mattermost Plugin API 5.37+
- Prometheus AlertManager types

**Frontend (React):**
- React 18.2.0
- react-color 2.19.3
- Webpack 5 build system

**Build Output:**
```
dist/alertmanager-0.5.3.tar.gz (76 MB)
- server/dist/ (multiple architectures)
  - plugin-linux-amd64
  - plugin-linux-arm64
  - plugin-darwin-amd64
  - plugin-darwin-arm64
  - plugin-windows-amd64.exe
- webapp/dist/
  - main.js (1.46 MB with react-color)
```

### Backward Compatibility

All changes are **backward compatible**:
- Existing configs work without color customization
- Default colors match previous behavior
- Empty color maps use defaults
- JSON string config format still supported
