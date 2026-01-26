// Package opensearch provides alert management repository
package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

// Alert represents an alert document
type Alert struct {
	ID         string                 `json:"id,omitempty"`
	Hostname   string                 `json:"hostname"`
	AlertType  string                 `json:"alert_type"`
	Severity   string                 `json:"severity"` // critical, high, medium, low
	Title      string                 `json:"title"`
	Message    string                 `json:"description"`
	Timestamp  int64                  `json:"timestamp"`
	ResolvedAt *int64                 `json:"resolved_at,omitempty"`
	Status     string                 `json:"status"` // active, resolved
	Value      float64                `json:"value,omitempty"`
	Threshold  float64                `json:"threshold,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AlertsRepository manages alert operations
type AlertsRepository struct {
	client *Client
}

// NewAlertsRepository creates a new alerts repository
func NewAlertsRepository(client *Client) (*AlertsRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure index exists
	if err := client.CreateIndex(ctx, AlertsIndex, AlertsIndexMapping); err != nil {
		return nil, fmt.Errorf("failed to create alerts index: %w", err)
	}

	return &AlertsRepository{
		client: client,
	}, nil
}

// CreateAlert creates a new alert
func (r *AlertsRepository) CreateAlert(ctx context.Context, alert *Alert) (string, error) {
	if alert == nil {
		return "", fmt.Errorf("alert cannot be nil")
	}

	if alert.ID == "" {
		alert.ID = fmt.Sprintf("%s-%d", alert.Hostname, time.Now().UnixMilli())
	}
	alert.Timestamp = time.Now().UnixMilli()
	alert.Status = "active"

	body, err := json.Marshal(alert)
	if err != nil {
		return "", fmt.Errorf("failed to marshal alert: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      AlertsIndex,
		DocumentID: alert.ID,
		Body:       bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return "", fmt.Errorf("failed to create alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenSearch error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return alert.ID, nil
}

// GetAlert retrieves an alert by ID
func (r *AlertsRepository) GetAlert(ctx context.Context, alertID string) (*Alert, error) {
	req := opensearchapi.GetRequest{
		Index:      AlertsIndex,
		DocumentID: alertID,
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("alert not found: %s", alertID)
	}

	var result struct {
		Source Alert `json:"_source"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Source, nil
}

// ResolveAlert marks an alert as resolved
func (r *AlertsRepository) ResolveAlert(ctx context.Context, alertID string) error {
	now := time.Now().UnixMilli()
	update := map[string]interface{}{
		"doc": map[string]interface{}{
			"status":      "resolved",
			"resolved_at": now,
		},
	}

	body, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal update: %w", err)
	}

	req := opensearchapi.UpdateRequest{
		Index:      AlertsIndex,
		DocumentID: alertID,
		Body:       bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to resolve alert: status %d", resp.StatusCode)
	}

	return nil
}

// ListAlerts retrieves alerts with optional filters
func (r *AlertsRepository) ListAlerts(ctx context.Context, hostname string, severity string, status string, limit int) ([]*Alert, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{},
			},
		},
		"sort": []map[string]interface{}{
			{
				"timestamp": map[string]interface{}{
					"order": "desc",
				},
			},
		},
		"size": limit,
	}

	must := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})

	if hostname != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"hostname": hostname,
			},
		})
	}

	if severity != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"severity": severity,
			},
		})
	}

	if status != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"status": status,
			},
		})
	}

	query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = must

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{AlertsIndex},
		Body:  bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to search alerts: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Hits struct {
			Hits []struct {
				Source Alert `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var alerts []*Alert
	for i := range result.Hits.Hits {
		alerts = append(alerts, &result.Hits.Hits[i].Source)
	}

	return alerts, nil
}

// GetAlertStats returns statistics about alerts
func (r *AlertsRepository) GetAlertStats(ctx context.Context) (map[string]interface{}, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"aggs": map[string]interface{}{
			"by_severity": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "severity",
				},
			},
			"by_status": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "status",
				},
			},
			"by_hostname": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "hostname",
					"size":  1000,
				},
			},
		},
		"size": 0,
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{AlertsIndex},
		Body:  bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert stats: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
