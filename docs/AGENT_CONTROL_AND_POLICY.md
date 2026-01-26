# Agent Control and Policy Management Features

## Overview
This document describes the new agent control and policy management features added to the Smart Monitor backend.

## New Features

### 1. Agent Control Operations
Control agent lifecycle through gRPC API endpoints.

#### Available Actions
- **Start**: Start a stopped agent
- **Shutdown**: Gracefully shutdown an agent  
- **Restart**: Restart an agent

#### API Endpoint
```
POST /v1/agent/{agent_id}/control
```

#### Request
```json
{
  "agent_id": "agent-123",
  "action": "restart",
  "reason": "applying new configuration"
}
```

#### Response
```json
{
  "success": true,
  "message": "Control command sent successfully",
  "agent_id": "agent-123",
  "action": "restart",
  "timestamp": 1706276400
}
```

### 2. Agent Blocking
Block or unblock agents from sending data to the system.

#### API Endpoint
```
POST /v1/agent/{agent_id}/block
```

#### Request
```json
{
  "agent_id": "agent-123",
  "blocked": true,
  "reason": "suspicious activity detected"
}
```

#### Response
```json
{
  "success": true,
  "message": "Agent blocked successfully",
  "agent_id": "agent-123",
  "blocked": true,
  "timestamp": 1706276400
}
```

### 3. Policy Management

#### 3.1 Add Policy
Create a new monitoring policy with thresholds and actions.

**API Endpoint**: `POST /v1/policies`

**Request**:
```json
{
  "name": "High CPU Alert",
  "description": "Alert when CPU usage exceeds threshold",
  "thresholds": {
    "cpu_usage": "80",
    "memory_usage": "90"
  },
  "actions": ["email", "webhook"],
  "enabled": true
}
```

**Response**:
```json
{
  "success": true,
  "message": "Policy created successfully",
  "policy_id": "policy-abc123",
  "timestamp": 1706276400
}
```

#### 3.2 Update Policy
Update an existing policy.

**API Endpoint**: `PUT /v1/policies/{policy_id}`

**Request**:
```json
{
  "id": "policy-abc123",
  "name": "Critical CPU Alert",
  "thresholds": {
    "cpu_usage": "90",
    "memory_usage": "95"
  },
  "enabled": true
}
```

#### 3.3 Remove Policy
Delete a policy from the system.

**API Endpoint**: `DELETE /v1/policies/{policy_id}`

#### 3.4 List Policies
Get all policies with pagination.

**API Endpoint**: `GET /v1/policies?page=1&page_size=10`

**Response**:
```json
{
  "policies": [
    {
      "id": "policy-abc123",
      "name": "High CPU Alert",
      "description": "Alert when CPU usage exceeds threshold",
      "enabled": true,
      "created_at": 1706276400,
      "applied_agents": []
    }
  ],
  "total": 1
}
```

#### 3.5 Apply Policy to Agent
Apply a policy to a specific agent.

**API Endpoint**: `POST /v1/agent/{agent_id}/policy/{policy_id}/apply`

**Response**:
```json
{
  "success": true,
  "message": "Policy applied to agent successfully",
  "policy_id": "policy-abc123",
  "timestamp": 1706276400
}
```

#### 3.6 Unapply Policy from Agent
Remove a policy from a specific agent.

**API Endpoint**: `POST /v1/agent/{agent_id}/policy/{policy_id}/unapply`

**Response**:
```json
{
  "success": true,
  "message": "Policy unapplied from agent successfully",
  "policy_id": "policy-abc123",
  "timestamp": 1706276400
}
```

## Architecture

### Domain Layer

#### Entities
1. **Policy** (`backend/internal/domain/entity/policy.go`)
   - Represents a monitoring policy with thresholds and actions
   - Methods: `NewPolicy()`, `Update()`, `Enable()`, `Disable()`, `ApplyToAgent()`, `UnapplyFromAgent()`

2. **AgentRegistry** (Enhanced)
   - Added blocking capability
   - New fields: `Blocked bool`, `BlockReason string`
   - New methods: `Block()`, `Unblock()`, `IsBlocked()`
   - New status: `AgentStatusBlocked`

#### Repositories
1. **PolicyRepository** (`backend/internal/domain/repository/policy_repository.go`)
   - Interface for policy persistence
   - Methods: `Create()`, `GetByID()`, `Update()`, `Delete()`, `List()`, `ApplyToAgent()`, `UnapplyFromAgent()`, `GetByAgentID()`

