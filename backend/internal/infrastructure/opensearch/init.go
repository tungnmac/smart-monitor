// Package opensearch provides index initialization
package opensearch

import (
	"context"
	"fmt"
	"log"
	"time"
)

// InitializeIndexes creates all necessary indexes in OpenSearch
func InitializeIndexes(client *Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []struct {
		name    string
		mapping string
	}{
		{StatsIndex, StatsIndexMapping},
		{AlertsIndex, AlertsIndexMapping},
		{EventsIndex, EventsIndexMapping},
	}

	for _, idx := range indexes {
		exists, err := client.IndexExists(ctx, idx.name)
		if err != nil {
			return fmt.Errorf("failed to check index %s: %w", idx.name, err)
		}

		if !exists {
			if err := client.CreateIndex(ctx, idx.name, idx.mapping); err != nil {
				return fmt.Errorf("failed to create index %s: %w", idx.name, err)
			}
			log.Printf("✓ Created index: %s", idx.name)
		} else {
			log.Printf("✓ Index already exists: %s", idx.name)
		}
	}

	return nil
}
