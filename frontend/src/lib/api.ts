const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:50051";

export interface Agent {
  id: string;
  host: string;
  env: string;
  region: string;
  status: "online" | "offline" | "degraded";
  version: string;
  cpu: number;
  ram: number;
  disk: number;
  ip_address: string;
  last_seen?: number;
}

export interface Metrics {
  hostname: string;
  agent_id: string;
  cpu: number;
  ram: number;
  disk: number;
  timestamp: number;
  ip_address?: string;
}

export interface Policy {
  policy_id: string;
  name: string;
  description: string;
  thresholds: Record<string, string>;
  actions: string[];
  enabled: boolean;
  applied_agents: string[];
  created_at?: number;
  updated_at?: number;
}

// Fetch agents list
export async function fetchAgents(): Promise<Agent[]> {
  try {
    const res = await fetch(`${BACKEND_URL}/v1/agents`, {
      method: "GET",
      headers: { "Content-Type": "application/json" },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    return data.agents || [];
  } catch (err) {
    console.error("Failed to fetch agents:", err);
    return [];
  }
}

// Fetch metrics for a specific agent/host
export async function fetchMetrics(hostname: string): Promise<Metrics | null> {
  try {
    const res = await fetch(`${BACKEND_URL}/v1/stats/${hostname}`, {
      method: "GET",
      headers: { "Content-Type": "application/json" },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    return data as Metrics;
  } catch (err) {
    console.error(`Failed to fetch metrics for ${hostname}:`, err);
    return null;
  }
}

// Stream metrics using EventSource (SSE)
export function streamMetrics(
  onMetrics: (metrics: Metrics) => void,
  onError: (error: Error) => void
): () => void {
  const eventSource = new EventSource(`${BACKEND_URL}/v1/stats/stream`);

  eventSource.onmessage = (event) => {
    try {
      const metrics = JSON.parse(event.data);
      onMetrics(metrics);
    } catch (err) {
      console.error("Failed to parse metrics:", err);
    }
  };

  eventSource.onerror = () => {
    onError(new Error("Connection lost"));
    eventSource.close();
  };

  return () => eventSource.close();
}

// Fetch policies
export async function fetchPolicies(): Promise<Policy[]> {
  try {
    const res = await fetch(`${BACKEND_URL}/v1/policies`, {
      method: "GET",
      headers: { "Content-Type": "application/json" },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    return data.policies || [];
  } catch (err) {
    console.error("Failed to fetch policies:", err);
    return [];
  }
}

// Apply policy to agent
export async function applyPolicy(
  agentId: string,
  policyId: string
): Promise<boolean> {
  try {
    const res = await fetch(
      `${BACKEND_URL}/v1/agent/${agentId}/policy/${policyId}/apply`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
      }
    );
    return res.ok;
  } catch (err) {
    console.error("Failed to apply policy:", err);
    return false;
  }
}

// Control agent (start, restart, shutdown)
export async function controlAgent(
  agentId: string,
  action: "start" | "restart" | "shutdown"
): Promise<boolean> {
  try {
    const res = await fetch(`${BACKEND_URL}/v1/agent/${agentId}/control`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action }),
    });
    return res.ok;
  } catch (err) {
    console.error("Failed to control agent:", err);
    return false;
  }
}

// Block/unblock agent
export async function blockAgent(
  agentId: string,
  blocked: boolean
): Promise<boolean> {
  try {
    const res = await fetch(`${BACKEND_URL}/v1/agent/${agentId}/block`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ blocked }),
    });
    return res.ok;
  } catch (err) {
    console.error("Failed to block agent:", err);
    return false;
  }
}
