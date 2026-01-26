// Package opensearch provides event logging repository
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

// Event represents an event document
type Event struct {
	ID          string                 `json:"id,omitempty"`
	Hostname    string                 `json:"hostname"`
	EventType   string                 `json:"event_type"` // security, performance, availability, etc
	EventName   string                 `json:"event_name"`
	Timestamp   int64                  `json:"timestamp"`
	Message     string                 `json:"message"`
	Source      string                 `json:"source"`
	User        string                 `json:"user,omitempty"`
	ProcessID   int32                  `json:"process_id,omitempty"`
	ProcessName string                 `json:"process_name,omitempty"`
	Level       string                 `json:"level"` // info, warning, error, critical
	Details     map[string]interface{} `json:"details,omitempty"`
}

// EventsRepository manages event operations
type EventsRepository struct {
	client *Client
}

// NewEventsRepository creates a new events repository
func NewEventsRepository(client *Client) (*EventsRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure index exists
	if err := client.CreateIndex(ctx, EventsIndex, EventsIndexMapping); err != nil {
		return nil, fmt.Errorf("failed to create events index: %w", err)
	}

	return &EventsRepository{
		client: client,
	}, nil
}

// LogEvent creates a new event record
func (r *EventsRepository) LogEvent(ctx context.Context, event *Event) (string, error) {
	if event == nil {
		return "", fmt.Errorf("event cannot be nil")
	}

	if event.ID == "" {
		event.ID = fmt.Sprintf("%s-%d", event.Hostname, time.Now().UnixNano())
	}
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}

	body, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("failed to marshal event: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      EventsIndex,
		DocumentID: event.ID,
		Body:       bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return "", fmt.Errorf("failed to log event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenSearch error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return event.ID, nil
}

// GetEvent retrieves an event by ID
func (r *EventsRepository) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	req := opensearchapi.GetRequest{
		Index:      EventsIndex,
		DocumentID: eventID,
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}

	var result struct {
		Source Event `json:"_source"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Source, nil
}

// ListEvents retrieves events with optional filters
func (r *EventsRepository) ListEvents(ctx context.Context, hostname string, eventType string, level string, limit int, offset int) ([]*Event, error) {
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
		"from": offset,
	}

	must := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})

	if hostname != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"hostname": hostname,
			},
		})
	}

	if eventType != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"event_type": eventType,
			},
		})
	}

	if level != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"level": level,
			},
		})
	}

	query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = must

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{EventsIndex},
		Body:  bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Hits struct {
			Hits []struct {
				Source Event `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var events []*Event
	for i := range result.Hits.Hits {
		events = append(events, &result.Hits.Hits[i].Source)
	}

	return events, nil
}

// SearchEvents performs full-text search on events
func (r *EventsRepository) SearchEvents(ctx context.Context, query string, hostname string, limit int) ([]*Event, error) {
	searchQuery := map[string]interface{}{
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

	must := searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})

	if query != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"message", "event_name", "source"},
			},
		})
	}

	if hostname != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"hostname": hostname,
			},
		})
	}

	searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = must

	body, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{EventsIndex},
		Body:  bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Hits struct {
			Hits []struct {
				Source Event `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var events []*Event
	for i := range result.Hits.Hits {
		events = append(events, &result.Hits.Hits[i].Source)
	}

	return events, nil
}

// GetEventStats returns statistics about events
func (r *EventsRepository) GetEventStats(ctx context.Context) (map[string]interface{}, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"aggs": map[string]interface{}{
			"by_type": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "event_type",
				},
			},
			"by_level": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "level",
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
		Index: []string{EventsIndex},
		Body:  bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to get event stats: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
