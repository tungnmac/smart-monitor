"use client";

import { useAgents, useAgentControl } from "@/lib/hooks";
import { useState } from "react";

export default function AgentsPage() {
  const { agents, loading, error } = useAgents();
  const { restart, block, unblock, loading: controlLoading } = useAgentControl();
  const [selectedAction, setSelectedAction] = useState<string | null>(null);
  return (
    <div className="space-y-4">
      <div className="rounded-3xl border border-white/10 bg-slate-900/60 p-6 shadow-xl">
        <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
          <div>
            <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Agents</p>
            <h2 className="text-xl font-semibold text-white">Fleet directory</h2>
            <p className="text-sm text-slate-300">Registration, tokens, and process control.</p>
          </div>
          <div className="flex gap-2">
            <div className="rounded-full bg-white/10 px-3 py-2 text-xs text-slate-200">
              {loading ? "Loading..." : `${agents.length} agents`}
            </div>
          </div>
        </div>

        {error && (
          <div className="mt-4 rounded-xl border border-rose-500/30 bg-rose-900/30 px-4 py-2 text-xs text-rose-100">
            {error}
          </div>
        )}

        <div className="mt-4 overflow-hidden rounded-2xl border border-white/10">
          <table className="w-full text-left text-sm text-slate-200">
            <thead className="bg-white/5 text-xs uppercase tracking-wide text-slate-300">
              <tr>
                <th className="px-4 py-3">Agent</th>
                <th className="px-4 py-3">Host</th>
                <th className="px-4 py-3">Env</th>
                <th className="px-4 py-3">Region</th>
                <th className="px-4 py-3">CPU</th>
                <th className="px-4 py-3">RAM</th>
                <th className="px-4 py-3">Disk</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {agents.map((agent) => (
                <tr key={agent.id} className="hover:bg-white/5">
                  <td className="px-4 py-3 font-semibold text-white">{agent.id}</td>
                  <td className="px-4 py-3">{agent.host}</td>
                  <td className="px-4 py-3 uppercase text-indigo-200">{agent.env}</td>
                  <td className="px-4 py-3">{agent.region}</td>
                  <td className="px-4 py-3 text-amber-200">{agent.cpu?.toFixed(1) || "N/A"}%</td>
                  <td className="px-4 py-3 text-amber-200">{agent.ram?.toFixed(1) || "N/A"}%</td>
                  <td className="px-4 py-3 text-amber-200">{agent.disk?.toFixed(1) || "N/A"}%</td>
                  <td className="px-4 py-3">
                    <span
                      className={`rounded-full px-3 py-1 text-xs font-semibold ${
                        agent.status === "online"
                          ? "bg-emerald-500/20 text-emerald-100"
                          : agent.status === "degraded"
                            ? "bg-amber-500/20 text-amber-100"
                            : "bg-rose-500/20 text-rose-100"
                      }`}
                    >
                      {agent.status}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2 text-xs font-semibold">
                      <button
                        onClick={() => {
                          restart(agent.id);
                          setSelectedAction(agent.id);
                        }}
                        disabled={controlLoading}
                        className="rounded-full bg-white/10 px-3 py-1 text-slate-100 hover:bg-white/20 disabled:opacity-50"
                      >
                        {selectedAction === agent.id && controlLoading ? "..." : "Restart"}
                      </button>
                      <button
                        onClick={() => block(agent.id)}
                        disabled={controlLoading}
                        className="rounded-full bg-rose-500/20 px-3 py-1 text-rose-100 hover:bg-rose-500/30 disabled:opacity-50"
                      >
                        Block
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
