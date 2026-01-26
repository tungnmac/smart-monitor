"use client";

import { useMetricsStream } from "@/lib/hooks";
import { useState, useEffect } from "react";

export default function MonitorPage() {
  const { allMetrics, error } = useMetricsStream();
  const [avgCpu, setAvgCpu] = useState(0);
  const [avgRam, setAvgRam] = useState(0);
  const [avgDisk, setAvgDisk] = useState(0);

  useEffect(() => {
    if (allMetrics.length === 0) return;

    const totalCpu = allMetrics.reduce((sum, m) => sum + m.cpu, 0) / allMetrics.length;
    const totalRam = allMetrics.reduce((sum, m) => sum + m.ram, 0) / allMetrics.length;
    const totalDisk = allMetrics.reduce((sum, m) => sum + m.disk, 0) / allMetrics.length;

    setAvgCpu(totalCpu);
    setAvgRam(totalRam);
    setAvgDisk(totalDisk);
  }, [allMetrics]);

  const getStatusColor = (value: number) => {
    if (value >= 80) return "text-rose-200";
    if (value >= 60) return "text-amber-200";
    return "text-emerald-200";
  };

  return (
    <div className="space-y-4">
      <div className="rounded-3xl border border-white/10 bg-slate-900/60 p-6 shadow-xl">
        <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Live Telemetry</p>
        <h2 className="text-xl font-semibold text-white">System metrics dashboard</h2>
        <p className="text-sm text-slate-300">Real-time CPU, RAM, Disk with streaming updates • {allMetrics.length} agents</p>

        {error && (
          <div className="mt-3 rounded-lg border border-amber-500/30 bg-amber-900/30 px-3 py-2 text-xs text-amber-100">
            Stream: {error}
          </div>
        )}

        <div className="mt-6 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {[
            { label: "Avg CPU", value: avgCpu.toFixed(1), icon: "📊" },
            { label: "Avg RAM", value: avgRam.toFixed(1), icon: "💾" },
            { label: "Avg Disk", value: avgDisk.toFixed(1), icon: "💿" },
            { label: "Active agents", value: allMetrics.length, icon: "🌐" },
          ].map((m) => (
            <div
              key={m.label}
              className="rounded-2xl border border-white/10 bg-slate-800/50 p-4 text-center text-slate-100"
            >
              <div className="text-2xl">{m.icon}</div>
              <p className="mt-2 text-xs uppercase tracking-wide text-slate-400">{m.label}</p>
              <p className={`text-2xl font-semibold ${getStatusColor(typeof m.value === 'number' ? m.value : parseFloat(m.value as string))}`}>
                {m.value}%
              </p>
            </div>
          ))}
        </div>

        <div className="mt-6 space-y-2 rounded-2xl border border-white/10 bg-slate-800/40 p-4">
          <p className="text-xs uppercase tracking-wide text-slate-400">Live metrics stream (last 8 updates)</p>
          <div className="h-32 rounded-lg border border-white/10 bg-slate-900/30 p-4 flex items-end justify-around gap-1">
            {allMetrics.slice(-8).map((m, i) => (
              <div
                key={i}
                className="w-full max-w-16 rounded-t-lg bg-indigo-500/60 relative group"
                style={{ height: `${Math.max(m.cpu, 10)}%` }}
                title={`${m.hostname}: CPU ${m.cpu.toFixed(1)}%`}
              >
                <div className="absolute -top-6 left-0 right-0 text-xs text-slate-300 opacity-0 group-hover:opacity-100 text-center">
                  {m.cpu.toFixed(0)}%
                </div>
              </div>
            ))}
          </div>
          <p className="text-xs text-slate-400">← Real-time CPU from streaming agents</p>
        </div>

        <div className="mt-4 rounded-2xl border border-white/10 bg-slate-800/40 p-4">
          <p className="text-xs uppercase tracking-wide text-slate-300 mb-3">Per-agent breakdown</p>
          <div className="space-y-2 max-h-48 overflow-y-auto">
            {allMetrics.map((m, i) => (
              <div
                key={i}
                className="flex items-center justify-between rounded-lg border border-white/10 bg-slate-900/30 px-3 py-2 text-xs text-slate-200"
              >
                <div>
                  <p className="font-semibold text-white">{m.hostname}</p>
                  <p className="text-slate-400">{m.agent_id}</p>
                </div>
                <div className="flex gap-3">
                  <span className={getStatusColor(m.cpu)}>CPU {m.cpu.toFixed(1)}%</span>
                  <span className={getStatusColor(m.ram)}>RAM {m.ram.toFixed(1)}%</span>
                  <span className={getStatusColor(m.disk)}>Disk {m.disk.toFixed(1)}%</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="rounded-3xl border border-white/10 bg-slate-900/60 p-6 shadow-xl">
        <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Fleet Baselines</p>
        <h2 className="text-xl font-semibold text-white">Anomaly thresholds</h2>
        <div className="mt-4 space-y-2">
          {[
            { metric: "CPU spike", threshold: "80%", deviation: "+2σ" },
            { metric: "RAM peak", threshold: "90%", deviation: "+1.8σ" },
            { metric: "Disk usage", threshold: "85%", deviation: "+1.5σ" },
          ].map((a) => (
            <div
              key={a.metric}
              className="flex items-center justify-between rounded-lg border border-white/10 bg-slate-800/30 px-4 py-2 text-sm text-slate-200"
            >
              <span>{a.metric}</span>
              <span className="text-amber-200">{a.threshold}</span>
              <span className="text-slate-400">{a.deviation}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
