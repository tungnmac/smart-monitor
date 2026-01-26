// Package opensearch provides OpenSearch-based repositories
package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"smart-monitor/backend/internal/domain/entity"

	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

// statsDoc is the OpenSearch document representation of Stats
type statsDoc struct {
	Hostname     string            `json:"hostname"`
	AgentID      string            `json:"agent_id"`
	IPAddress    string            `json:"ip_address"`
	CPU          float64           `json:"cpu"`
	RAM          float64           `json:"ram"`
	Disk         float64           `json:"disk"`
	Timestamp    int64             `json:"timestamp"`
	LastReceived int64             `json:"last_received"`
	Metadata     map[string]string `json:"metadata"`
}

func (d *statsDoc) toEntity() *entity.Stats {
	return &entity.Stats{
		Hostname:     d.Hostname,
		AgentID:      d.AgentID,
		IPAddress:    d.IPAddress,
		CPU:          d.CPU,
		RAM:          d.RAM,
		Disk:         d.Disk,
		Timestamp:    time.UnixMilli(d.Timestamp),
		LastReceived: time.UnixMilli(d.LastReceived),
		Metadata:     d.Metadata,
	}
}

// OpenSearchStatsRepository implements StatsRepository using OpenSearch
type OpenSearchStatsRepository struct {
	client *Client
}

// NewOpenSearchStatsRepository creates a new OpenSearch stats repository
func NewOpenSearchStatsRepository(client *Client) (*OpenSearchStatsRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure index exists
	if err := client.CreateIndex(ctx, StatsIndex, StatsIndexMapping); err != nil {
		return nil, fmt.Errorf("failed to create stats index: %w", err)
	}

	return &OpenSearchStatsRepository{
		client: client,
	}, nil
}

// Save stores stats in OpenSearch
func (r *OpenSearchStatsRepository) Save(ctx context.Context, stats *entity.Stats) error {
	if stats == nil {
		return fmt.Errorf("stats cannot be nil")
	}

	doc := map[string]interface{}{
		"hostname":      stats.Hostname,
		"agent_id":      stats.AgentID,
		"ip_address":    stats.IPAddress,
		"cpu":           stats.CPU,
		"ram":           stats.RAM,
		"disk":          stats.Disk,
		"timestamp":     stats.Timestamp.UnixMilli(),
		"last_received": stats.LastReceived.UnixMilli(),
		"metadata":      stats.Metadata,
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      StatsIndex,
		DocumentID: fmt.Sprintf("%s-%d", stats.Hostname, stats.Timestamp.UnixMilli()),
		Body:       bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return fmt.Errorf("failed to index stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OpenSearch error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Get retrieves the latest stats for a hostname
func (r *OpenSearchStatsRepository) Get(ctx context.Context, hostname string) (*entity.Stats, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"hostname": hostname,
						},
					},
				},
			},
		},
		"sort": []map[string]interface{}{
			{
				"timestamp": map[string]interface{}{
					"order": "desc",
				},
			},
		},
		"size": 1,
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{StatsIndex},
		Body:  bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to search stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stats not found for hostname: %s", hostname)
	}

	var result struct {
		Hits struct {
			Hits []struct {
				Source statsDoc `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Hits.Hits) == 0 {
		return nil, fmt.Errorf("stats not found for hostname: %s", hostname)
	}

	return result.Hits.Hits[0].Source.toEntity(), nil
}

// GetAll retrieves all latest stats (one per hostname)
func (r *OpenSearchStatsRepository) GetAll(ctx context.Context) ([]*entity.Stats, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"aggs": map[string]interface{}{
			"hosts": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "hostname",
					"size":  1000,
				},
				"aggs": map[string]interface{}{
					"latest": map[string]interface{}{
						"top_hits": map[string]interface{}{
							"size": 1,
							"sort": []map[string]interface{}{
								{
									"timestamp": map[string]interface{}{
										"order": "desc",
									},
								},
							},
						},
					},
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
		Index: []string{StatsIndex},
		Body:  bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to search stats: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Aggregations struct {
			Hosts struct {
				Buckets []struct {
					Aggregations struct {
						Latest struct {
							Hits struct {
								Hits []struct {
									Source statsDoc `json:"_source"`
								} `json:"hits"`
							} `json:"hits"`
						} `json:"latest"`
					} `json:"aggregations"`
				} `json:"buckets"`
			} `json:"hosts"`
		} `json:"aggregations"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var allStats []*entity.Stats
	for _, bucket := range result.Aggregations.Hosts.Buckets {
		if len(bucket.Aggregations.Latest.Hits.Hits) > 0 {
			d := bucket.Aggregations.Latest.Hits.Hits[0].Source
			allStats = append(allStats, d.toEntity())
		}
	}

	return allStats, nil
}

// Delete removes all stats for a hostname
func (r *OpenSearchStatsRepository) Delete(ctx context.Context, hostname string) error {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"hostname": hostname,
			},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("failed to marshal query: %w", err)
	}

	req := opensearchapi.DeleteByQueryRequest{
		Index: []string{StatsIndex},
		Body:  bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return fmt.Errorf("failed to delete stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to delete stats: status %d", resp.StatusCode)
	}

	log.Printf("✓ Deleted stats for hostname: %s", hostname)
	return nil
}

// GetActiveHosts returns list of active hostnames
func (r *OpenSearchStatsRepository) GetActiveHosts(ctx context.Context) ([]string, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"aggs": map[string]interface{}{
			"unique_hosts": map[string]interface{}{
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
		Index: []string{StatsIndex},
		Body:  bytes.NewReader(body),
	}

	resp, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to search hosts: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Aggregations struct {
			UniqueHosts struct {
				Buckets []struct {
					Key string `json:"key"`
				} `json:"buckets"`
			} `json:"unique_hosts"`
		} `json:"aggregations"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	hosts := make([]string, 0, len(result.Aggregations.UniqueHosts.Buckets))
	for _, bucket := range result.Aggregations.UniqueHosts.Buckets {
		hosts = append(hosts, bucket.Key)
	}

	return hosts, nil
}

// SearchStats performs full-text search on stats
func (r *OpenSearchStatsRepository) SearchStats(ctx context.Context, query string, hostname string, limit int) ([]*entity.Stats, error) {
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

	if query != "" {
		mustQueries := searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})
		mustQueries = append(mustQueries, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"hostname", "agent_id", "ip_address"},
			},
		})
		searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = mustQueries
	}

	if hostname != "" {
		mustQueries := searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})
		mustQueries = append(mustQueries, map[string]interface{}{
			"term": map[string]interface{}{
				"hostname": hostname,
			},
		})
		searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = mustQueries
	}

	body, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{StatsIndex},
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
				Source statsDoc `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var allStats []*entity.Stats
	for i := range result.Hits.Hits {
		allStats = append(allStats, result.Hits.Hits[i].Source.toEntity())
	}

	return allStats, nil
}
