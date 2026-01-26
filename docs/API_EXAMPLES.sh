#!/bin/bash

# Smart Monitor API Examples
# This script demonstrates all the new agent control and policy management features

BASE_URL="http://localhost:8080"

echo "=== Smart Monitor API Testing ==="
echo ""

# 1. Register a test agent
echo "1. Registering test agent..."
REGISTER_RESPONSE=$(curl -s -X POST ${BASE_URL}/v1/agent/register \
  -H "Content-Type: application/json" \
  -d '{
    "agent_name": "demo-agent",
    "hostname": "demo-host",
    "os": "linux",
    "ip_address": "192.168.1.200"
  }')
echo "$REGISTER_RESPONSE" | jq .
AGENT_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.agentId')
echo "Agent ID: $AGENT_ID"
echo ""

# 2. Create a policy
echo "2. Creating monitoring policy..."
POLICY_RESPONSE=$(curl -s -X POST ${BASE_URL}/v1/policies \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Demo CPU Alert",
    "description": "Alert when CPU exceeds 85%",
    "thresholds": {"cpu_usage": "85", "memory_usage": "90"},
    "actions": ["email", "webhook", "slack"],
    "enabled": true
  }')
echo "$POLICY_RESPONSE" | jq .
POLICY_ID=$(echo "$POLICY_RESPONSE" | jq -r '.policyId')
echo "Policy ID: $POLICY_ID"
echo ""

# 3. List all policies
echo "3. Listing all policies..."
curl -s -X GET "${BASE_URL}/v1/policies?page=1&page_size=10" | jq .
echo ""

# 4. Update the policy
echo "4. Updating policy thresholds..."
curl -s -X PUT ${BASE_URL}/v1/policies/${POLICY_ID} \
  -H "Content-Type: application/json" \
  -d '{
    "id": "'$POLICY_ID'",
    "name": "Demo Critical CPU Alert",
    "description": "Alert when CPU exceeds 90%",
    "thresholds": {"cpu_usage": "90", "memory_usage": "95"},
    "actions": ["email", "webhook", "slack", "pagerduty"],
    "enabled": true
  }' | jq .
echo ""

# 5. Apply policy to agent
echo "5. Applying policy to agent..."
curl -s -X POST ${BASE_URL}/v1/agent/${AGENT_ID}/policy/${POLICY_ID}/apply \
  -H "Content-Type: application/json" \
  -d '{}' | jq .
echo ""

# 6. Verify policy is applied
echo "6. Verifying policy application..."
curl -s -X GET "${BASE_URL}/v1/policies?page=1&page_size=10" | jq '.policies[] | select(.policyId == "'$POLICY_ID'")'
echo ""

# 7. Control agent - restart
echo "7. Sending restart command to agent..."
curl -s -X POST ${BASE_URL}/v1/agent/${AGENT_ID}/control \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "'$AGENT_ID'",
    "action": "restart",
    "reason": "applying new configuration"
  }' | jq .
echo ""

# 8. Control agent - shutdown
echo "8. Sending shutdown command to agent..."
curl -s -X POST ${BASE_URL}/v1/agent/${AGENT_ID}/control \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "'$AGENT_ID'",
    "action": "shutdown",
    "reason": "maintenance window"
  }' | jq .
echo ""

# 9. Block agent
echo "9. Blocking agent..."
curl -s -X POST ${BASE_URL}/v1/agent/${AGENT_ID}/block \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "'$AGENT_ID'",
    "blocked": true,
    "reason": "suspicious activity detected"
  }' | jq .
echo ""

# 10. Try to control blocked agent (should fail)
echo "10. Attempting to control blocked agent (should fail)..."
curl -s -X POST ${BASE_URL}/v1/agent/${AGENT_ID}/control \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "'$AGENT_ID'",
    "action": "start",
    "reason": "test"
  }' | jq .
echo ""

# 11. Unblock agent
echo "11. Unblocking agent..."
curl -s -X POST ${BASE_URL}/v1/agent/${AGENT_ID}/block \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "'$AGENT_ID'",
    "blocked": false,
    "reason": "investigation complete"
  }' | jq .
echo ""

# 12. Unapply policy from agent
echo "12. Removing policy from agent..."
curl -s -X POST ${BASE_URL}/v1/agent/${AGENT_ID}/policy/${POLICY_ID}/unapply \
  -H "Content-Type: application/json" \
  -d '{}' | jq .
echo ""

# 13. Delete policy
echo "13. Deleting policy..."
curl -s -X DELETE ${BASE_URL}/v1/policies/${POLICY_ID} | jq .
echo ""

# 14. List policies again (should be empty or policy removed)
echo "14. Final policy list..."
curl -s -X GET "${BASE_URL}/v1/policies?page=1&page_size=10" | jq .
echo ""

echo "=== Testing Complete ==="
