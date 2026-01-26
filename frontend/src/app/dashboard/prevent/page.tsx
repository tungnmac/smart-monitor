export default function PreventPage() {
  const standards = [
    {
      id: "STD-001",
      name: "Agent Version",
      requirement: "≥ 2.0.0",
      compliance: "94%",
      status: "compliant",
    },
    {
      id: "STD-002",
      name: "TLS Encryption",
      requirement: "enforce",
      compliance: "100%",
      status: "compliant",
    },
    {
      id: "STD-003",
      name: "Token Rotation",
      requirement: "90d max",
      compliance: "87%",
      status: "warning",
    },
    {
      id: "STD-004",
      name: "Resource Limits",
      requirement: "cpu ≤ 80%, ram ≤ 85%",
      compliance: "78%",
      status: "critical",
    },
  ];

  return (
    <div className="space-y-4">
      <div className="rounded-3xl border border-white/10 bg-slate-900/60 p-6 shadow-xl">
        <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Prevention</p>
        <h2 className="text-xl font-semibold text-white">Proactive hardening</h2>
        <p className="text-sm text-slate-300">Compliance checks, guardrails, and continuous posture verification.</p>

        <div className="mt-4 grid gap-3 md:grid-cols-4">
          {[
            { label: "Fleet health", val: "92.5%", icon: "🛡️" },
            { label: "Non-compliant", val: "8", icon: "⚠️" },
            { label: "Audit events", val: "1.2K/d", icon: "📋" },
            { label: "Posture score", val: "9.2/10", icon: "⭐" },
          ].map((m) => (
            <div
              key={m.label}
              className="rounded-2xl border border-white/10 bg-slate-800/40 p-3 text-center text-sm"
            >
              <div className="text-2xl">{m.icon}</div>
              <p className="mt-2 text-xs text-slate-400">{m.label}</p>
              <p className="text-xl font-semibold text-white">{m.val}</p>
            </div>
          ))}
        </div>

        <div className="mt-6 overflow-hidden rounded-2xl border border-white/10">
          <table className="w-full text-sm text-slate-200">
            <thead className="bg-white/5 text-xs uppercase tracking-wide text-slate-300">
              <tr>
                <th className="px-4 py-3">Standard</th>
                <th className="px-4 py-3">Requirement</th>
                <th className="px-4 py-3">Compliance</th>
                <th className="px-4 py-3">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {standards.map((std) => (
                <tr key={std.id} className="hover:bg-white/5">
                  <td className="px-4 py-3 font-semibold text-white">{std.name}</td>
                  <td className="px-4 py-3 font-mono text-xs text-indigo-200">{std.requirement}</td>
                  <td className="px-4 py-3">
                    <div className="w-32">
                      <div className="h-1.5 rounded-full bg-slate-700">
                        <div
                          className={`h-full rounded-full ${
                            Number(std.compliance) >= 90
                              ? "bg-emerald-500"
                              : Number(std.compliance) >= 80
                                ? "bg-amber-500"
                                : "bg-rose-500"
                          }`}
                          style={{ width: std.compliance }}
                        />
                      </div>
                      <p className="mt-1 text-xs text-slate-400">{std.compliance}</p>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`rounded-full px-3 py-1 text-xs font-semibold ${
                        std.status === "compliant"
                          ? "bg-emerald-500/20 text-emerald-100"
                          : std.status === "warning"
                            ? "bg-amber-500/20 text-amber-100"
                            : "bg-rose-500/20 text-rose-100"
                      }`}
                    >
                      {std.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="rounded-3xl border border-white/10 bg-slate-900/60 p-6 shadow-xl">
        <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Guardrails</p>
        <h2 className="text-xl font-semibold text-white">Continuous verification</h2>
        <div className="mt-4 space-y-2">
          {[
            { rule: "Only signed agent binaries allowed", freq: "per deployment" },
            { rule: "Auth tokens rotate every 90 days", freq: "automated" },
            { rule: "Metrics encrypted in transit (TLS)", freq: "always" },
            { rule: "Audit logs archived to immutable store", freq: "daily" },
          ].map((g) => (
            <div
              key={g.rule}
              className="rounded-lg border border-white/10 bg-slate-800/30 px-4 py-3 flex items-center justify-between text-sm text-slate-200"
            >
              <span className="flex-1">{g.rule}</span>
              <span className="text-xs text-slate-400">{g.freq}</span>
              <span className="ml-2 text-emerald-200">✓</span>
            </div>
          ))}
        </div>
      </div>

      <div className="rounded-3xl border border-white/10 bg-slate-900/60 p-6 shadow-xl">
        <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Audit & Events</p>
        <h2 className="text-xl font-semibold text-white">Change log</h2>
        <div className="mt-4 space-y-2">
          {[
            { evt: "Agent edge-01 registered", ts: "2 min ago", actor: "bootstrap" },
            { evt: "Policy POL-001 updated", ts: "1 hr ago", actor: "admin" },
            { evt: "Token key rotation", ts: "2 hrs ago", actor: "system" },
          ].map((e) => (
            <div
              key={e.evt}
              className="rounded-lg border border-white/10 bg-slate-800/30 px-4 py-2 text-sm text-slate-300"
            >
              <div className="flex items-center justify-between">
                <p className="text-slate-100">{e.evt}</p>
                <div className="flex gap-4 text-xs">
                  <span className="text-slate-400">{e.ts}</span>
                  <span className="rounded-full bg-white/10 px-2 py-0.5 text-slate-300">{e.actor}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
