"use client";

import { useAgents, useAgentControl, useProcesses } from "@/lib/hooks";
import { useState } from "react";

function ProcessList({ hostname }: { hostname: string }) {
  const [sortField, setSortField] = useState<'name' | 'cpu' | 'memory' | 'pid'>('cpu');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const { processes, loading, error, kill, killingPid } = useProcesses(hostname, sortField, sortOrder);

  if (loading && processes.length === 0) return <div className="p-4 text-xs text-slate-400">Loading processes...</div>;
  if (error) return <div className="p-4 text-xs text-rose-400">Error: {error}</div>;

  const sortedProcesses = processes; // Now sorted by backend

  const toggleSort = (field: 'name' | 'cpu' | 'memory' | 'pid') => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortOrder('desc');
    }
  };

  const SortIndicator = ({ field }: { field: string }) => {
    if (sortField !== field) return <span className="ml-1 opacity-10">⇅</span>;
    return <span className="ml-1 text-indigo-400 font-bold">{sortOrder === 'asc' ? '↑' : '↓'}</span>;
  };

  return (
    <div className="mt-2 overflow-hidden rounded-xl border border-white/5 bg-slate-800/30">
      <table className="w-full text-left text-xs text-slate-300">
        <thead className="bg-white/5 text-[10px] uppercase tracking-wider text-slate-400">
          <tr>
            <th className="px-3 py-2 text-center cursor-pointer hover:text-white transition-colors" onClick={() => toggleSort('pid')}>
              PID <SortIndicator field="pid" />
            </th>
            <th className="px-3 py-2 cursor-pointer hover:text-white transition-colors" onClick={() => toggleSort('name')}>
              Name & Command <SortIndicator field="name" />
            </th>
            <th className="px-3 py-2">Port</th>
            <th className="px-3 py-2 cursor-pointer hover:text-white transition-colors" onClick={() => toggleSort('cpu')}>
              CPU% <SortIndicator field="cpu" />
            </th>
            <th className="px-3 py-2 cursor-pointer hover:text-white transition-colors" onClick={() => toggleSort('memory')}>
              MEM% <SortIndicator field="memory" />
            </th>
            <th className="px-3 py-2 text-right">Actions</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-white/5">
          {sortedProcesses.map((proc) => {
            const isCritical = proc.cpu > 80 || proc.memory > 80;
            
            // Comprehensive list of protected OS and core application processes
            const protectedNames = [
              // Core OS (Linux/Unix/macOS)
              'kernel_task', 'launchd', 'systemd', 'init', 'kthreadd', 'kupdate', 'kblockd',
              'loginwindow', 'windowserver', 'cfprefsd', 'configd', 'distnoted', 'notifyd', 
              'syslogd', 'mdnsresponder', 'chronyd', 'dbus-daemon', 'udevd', 'syslog-ng',
              
              // Infrastructure & Databases
              'postgres', 'mysql', 'mysqld', 'redis-server', 'mongod', 'nginx', 'apache2', 'httpd',
              'docker', 'dockerd', 'containerd', 'kubelet', 'kube-proxy',
              
              // Smart Monitor Core
              'agent', 'backend', 'smart-monitor-backend', 'collector', 'procctl'
            ];

            const isProtected = 
              proc.pid <= 100 || // Usually reserved for kernel/system processes
              protectedNames.some(name => proc.name.toLowerCase().includes(name)) ||
              proc.command.toLowerCase().includes('/sbin/') ||
              proc.command.toLowerCase().includes('/usr/libexec/') ||
              proc.command.toLowerCase().includes('smart-monitor');

            return (
              <tr 
                key={proc.pid} 
                className={`hover:bg-white/5 transition-colors ${isCritical ? 'bg-rose-500/10' : ''}`}
              >
                <td className="px-3 py-2 text-center font-mono text-indigo-300">{proc.pid}</td>
                <td className="px-3 py-2">
                  <div className="font-medium text-slate-100 flex items-center gap-2">
                    {proc.name}
                    {isProtected && (
                      <span className="bg-indigo-500/20 text-indigo-300 text-[8px] px-1.5 py-0.5 rounded-full uppercase font-bold tracking-tighter">
                        Protected
                      </span>
                    )}
                    {isCritical && (
                      <span className="animate-pulse bg-rose-500 text-[8px] text-white px-1.5 py-0.5 rounded-full uppercase font-bold tracking-tighter">
                        Critical Alert
                      </span>
                    )}
                  </div>
                  <div className="text-[10px] text-slate-500 font-mono truncate max-w-[200px]" title={proc.command}>
                    {proc.command}
                  </div>
                </td>
                <td className="px-3 py-2">
                   {proc.port > 0 ? (
                     <span className="bg-indigo-500/20 text-indigo-200 px-2 py-0.5 rounded-md border border-indigo-500/30">
                       :{proc.port}
                     </span>
                   ) : (
                     <span className="text-slate-600">-</span>
                   )}
                </td>
                <td className={`px-3 py-2 font-semibold ${proc.cpu > 80 ? 'text-rose-400' : proc.cpu > 50 ? 'text-amber-400' : 'text-emerald-400'}`}>
                  {proc.cpu.toFixed(1)}%
                </td>
                <td className={`px-3 py-2 font-semibold ${proc.memory > 80 ? 'text-rose-400' : proc.memory > 50 ? 'text-amber-400' : 'text-emerald-400'}`}>
                  {proc.memory.toFixed(1)}%
                </td>
                <td className="px-3 py-2 text-right">
                  {!isProtected && (
                    <button
                      onClick={() => {
                        if (confirm(`❗ DANGER: Are you sure you want to terminate process "${proc.name}" (PID: ${proc.pid})?\n\nThis action cannot be undone.`)) {
                          kill(proc.pid);
                        }
                      }}
                      disabled={killingPid === proc.pid}
                      className="rounded bg-rose-500/10 px-2 py-1 text-[10px] font-bold text-rose-400 hover:bg-rose-500/20 disabled:opacity-50"
                    >
                      {killingPid === proc.pid ? "..." : "KILL"}
                    </button>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

export default function AgentsPage() {
  const { agents, loading, error } = useAgents();
  const { restart, block, unblock, loading: controlLoading } = useAgentControl();
  const [selectedAgent, setSelectedAgent] = useState<string | null>(null);
  const [selectedAction, setSelectedAction] = useState<string | null>(null);

  return (
    <div className="space-y-4">
      {/* ...existing code... */}
      <div className="rounded-3xl border border-white/10 bg-slate-900/60 p-6 shadow-xl">
        {/* ...existing code... */}
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
                <th className="px-4 py-3">CPU</th>
                <th className="px-4 py-3">RAM</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {agents.map((agent) => (
                <>
                  <tr key={agent.id} className="hover:bg-white/5">
                    <td className="px-4 py-3 font-semibold text-white">
                      <button 
                        onClick={() => setSelectedAgent(selectedAgent === agent.id ? null : agent.id)}
                        className="text-indigo-400 hover:text-indigo-300 hover:underline"
                      >
                        {agent.id}
                      </button>
                    </td>
                    <td className="px-4 py-3">{agent.host}</td>
                    <td className="px-4 py-3 uppercase text-indigo-200 text-xs">{agent.env}</td>
                    <td className="px-4 py-3 text-amber-200">{agent.cpu?.toFixed(1) || "0.0"}%</td>
                    <td className="px-4 py-3 text-amber-200">{agent.ram?.toFixed(1) || "0.0"}%</td>
                    <td className="px-4 py-3">
                      <span
                        className={`rounded-full px-3 py-1 text-[10px] font-semibold ${
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
                  {selectedAgent === agent.id && (
                    <tr>
                      <td colSpan={7} className="px-4 pb-4 pt-0">
                        <div className="rounded-2xl bg-slate-900/40 p-1">
                          <ProcessList hostname={agent.id} />
                        </div>
                      </td>
                    </tr>
                  )}
                </>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
