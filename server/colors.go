package main

const (
	// Default state colors
	colorFiring       = "#FF0000" // red
	colorAcknowledged = "#9013FE" // purple
	colorResolved     = "#008000" // green
	colorExpired      = "#F0F8FF" // aliceBlue

	// Default severity colors
	colorCritical = "#FF0000" // red
	colorError    = "#F5A623" // golden-orange
	colorWarning  = "#F8E71C" // bright yellow
	colorInfo     = "#0080FF" // blue
	colorDebug    = "#87CEEB" // light blue

	// Alert states
	stateFiring   = "firing"
	stateAcked    = "acked"
	stateResolved = "resolved"
)

// getAlertColor returns the appropriate color for an alert
// Priority: severity color > state color > default color
func getAlertColor(alertConfig alertConfig, severity, state string) string {
	// Priority 1: Check severity color mapping
	if severity != "" && alertConfig.SeverityColors != nil {
		if color, ok := alertConfig.SeverityColors[severity]; ok && color != "" {
			return color
		}
	}

	// Priority 2: Check state color mapping
	if state != "" && alertConfig.StateColors != nil {
		if color, ok := alertConfig.StateColors[state]; ok && color != "" {
			return color
		}
	}

	// Priority 3: Default colors based on state
	switch state {
	case stateAcked:
		return colorAcknowledged
	case stateResolved:
		return colorResolved
	case stateFiring:
		// If we have severity but no custom color, use severity defaults
		switch severity {
		case "critical":
			return colorCritical
		case "error":
			return colorError
		case "warning":
			return colorWarning
		case "info":
			return colorInfo
		case "debug":
			return colorDebug
		default:
			return colorFiring
		}
	default:
		return colorFiring
	}
}
