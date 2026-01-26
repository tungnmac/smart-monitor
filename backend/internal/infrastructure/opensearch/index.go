// Package opensearch provides index definitions and mappings
package opensearch

// StatsIndexMapping defines the mapping for stats index
const StatsIndexMapping = `{
  "settings": {
    "number_of_shards": 2,
    "number_of_replicas": 1,
    "index.lifecycle.name": "stats_policy",
    "index.lifecycle.rollover_alias": "stats"
  },
  "mappings": {
    "properties": {
      "hostname": {"type": "keyword"},
      "agent_id": {"type": "keyword"},
      "ip_address": {"type": "keyword"},
      "cpu": {"type": "float"},
      "ram": {"type": "float"},
      "disk": {"type": "float"},
      "timestamp": {"type": "date", "format": "epoch_millis"},
      "last_received": {"type": "date", "format": "epoch_millis"},
      "metadata": {"type": "object"}
    }
  }
}`

// AlertsIndexMapping defines the mapping for alerts index
const AlertsIndexMapping = `{
  "settings": {
    "number_of_shards": 2,
    "number_of_replicas": 1
  },
  "mappings": {
    "properties": {
      "hostname": {
        "type": "keyword"
      },
      "alert_type": {
        "type": "keyword"
      },
      "severity": {
        "type": "keyword"
      },
      "title": {
        "type": "text"
      },
      "description": {
        "type": "text"
      },
      "timestamp": {
        "type": "date",
        "format": "epoch_millis"
      },
      "resolved_at": {
        "type": "date",
        "format": "epoch_millis"
      },
      "status": {
        "type": "keyword"
      },
      "value": {
        "type": "float"
      },
      "threshold": {
        "type": "float"
      },
      "metadata": {
        "type": "object"
      }
    }
  }
}`

// EventsIndexMapping defines the mapping for events index
const EventsIndexMapping = `{
  "settings": {
    "number_of_shards": 2,
    "number_of_replicas": 1
  },
  "mappings": {
    "properties": {
      "hostname": {
        "type": "keyword"
      },
      "event_type": {
        "type": "keyword"
      },
      "event_name": {
        "type": "keyword"
      },
      "timestamp": {
        "type": "date",
        "format": "epoch_millis"
      },
      "message": {
        "type": "text"
      },
      "source": {
        "type": "keyword"
      },
      "user": {
        "type": "keyword"
      },
      "process_id": {
        "type": "integer"
      },
      "process_name": {
        "type": "keyword"
      },
      "level": {
        "type": "keyword"
      },
      "details": {
        "type": "object"
      }
    }
  }
}`

// Index names
const (
	StatsIndex  = "stats"
	AlertsIndex = "alerts"
	EventsIndex = "events"
)
