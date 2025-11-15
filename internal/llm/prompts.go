package llm

import (
	"fmt"
	"kubehelp/internal/k8s"
	"strings"
	"time"
)

// BuildDiagnosticPrompt creates a structured prompt from diagnostic data
func BuildDiagnosticPrompt(data *k8s.DiagnosticData) string {
	var sb strings.Builder

	sb.WriteString("# Kubernetes Diagnostic Report\n\n")
	sb.WriteString(fmt.Sprintf("**Cluster Context:** %s\n", data.ContextName))
	sb.WriteString(fmt.Sprintf("**Namespace:** %s\n", data.Namespace))
	sb.WriteString(fmt.Sprintf("**Collection Time:** %s\n\n", data.CollectedAt.Format(time.RFC3339)))

	if len(data.Workloads) > 0 {
		sb.WriteString(fmt.Sprintf("**Focused Workloads:** %s\n\n", strings.Join(data.Workloads, ", ")))
	}

	// Pod Status Summary
	sb.WriteString("## Pod Status Summary\n\n")
	if len(data.Pods) == 0 {
		sb.WriteString("No pods found in this namespace.\n\n")
	} else {
		sb.WriteString("| Pod Name | Phase | Ready | Restarts | Age | Node |\n")
		sb.WriteString("|----------|-------|-------|----------|-----|------|\n")
		for _, pod := range data.Pods {
			age := formatDuration(pod.Age)
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %s | %s |\n",
				pod.Name, pod.Phase, pod.Ready, pod.Restarts, age, pod.NodeName))
		}
		sb.WriteString("\n")
	}

	// Container Details
	sb.WriteString("## Container Details\n\n")
	for _, pod := range data.Pods {
		if len(pod.ContainerStatuses) == 0 {
			continue
		}

		// Only include pods with issues
		hasIssues := false
		for _, cs := range pod.ContainerStatuses {
			if !cs.Ready || cs.State != "Running" || cs.RestartCount > 0 {
				hasIssues = true
				break
			}
		}

		if !hasIssues && len(pod.Conditions) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("### Pod: %s\n\n", pod.Name))
		for _, cs := range pod.ContainerStatuses {
			sb.WriteString(fmt.Sprintf("**Container:** %s\n", cs.Name))
			sb.WriteString(fmt.Sprintf("- Image: %s\n", cs.Image))
			sb.WriteString(fmt.Sprintf("- State: %s\n", cs.State))
			sb.WriteString(fmt.Sprintf("- Ready: %v\n", cs.Ready))
			sb.WriteString(fmt.Sprintf("- Restart Count: %d\n", cs.RestartCount))
			if cs.Reason != "" {
				sb.WriteString(fmt.Sprintf("- Reason: %s\n", cs.Reason))
			}
			if cs.Message != "" {
				sb.WriteString(fmt.Sprintf("- Message: %s\n", cs.Message))
			}
			sb.WriteString("\n")
		}

		// Add pod conditions if any
		if len(pod.Conditions) > 0 {
			sb.WriteString("**Pod Conditions:**\n")
			for _, cond := range pod.Conditions {
				sb.WriteString(fmt.Sprintf("- %s: %s", cond.Type, cond.Status))
				if cond.Reason != "" {
					sb.WriteString(fmt.Sprintf(" (Reason: %s)", cond.Reason))
				}
				if cond.Message != "" {
					sb.WriteString(fmt.Sprintf(" - %s", cond.Message))
				}
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}
	}

	// Recent Events
	sb.WriteString("## Recent Events (Last Hour)\n\n")
	if len(data.Events) == 0 {
		sb.WriteString("No warning or error events in the last hour.\n\n")
	} else {
		sb.WriteString("| Type | Reason | Object | Count | Message |\n")
		sb.WriteString("|------|--------|--------|-------|----------|\n")
		for _, event := range data.Events {
			// Truncate long messages
			msg := event.Message
			if len(msg) > 80 {
				msg = msg[:77] + "..."
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %s |\n",
				event.Type, event.Reason, event.InvolvedObject, event.Count, msg))
		}
		sb.WriteString("\n")
	}

	// Request analysis
	sb.WriteString("## Analysis Request\n\n")
	sb.WriteString("Please analyze the above diagnostic data and provide:\n\n")
	sb.WriteString("1. **Summary of Issues**: Identify the main problems affecting this namespace\n")
	sb.WriteString("2. **Root Cause Analysis**: Explain the likely root causes\n")
	sb.WriteString("3. **Remediation Steps**: Provide specific, actionable steps to resolve the issues\n")
	sb.WriteString("4. **kubectl Commands**: Include relevant kubectl commands that might help\n")
	sb.WriteString("5. **Prevention**: Suggest how to prevent similar issues in the future\n\n")
	sb.WriteString("Focus on the most critical issues first.\n")

	return sb.String()
}

// formatDuration converts a duration to a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%dd", days)
}
