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
// Priority for ACKED/RESOLVED: state color always wins
// Priority for FIRING: severity color > state color > default
func getAlertColor(alertConfig alertConfig, severity, state string) string {
	// For ACKED and RESOLVED states, state color takes priority over severity
	if state == stateAcked || state == stateResolved {
		// Priority 1: Custom state color
		if alertConfig.StateColors != nil {
			if color, ok := alertConfig.StateColors[state]; ok && color != "" {
				return color
			}
		}

		// Priority 2: Default state color
		if state == stateAcked {
			return colorAcknowledged
		}
		if state == stateResolved {
			return colorResolved
		}
	}

	// For FIRING state, severity color takes priority
	if state == stateFiring {
		// Priority 1: Custom severity color
		if severity != "" && alertConfig.SeverityColors != nil {
			if color, ok := alertConfig.SeverityColors[severity]; ok && color != "" {
				return color
			}
		}

		// Priority 2: Custom state color
		if alertConfig.StateColors != nil {
			if color, ok := alertConfig.StateColors[stateFiring]; ok && color != "" {
				return color
			}
		}

		// Priority 3: Default severity color
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
		}

		// Priority 4: Default firing color
		return colorFiring
	}

	// Fallback for unknown states
	return colorFiring
}
