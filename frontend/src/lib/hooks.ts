"use client";

import { useEffect, useState, useCallback } from "react";
import type { Agent, Metrics, Policy } from "@/lib/api";
import {
  fetchAgents,
  fetchMetrics,
  fetchPolicies,
  streamMetrics,
  controlAgent,
  blockAgent,
} from "@/lib/api";

export function useAgents() {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      const data = await fetchAgents();
      if (data) {
        setAgents(data);
      } else {
        setError("Failed to load agents");
      }
      setLoading(false);
    };

    load();
    const timer = setInterval(load, 5000); // Refresh every 5s
    return () => clearInterval(timer);
  }, []);

  return { agents, loading, error };
}

export function useMetrics(hostname?: string) {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!hostname) return;

    const load = async () => {
      setLoading(true);
      const data = await fetchMetrics(hostname);
      setMetrics(data);
      setLoading(false);
    };

    load();
    const timer = setInterval(load, 2000);
    return () => clearInterval(timer);
  }, [hostname]);

  return { metrics, loading, error };
}

export function useMetricsStream() {
  const [allMetrics, setAllMetrics] = useState<Map<string, Metrics>>(new Map());
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const unsubscribe = streamMetrics(
      (metrics) => {
        setAllMetrics((prev) => {
          const next = new Map(prev);
          next.set(metrics.hostname, metrics);
          return next;
        });
      },
      (err) => setError(err.message)
    );

    return unsubscribe;
  }, []);

  return { allMetrics: Array.from(allMetrics.values()), error };
}

export function usePolicies() {
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      const data = await fetchPolicies();
      if (data) {
        setPolicies(data);
      } else {
        setError("Failed to load policies");
      }
      setLoading(false);
    };

    load();
    const timer = setInterval(load, 10000); // Refresh every 10s
    return () => clearInterval(timer);
  }, []);

  return { policies, loading, error };
}

export function useAgentControl() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const restart = useCallback(async (agentId: string) => {
    setLoading(true);
    setError(null);
    const ok = await controlAgent(agentId, "restart");
    setLoading(false);
    if (!ok) setError("Failed to restart agent");
    return ok;
  }, []);

  const block = useCallback(async (agentId: string) => {
    setLoading(true);
    setError(null);
    const ok = await blockAgent(agentId, true);
    setLoading(false);
    if (!ok) setError("Failed to block agent");
    return ok;
  }, []);

  const unblock = useCallback(async (agentId: string) => {
    setLoading(true);
    setError(null);
    const ok = await blockAgent(agentId, false);
    setLoading(false);
    if (!ok) setError("Failed to unblock agent");
    return ok;
  }, []);

  return { restart, block, unblock, loading, error };
}
