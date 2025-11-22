package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	alerttemplate "github.com/prometheus/alertmanager/template"
)

// renderAlertTemplate renders a custom template for an alert
func renderAlertTemplate(tmpl string, alert alerttemplate.Alert) (string, error) {
	if tmpl == "" {
		return "", nil
	}

	funcMap := template.FuncMap{
		"toUpper": func(s string) string {
			return strings.ToUpper(s)
		},
		"formatTime": func(t time.Time) string {
			return t.Format(time.RFC1123)
		},
	}

	t, err := template.New("alert").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, alert); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// Default templates
const (
	DefaultFiringTemplate = `🔥 **{{ .Labels.alertname }}**

{{ if .Annotations.summary }}**Summary:** {{ .Annotations.summary }}{{ end }}
{{ if .Annotations.description }}**Description:** {{ .Annotations.description }}{{ end }}

**Severity:** {{ .Labels.severity }}
**Started at:** {{ formatTime .StartsAt }}`

	DefaultResolvedTemplate = `✅ **{{ .Labels.alertname }} - RESOLVED**

{{ if .Annotations.summary }}**Summary:** {{ .Annotations.summary }}{{ end }}

**Started at:** {{ formatTime .StartsAt }}
**Resolved at:** {{ formatTime .EndsAt }}
**Duration:** {{ .EndsAt.Sub .StartsAt }}`
)
