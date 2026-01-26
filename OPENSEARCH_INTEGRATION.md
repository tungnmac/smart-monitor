# OpenSearch Integration Guide

## Overview

The Smart Monitor Backend now integrates with OpenSearch for advanced searching and storage capabilities:

- **Full-text search** on stats, alerts, and events
- **Persistent storage** of monitoring data
- **Advanced aggregations** and analytics
- **Time-series data** optimization
- **Real-time metrics** storage and retrieval

## Starting OpenSearch

### Using Docker Compose

```bash
# Start OpenSearch and OpenSearch Dashboards
docker-compose -f docker-compose.opensearch.yml up -d

# Check status
docker-compose -f docker-compose.opensearch.yml logs -f opensearch

# Stop services
docker-compose -f docker-compose.opensearch.yml down
```

### Manual Setup

For production or custom setups:

```bash
# Install OpenSearch locally or in your infrastructure
# See: https://opensearch.org/docs/latest/install-and-configure/

# Default credentials:
# Username: admin
# Password: SmartMonitor@2024
```

## Configuration

Set environment variables to connect to OpenSearch:

```bash
# Backend environment variables
export OPENSEARCH_HOST=localhost
export OPENSEARCH_PORT=9200
export OPENSEARCH_USERNAME=admin
export OPENSEARCH_PASSWORD=SmartMonitor@2024
export OPENSEARCH_INSECURE_SKIP_VERIFY=true  # For development only

# Run backend
cd backend && go run cmd/server/main.go
```

## API Endpoints

### Search Stats

```bash
# Search stats data
curl -X GET "http://localhost:8080/search/stats?q=cpu&hostname=server1&limit=50"

# Response:
{
  "total": 10,
  "limit": 50,
  "query": "cpu",
  "result": [...]
}
```

### Search Alerts

```bash
# List alerts with filters
curl -X GET "http://localhost:8080/search/alerts?hostname=server1&severity=high&status=active&limit=20"

# Get alert statistics
curl -X GET "http://localhost:8080/search/alerts/stats"

# Create alert
curl -X POST "http://localhost:8080/search/alerts/create" \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "server1",
    "alert_type": "cpu_high",
    "severity": "high",
    "title": "High CPU Usage",
    "description": "CPU usage exceeded 80%",
    "value": 85.5,
    "threshold": 80
  }'

# Resolve alert
curl -X POST "http://localhost:8080/search/alerts/resolve?id=alert-id-here"
```

### Search Events

```bash
# List events with filters
curl -X GET "http://localhost:8080/search/events?type=security&level=warning&limit=50"

# Get event statistics
curl -X GET "http://localhost:8080/search/events/stats"

# Log event
curl -X POST "http://localhost:8080/search/events/log" \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "server1",
    "event_type": "security",
    "event_name": "unauthorized_access_attempt",
    "message": "Unauthorized access attempt detected",
    "source": "ssh",
    "user": "unknown",
    "level": "warning"
  }'

# Search events
curl -X GET "http://localhost:8080/search/events?q=unauthorized&hostname=server1"
```

## Index Mappings

### Stats Index

Stores system metrics with the following fields:
- `hostname`: Server hostname (keyword)
- `timestamp`: Data collection time (date)
- `cpu_usage`: CPU percentage (float)
- `cpu_count`: Number of CPU cores (integer)
- `memory_*`: Memory metrics (long)
- `disk_*`: Disk metrics (long)
- `network_*`: Network metrics (long)
- `load_average`: System load (float)
- `uptime`: System uptime (long)

### Alerts Index

Stores alert records with fields:
- `hostname`: Source server (keyword)
- `alert_type`: Type of alert (keyword)
- `severity`: Alert severity - critical, high, medium, low (keyword)
- `title`: Alert title (text)
- `description`: Alert message (text)
- `timestamp`: Alert creation time (date)
- `resolved_at`: Resolution time (date, optional)
- `status`: active or resolved (keyword)
- `value`: Current metric value (float)
- `threshold`: Alert threshold (float)
- `metadata`: Custom metadata (object)

### Events Index

Stores system/security events with fields:
- `hostname`: Source server (keyword)
- `event_type`: security, performance, availability, etc (keyword)
- `event_name`: Specific event name (keyword)
- `timestamp`: Event time (date)
- `message`: Event description (text)
- `source`: Event source (keyword)
- `user`: Associated user (keyword)
- `process_id`: Process ID (integer)
- `process_name`: Process name (keyword)
- `level`: info, warning, error, critical (keyword)
- `details`: Additional details (object)

## OpenSearch Dashboards

Access the visualization dashboard:

```
http://localhost:5601
```

Default credentials:
- Username: `admin`
- Password: `SmartMonitor@2024`

### Creating Dashboards

1. Go to Management → Dev Tools
2. Create index patterns for each index (stats, alerts, events)
3. Create visualizations:
   - Time-series charts for CPU/Memory trends
   - Alert severity distribution
   - Event type breakdown
   - Host activity metrics

## Performance Tuning

### For Development

```yaml
settings:
  number_of_shards: 1
  number_of_replicas: 0
```

### For Production

```yaml
settings:
  number_of_shards: 3
  number_of_replicas: 2
  index.lifecycle.name: "retention_policy"
```

## Backup and Retention

### Set up Index Lifecycle Policy

```bash
curl -X PUT "http://localhost:9200/_plugins/_ism/policies/retention_policy" \
  -H "Content-Type: application/json" \
  -d '{
    "policy": {
      "description": "Auto-delete old indexes",
      "default_state": "active",
      "states": [{
        "name": "active",
        "actions": [],
        "transitions": [{
          "state_name": "delete",
          "conditions": {
            "min_index_age": "30d"
          }
        }]
      }, {
        "name": "delete",
        "actions": [{
          "type": "delete"
        }]
      }]
    }
  }'
```

## Troubleshooting

### Connection Issues

```bash
# Test OpenSearch connectivity
curl -k -u admin:SmartMonitor@2024 https://localhost:9200/_cluster/health

# Check backend logs for errors
docker logs smart-monitor-opensearch
```

### Index Issues

```bash
# List all indexes
curl -k -u admin:SmartMonitor@2024 https://localhost:9200/_cat/indices

# Check index mapping
curl -k -u admin:SmartMonitor@2024 https://localhost:9200/stats/_mapping

# Delete an index
curl -X DELETE -k -u admin:SmartMonitor@2024 https://localhost:9200/stats
```

### Performance Issues

- Check disk space: `docker exec smart-monitor-opensearch df -h`
- Check heap memory: `curl -k -u admin:SmartMonitor@2024 https://localhost:9200/_nodes/stats`
- Monitor index size: `curl -k -u admin:SmartMonitor@2024 https://localhost:9200/_cat/indices?h=index,store.size`

## References

- [OpenSearch Documentation](https://opensearch.org/docs/)
- [OpenSearch Go Client](https://github.com/opensearch-project/opensearch-go)
- [Index Management Plugin](https://opensearch.org/docs/latest/im-plugin/index-state-management/index/)