2. **InMemoryPolicyRepository** (`backend/internal/infrastructure/persistence/policy_repository.go`)
   - In-memory implementation of PolicyRepository
   - Thread-safe using mutex
   - Supports pagination

#### Services
1. **AgentControlService** (`backend/internal/domain/service/agent_control_service.go`)
   - Business logic for agent control operations
   - Methods: `ControlAgent()`, `BlockAgent()`, `UnblockAgent()`, `GetAgentStatus()`
   - Validates agent status before actions

2. **PolicyService** (`backend/internal/domain/service/policy_service.go`)
   - Business logic for policy management
   - Methods: `CreatePolicy()`, `UpdatePolicy()`, `RemovePolicy()`, `ListPolicies()`, `ApplyPolicyToAgent()`, `UnapplyPolicyFromAgent()`
   - Generates unique policy IDs
   - Validates agents before applying policies

### Infrastructure Layer

#### gRPC Handlers (`backend/internal/infrastructure/grpc/monitor_handler.go`)
New RPC methods implemented:
1. `ControlAgent` - Send control commands to agents
2. `BlockAgent` - Block/unblock agents
3. `AddPolicy` - Create new policies
4. `UpdatePolicy` - Update existing policies
5. `RemovePolicy` - Delete policies
6. `ListPolicies` - List all policies with pagination
7. `ApplyPolicy` - Apply policy to agent
8. `UnapplyPolicy` - Remove policy from agent

### Protocol Buffers

#### Proto Definitions (`pbtypes/monitor/monitor.proto`)
New message types:
- `ControlAgentRequest` / `ControlAgentResponse`
- `BlockAgentRequest` / `BlockAgentResponse`
- `PolicyRequest` / `PolicyResponse`
- `RemovePolicyRequest`
- `ListPoliciesRequest` / `ListPoliciesResponse`
- `Policy`
- `ApplyPolicyRequest` / `UnapplyPolicyRequest`

## Testing

### Using gRPC Client
```bash
# Control agent
grpcurl -plaintext -d '{
  "agent_id": "agent-123",
  "action": "restart",
  "reason": "test"
}' localhost:50051 monitor.MonitorService/ControlAgent

# Block agent
grpcurl -plaintext -d '{
  "agent_id": "agent-123",
  "blocked": true,
  "reason": "testing"
}' localhost:50051 monitor.MonitorService/BlockAgent

# Add policy
grpcurl -plaintext -d '{
  "name": "Test Policy",
  "description": "Test policy for demo",
  "thresholds": {"cpu": "80"},
  "actions": ["email"],
  "enabled": true
}' localhost:50051 monitor.MonitorService/AddPolicy
```

### Using HTTP Gateway
```bash
# Control agent
curl -X POST http://localhost:8080/v1/agent/control \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"agent-123","action":"restart","reason":"test"}'

# Block agent
curl -X POST http://localhost:8080/v1/agent/block \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"agent-123","blocked":true,"reason":"testing"}'

# List policies
curl http://localhost:8080/v1/policy/list?page=1&page_size=10
```

### Using Swagger UI
Navigate to http://localhost:8080/swagger/ to interact with the API through Swagger UI.

## Implementation Notes

1. **Thread Safety**: All repositories use mutexes for thread-safe operations
2. **Validation**: Services validate agent status and existence before operations
3. **Error Handling**: Proper error messages returned for all failure scenarios
4. **Logging**: All operations are logged for audit and debugging
5. **Context**: All repository methods accept `context.Context` for cancellation support
6. **Pagination**: Policy listing supports pagination for large datasets

## Build and Run

```bash
# Generate proto files
make gen-proto

# Build backend
make build-backend

# Run backend
make run-backend
```

The server will start on:
- gRPC: `localhost:50051`
- HTTP Gateway: `localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/`

## Future Enhancements{agent_id}/control \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"agent-123","action":"restart","reason":"test"}'

# Block agent
curl -X POST http://localhost:8080/v1/agent/{agent_id}/block \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"agent-123","blocked":true,"reason":"testing"}'

# Add policy
curl -X POST http://localhost:8080/v1/policies \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Policy","description":"Test","thresholds":{"cpu":"80"},"actions":["email"],"enabled":true}'

# List policies
curl http://localhost:8080/v1/policies?page=1&page_size=10

# Apply policy to agent
curl -X POST http://localhost:8080/v1/agent/{agent_id}/policy/{policy_id}/apply

# Complete test script
./docs/API_EXAMPLES.sh