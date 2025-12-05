package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DiagnosticData holds aggregated Kubernetes diagnostic information
type DiagnosticData struct {
	Namespace   string      `json:"namespace,omitempty"`
	Workloads   []string    `json:"workloads,omitempty"`
	Pods        []PodInfo   `json:"pods,omitempty"`
	Events      []EventInfo `json:"events,omitempty"`
	CollectedAt time.Time   `json:"collectedAt"`
	ContextName string      `json:"contextName,omitempty"`
}

// PodInfo contains relevant pod diagnostic information
type PodInfo struct {
	Name              string            `json:"name"`
	Phase             string            `json:"phase"`
	Ready             string            `json:"ready"`
	Restarts          int32             `json:"restarts"`
	Age               time.Duration     `json:"age"`
	Message           string            `json:"message,omitempty"`
	ContainerStatuses []ContainerStatus `json:"containerStatuses,omitempty"`
	NodeName          string            `json:"nodeName,omitempty"`
	Conditions        []PodCondition    `json:"conditions,omitempty"`
}

// ContainerStatus holds container-level diagnostic info
type ContainerStatus struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restartCount"`
	State        string `json:"state,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Message      string `json:"message,omitempty"`
	Image        string `json:"image,omitempty"`
}

// PodCondition represents a pod condition
type PodCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// EventInfo contains Kubernetes event information
type EventInfo struct {
	Type           string    `json:"type"`
	Reason         string    `json:"reason"`
	Message        string    `json:"message"`
	InvolvedObject string    `json:"involvedObject,omitempty"`
	FirstTimestamp time.Time `json:"firstTimestamp,omitempty"`
	LastTimestamp  time.Time `json:"lastTimestamp"`
	Count          int32     `json:"count"`
}

// Aggregator collects diagnostic data from Kubernetes
type Aggregator struct {
	client *Client
}

// NewAggregator creates a new diagnostic aggregator
func NewAggregator(client *Client) *Aggregator {
	return &Aggregator{
		client: client,
	}
}

// CollectDiagnostics gathers diagnostic data for a namespace and optional workloads
func (a *Aggregator) CollectDiagnostics(ctx context.Context, namespace string, workloads []string) (*DiagnosticData, error) {
	data := &DiagnosticData{
		Namespace:   namespace,
		Workloads:   workloads,
		CollectedAt: time.Now(),
	}

	// Get current context name
	contextName, err := GetCurrentContext("")
	if err == nil {
		data.ContextName = contextName
	}

	// Collect pods
	pods, err := a.collectPods(ctx, namespace, workloads)
	if err != nil {
		return nil, fmt.Errorf("failed to collect pods: %w", err)
	}
	data.Pods = pods

	// Collect events
	events, err := a.collectEvents(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to collect events: %w", err)
	}
	data.Events = events

	return data, nil
}

func (a *Aggregator) collectPods(ctx context.Context, namespace string, workloads []string) ([]PodInfo, error) {
	listOpts := metav1.ListOptions{}

	// If specific workloads are requested, filter by labels or names
	if len(workloads) > 0 {
		// For simplicity, we'll collect all and filter in memory
		// In production, you'd want to use label selectors
	}

	podList, err := a.client.Clientset().CoreV1().Pods(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	var pods []PodInfo
	for _, pod := range podList.Items {
		// Filter by workload if specified
		if len(workloads) > 0 && !a.matchesWorkload(&pod, workloads) {
			continue
		}

		podInfo := a.extractPodInfo(&pod)
		pods = append(pods, podInfo)
	}

	return pods, nil
}

func (a *Aggregator) extractPodInfo(pod *corev1.Pod) PodInfo {
	info := PodInfo{
		Name:     pod.Name,
		Phase:    string(pod.Status.Phase),
		NodeName: pod.Spec.NodeName,
		Age:      time.Since(pod.CreationTimestamp.Time),
	}

	// Calculate ready status
	readyCount := 0
	totalCount := len(pod.Status.ContainerStatuses)
	var totalRestarts int32

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			readyCount++
		}
		totalRestarts += cs.RestartCount

		containerStatus := ContainerStatus{
			Name:         cs.Name,
			Ready:        cs.Ready,
			RestartCount: cs.RestartCount,
			Image:        cs.Image,
		}

		// Extract state information
		if cs.State.Running != nil {
			containerStatus.State = "Running"
		} else if cs.State.Waiting != nil {
			containerStatus.State = "Waiting"
			containerStatus.Reason = cs.State.Waiting.Reason
			containerStatus.Message = cs.State.Waiting.Message
		} else if cs.State.Terminated != nil {
			containerStatus.State = "Terminated"
			containerStatus.Reason = cs.State.Terminated.Reason
			containerStatus.Message = cs.State.Terminated.Message
		}

		info.ContainerStatuses = append(info.ContainerStatuses, containerStatus)
	}

	info.Ready = fmt.Sprintf("%d/%d", readyCount, totalCount)
	info.Restarts = totalRestarts

	// Extract pod conditions
	for _, cond := range pod.Status.Conditions {
		if cond.Status == corev1.ConditionFalse || cond.Reason != "" {
			info.Conditions = append(info.Conditions, PodCondition{
				Type:    string(cond.Type),
				Status:  string(cond.Status),
				Reason:  cond.Reason,
				Message: cond.Message,
			})
		}
	}

	return info
}

func (a *Aggregator) collectEvents(ctx context.Context, namespace string) ([]EventInfo, error) {
	listOpts := metav1.ListOptions{
		// Get events from the last hour
		FieldSelector: fmt.Sprintf("involvedObject.namespace=%s", namespace),
	}

	eventList, err := a.client.Clientset().CoreV1().Events(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	var events []EventInfo
	cutoff := time.Now().Add(-1 * time.Hour)

	for _, event := range eventList.Items {
		// Filter recent events
		if event.LastTimestamp.Time.Before(cutoff) {
			continue
		}

		// Focus on warning and error events
		if event.Type != "Warning" && event.Type != "Error" {
			continue
		}

		events = append(events, EventInfo{
			Type:           event.Type,
			Reason:         event.Reason,
			Message:        event.Message,
			InvolvedObject: fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
			FirstTimestamp: event.FirstTimestamp.Time,
			LastTimestamp:  event.LastTimestamp.Time,
			Count:          event.Count,
		})
	}

	return events, nil
}

func (a *Aggregator) matchesWorkload(pod *corev1.Pod, workloads []string) bool {
	// Check if pod name starts with any of the workload names
	// This is a simple heuristic; in production, use owner references
	for _, workload := range workloads {
		if len(pod.Name) >= len(workload) && pod.Name[:len(workload)] == workload {
			return true
		}
	}
	return false
}
