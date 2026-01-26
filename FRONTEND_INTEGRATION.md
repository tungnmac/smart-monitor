# Frontend Backend Integration Summary

## Overview

Updated the Smart Monitor frontend to call backend HTTP APIs for:
- **Agents Management** - HTTP polling for agent list
- **Monitor Metrics** - Real-time streaming via EventSource (SSE)
- **Protection Policies** - HTTP polling for policies
- **Agent Control** - HTTP POST for restart/block actions

## Key Files Added/Updated

### API Layer (`src/lib/`)

**`api.ts`** - Core HTTP client
- `fetchAgents()` - GET `/v1/agents`
- `fetchMetrics(hostname)` - GET `/v1/stats/{hostname}`
- `streamMetrics(onMetrics, onError)` - EventSource `/v1/stats/stream`
- `fetchPolicies()` - GET `/v1/policies`
- `controlAgent(agentId, action)` - POST `/v1/agent/{agentId}/control`
- `blockAgent(agentId, blocked)` - POST `/v1/agent/{agentId}/block`
- `applyPolicy(agentId, policyId)` - POST `/v1/agent/{agentId}/policy/{policyId}/apply`

**`hooks.ts`** - React hooks for data fetching
- `useAgents()` - Polls every 5s
- `useMetricsStream()` - Real-time streaming
- `useMetrics(hostname)` - Polls every 2s for single host
- `usePolicies()` - Polls every 10s
- `useAgentControl()` - Manages restart/block operations

### Dashboard Pages Updated

**`/dashboard/agents`** - Agent management
- Fetches agent list via `useAgents()`
- Restart/block controls via `useAgentControl()`
- Live metric display (CPU, RAM, Disk)
- Status color coding based on agent state

**`/dashboard/monitor`** - Real-time metrics
- Real-time streaming via `useMetricsStream()` using EventSource
- Live chart with last 8 metric updates
- Per-agent breakdown with status colors
- Average calculations across fleet

**`/dashboard/protect`** - Policy management
- Fetches policies via `usePolicies()`
- Shows policy status and applied agent count
- Summary statistics (active policies, total applied)

### Other Updates

**`src/app/(auth)/login/page.tsx`** - Fixed Suspense boundary
- Wrapped `useSearchParams()` in Suspense for Next.js 16+ compatibility

**`.env.local` configuration**
```bash
NEXT_PUBLIC_BACKEND_URL=http://localhost:50051
AUTH_USERNAME=admin
AUTH_PASSWORD=changeme
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=<generated>
```

## Real-Time Streaming Implementation

Monitor page uses **Server-Sent Events (EventSource)** for real-time metrics:

```typescript
// In useMetricsStream hook
const eventSource = new EventSource(`${BACKEND_URL}/v1/stats/stream`);
eventSource.onmessage = (event) => {
  const metrics = JSON.parse(event.data);
  setAllMetrics(prev => new Map(prev).set(metrics.hostname, metrics));
};
```

Benefits:
- One-directional connection (ideal for pushing updates from server)
- Auto-reconnect on network failure
- Lower overhead than WebSocket for simple streams
- Works through HTTP proxies

## Build Status

✅ Frontend builds successfully
✅ All routes properly configured
✅ NextAuth authentication working
✅ API client ready for backend integration

## Next Steps

1. Configure `NEXT_PUBLIC_BACKEND_URL` to point to actual backend
2. Ensure backend exposes HTTP endpoints:
   - GET `/v1/agents`
   - GET `/v1/stats/{hostname}`
   - SSE `/v1/stats/stream`
   - GET `/v1/policies`
   - POST `/v1/agent/{agentId}/control`
   - POST `/v1/agent/{agentId}/block`
3. Test frontend by running:
   ```bash
   npm run dev
   # Then visit http://localhost:3000
   ```

## Tech Stack

- **Framework**: Next.js 14 with App Router
- **Styling**: Tailwind CSS 4 with custom dark theme
- **Auth**: NextAuth.js v5 with Credentials provider
- **Data Fetching**: Native Fetch API + EventSource for streaming
- **Real-time**: Server-Sent Events (SSE/EventSource)
