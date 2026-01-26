"use client";

import { usePolicies } from "@/lib/hooks";

export default function ProtectPage() {
  const { policies, loading, error } = usePolicies();

  return (
    <div className="space-y-4">
      <div className="rounded-3xl border border-white/10 bg-slate-900/60 p-6 shadow-xl">
        <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Protections</p>
        <h2 className="text-xl font-semibold text-white">Automated response policies</h2>
        <p className="text-sm text-slate-300">Apply policies to isolate, throttle, or terminate risky workloads • {loading ? "Loading..." : `${policies.length} policies`}</p>

        {error && (
          <div className="mt-3 rounded-lg border border-rose-500/30 bg-rose-900/30 px-3 py-2 text-xs text-rose-100">
            {error}
          </div>
        )}

        <div className="mt-4 grid gap-3 md:grid-cols-3">
          {[
            { label: "Active policies", val: policies.filter(p => p.enabled).length, color: "text-emerald-200" },
            { label: "Total applied to", val: policies.reduce((sum, p) => sum + (p.applied_agents?.length || 0), 0), color: "text-indigo-200" },
            { label: "Disabled", val: policies.filter(p => !p.enabled).length, color: "text-amber-200" },
          ].map((s) => (
            <div
              key={s.label}
              className="rounded-2xl border border-white/10 bg-slate-800/40 p-3 text-sm"
            >
              <p className="text-xs text-slate-400">{s.label}</p>
              <p className={`text-2xl font-semibold ${s.color}`}>{s.val}</p>
            </div>
          ))}
        </div>

        <div className="mt-6 overflow-hidden rounded-2xl border border-white/10">
          <table className="w-full text-sm text-slate-200">
            <thead className="bg-white/5 text-xs uppercase tracking-wide text-slate-300">
              <tr>
                <th className="px-4 py-3">Policy</th>
                <th className="px-4 py-3">Description</th>
                <th className="px-4 py-3">Actions</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Applied to</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {policies.map((p) => (
                <tr key={p.policy_id} className="hover:bg-white/5">
                  <td className="px-4 py-3 font-semibold text-white">{p.name}</td>
                  <td className="px-4 py-3 text-xs">{p.description}</td>
                  <td className="px-4 py-3 font-mono text-xs text-indigo-200">{p.actions.join(", ")}</td>
                  <td className="px-4 py-3">
                    <span
                      className={`rounded-full px-3 py-1 text-xs font-semibold ${
                        p.enabled
                          ? "bg-emerald-500/20 text-emerald-100"
                          : "bg-amber-500/20 text-amber-100"
                      }`}
                    >
                      {p.enabled ? "enabled" : "disabled"}
                    </span>
                  </td>
                  <td className="px-4 py-3">{p.applied_agents?.length || 0} agents</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="rounded-3xl border border-white/10 bg-slate-900/60 p-6 shadow-xl">
        <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Policy Templates</p>
        <h2 className="text-xl font-semibold text-white">Quick-start responses</h2>
        <div className="mt-4 grid gap-4 md:grid-cols-2">
          {[
            { tmpl: "Rate Limit CPU", desc: "Throttle to safe % when spike detected" },
            { tmpl: "Isolate Rogue", desc: "Terminate and quarantine unauthorized processes" },
            { tmpl: "Disk Quota", desc: "Cap partition usage and alert operators" },
            { tmpl: "Network Segment", desc: "Restrict traffic to trusted subnets only" },
          ].map((t) => (
            <button
              key={t.tmpl}
              className="rounded-xl border border-white/10 bg-slate-800/40 px-4 py-3 text-left text-sm text-slate-200 transition hover:-translate-y-0.5 hover:border-indigo-400/50"
            >
              <p className="font-semibold text-white">{t.tmpl}</p>
              <p className="text-xs text-slate-400">{t.desc}</p>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
