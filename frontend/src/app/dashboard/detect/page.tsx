export default function DetectPage() {
  const incidents = [
    {
      id: "INC-001",
      host: "edge-01",
      type: "CPU Spike",
      severity: "critical",
      value: "98.5%",
      time: "2 min ago",
    },
    {
      id: "INC-002",
      host: "core-02",
      type: "Memory leak",
      severity: "high",
      value: "85% → 92%",
      time: "8 min ago",
    },
    {
      id: "INC-003",
      host: "db-01",
      type: "Disk pressure",
      severity: "medium",
      value: "79%",
      time: "1 hr ago",
    },
  ];

  return (
    <div className="space-y-4">
      <div className="rounded-3xl border border-white/10 bg-slate-900/60 p-6 shadow-xl">
        <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Detection</p>
        <h2 className="text-xl font-semibold text-white">Anomaly & threat alerts</h2>
        <p className="text-sm text-slate-300">Real-time surface of spikes, rogue processes, lateral moves.</p>

        <div className="mt-4 grid gap-3 md:grid-cols-3">
          {[
            { label: "Active alerts", val: "3", color: "text-rose-200" },
            { label: "Last 24h", val: "12", color: "text-amber-200" },
            { label: "False positives", val: "0.2%", color: "text-emerald-200" },
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

        <div className="mt-6 space-y-2">
          <h3 className="text-sm font-semibold text-slate-300">Recent Incidents</h3>
          {incidents.map((inc) => (
            <div
              key={inc.id}
              className={`rounded-xl border px-4 py-3 flex items-center justify-between text-sm ${
                inc.severity === "critical"
                  ? "border-rose-500/30 bg-rose-900/20"
                  : inc.severity === "high"
                    ? "border-amber-500/30 bg-amber-900/20"
                    : "border-yellow-500/30 bg-yellow-900/20"
              }`}
            >
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-slate-100">{inc.id}</span>
                  <span
                    className={`rounded-full px-2 py-0.5 text-xs font-semibold ${
                      inc.severity === "critical"
                        ? "bg-rose-500/20 text-rose-100"
                        : inc.severity === "high"
                          ? "bg-amber-500/20 text-amber-100"
                          : "bg-yellow-500/20 text-yellow-100"
                    }`}
                  >
                    {inc.severity}
                  </span>
                </div>
                <p className="text-xs text-slate-300">
                  {inc.host} — {inc.type}: <span className="text-white">{inc.value}</span>
                </p>
                <p className="text-xs text-slate-400">{inc.time}</p>
              </div>
              <button className="rounded-full bg-white/10 px-3 py-1 text-xs text-slate-100 hover:bg-white/20">
                View
              </button>
            </div>
          ))}
        </div>
      </div>

      <div className="rounded-3xl border border-white/10 bg-slate-900/60 p-6 shadow-xl">
        <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Baseline Rules</p>
        <h2 className="text-xl font-semibold text-white">Customizable detection thresholds</h2>
        <div className="mt-4 space-y-2">
          {[
            { metric: "CPU exceeded", rule: "cpu > 85% for 2min", action: "alert + log" },
            { metric: "Memory trend", rule: "mem ↑ >15% in 10min", action: "alert" },
            { metric: "Disk full", rule: "disk > 90%", action: "alert + protect" },
          ].map((r) => (
            <div
              key={r.metric}
              className="flex items-center justify-between rounded-lg border border-white/10 bg-slate-800/30 px-4 py-2 text-sm text-slate-200"
            >
              <span className="font-semibold">{r.metric}</span>
              <code className="text-xs text-indigo-200">{r.rule}</code>
              <span className="text-amber-200">{r.action}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
