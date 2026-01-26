// Package opensearch provides OpenSearch client initialization and utilities
package opensearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

// Client wraps OpenSearch client with utility methods
type Client struct {
	*opensearch.Client
}

// NewClient creates a new OpenSearch client
func NewClient(host string, port int, username, password string, insecureSkipVerify bool) (*Client, error) {
	cfg := opensearch.Config{
		Addresses: []string{fmt.Sprintf("https://%s:%d", host, port)},
		Username:  username,
		Password:  password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureSkipVerify,
			},
		},
	}

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Info request to validate connectivity
	infoReq := opensearchapi.InfoRequest{}
	resp, err := infoReq.Do(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to OpenSearch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenSearch returned status %d", resp.StatusCode)
	}

	log.Println("✓ Connected to OpenSearch successfully")
	return &Client{Client: client}, nil
}

// CreateIndex creates an index with mapping
func (c *Client) CreateIndex(ctx context.Context, indexName string, mapping string) error {
	req := opensearchapi.IndicesCreateRequest{
		Index: indexName,
		Body:  strings.NewReader(mapping),
	}

	resp, err := req.Do(ctx, c.Client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("failed to create index: status %d", resp.StatusCode)
	}

	log.Printf("✓ Index '%s' created/verified", indexName)
	return nil
}

// IndexExists checks if index exists
func (c *Client) IndexExists(ctx context.Context, indexName string) (bool, error) {
	req := opensearchapi.IndicesExistsRequest{
		Index: []string{indexName},
	}

	resp, err := req.Do(ctx, c.Client)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// DeleteIndex deletes an index
func (c *Client) DeleteIndex(ctx context.Context, indexName string) error {
	req := opensearchapi.IndicesDeleteRequest{
		Index: []string{indexName},
	}

	resp, err := req.Do(ctx, c.Client)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete index: status %d", resp.StatusCode)
	}

	log.Printf("✓ Index '%s' deleted", indexName)
	return nil
}

// Close is a no-op for OpenSearch client (kept for symmetry)
func (c *Client) Close() error { return nil }
