This is the Smart Monitor Console - a Next.js 14 frontend with NextAuth and Tailwind CSS for centralized agent management, monitoring, analytics, detection, protection, and prevention.

## Features

- **Authentication**: Credentials-based login with JWT sessions  
- **Dark Theme**: Modern gradient design with glassmorphic cards  
- **Backend Integration**: HTTP REST API calls + real-time metrics streaming
- **6 Core Sections**:
  - **Agents**: Fleet directory, registration, process control (HTTP polling)
  - **Monitor**: Live telemetry (CPU/RAM/Disk), real-time stream via EventSource
  - **Analytics**: 7-day trends, anomaly scoring, policy efficacy  
  - **Detect**: Real-time alerts, incident tracking, customizable rules
  - **Protect**: Policy enforcement, response automation, templates (HTTP fetch)
  - **Prevent**: Compliance scoring, guardrails, audit logs

## Getting Started

Install dependencies:
```bash
cd frontend
npm install
```

Set environment variables in `.env.local`:
```bash
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=$(openssl rand -base64 32)
AUTH_USERNAME=admin
AUTH_PASSWORD=changeme
NEXT_PUBLIC_BACKEND_URL=http://localhost:50051
```

Run the development server:
```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) and log in with `admin` / `changeme`.

## API Integration

### Backend Communication

- **Agents Page**: Fetches agent list via HTTP GET `/v1/agents` (refreshes every 5s)
- **Monitor Page**: Streams metrics via EventSource `/v1/stats/stream` (real-time)
- **Protect Page**: Fetches policies via HTTP GET `/v1/policies` (refreshes every 10s)
- **Actions**: HTTP POST for control (restart/block) operations

### Hooks for Data Fetching

Located in `lib/hooks.ts`:

- `useAgents()` - Polls agent list
- `useMetricsStream()` - Real-time metrics streaming
- `usePolicies()` - Polls policies
- `useAgentControl()` - Agent control operations

### API Client

Located in `lib/api.ts` - low-level API functions:

- `fetchAgents()`, `fetchMetrics()`, `fetchPolicies()`
- `streamMetrics()` - EventSource subscription
- `controlAgent()`, `blockAgent()`, `applyPolicy()`

## Project Structure

```
src/
├── app/
│   ├── page.tsx              # Landing page
│   ├── layout.tsx            # Root layout with providers
│   ├── globals.css           # Global styles (dark gradient)
│   ├── (auth)/login/         # Login page
│   ├── api/auth/[...nextauth]/ # NextAuth route
│   └── dashboard/            # Protected dashboard
│       ├── page.tsx          # Overview
│       ├── agents/           # Agent management (HTTP polling)
│       ├── monitor/          # Live metrics
│       ├── analytics/        # Trends
│       ├── detect/           # Anomalies
│       ├── protect/          # Policies
│       └── prevent/          # Compliance
├── lib/auth.ts              # NextAuth config
└── types/next-auth.d.ts     # Type defs
```

## Build & Deploy

```bash
npm run build
npm start
```

## Notes

- All dashboard routes protected by middleware
- Sample data inline (ready for gRPC API integration)
- Responsive mobile-first design
- Glass morphism effects via Tailwind utility classes
