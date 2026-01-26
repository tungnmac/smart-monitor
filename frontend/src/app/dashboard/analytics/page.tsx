export default function AnalyticsPage() {
  return (
    <div className="space-y-4">
      <div className="rounded-3xl border border-white/10 bg-slate-900/60 p-6 shadow-xl">
        <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Analytics</p>
        <h2 className="text-xl font-semibold text-white">Insights and trends</h2>
        <p className="text-sm text-slate-300">Historical patterns, forecasts, and policy impact scoring.</p>

        <div className="mt-6 grid gap-4 md:grid-cols-2">
          <div className="rounded-2xl border border-white/10 bg-slate-800/40 p-4">
            <p className="text-xs uppercase tracking-wide text-slate-400">7-Day Trend</p>
            <div className="mt-3 h-24 rounded-lg border border-white/10 bg-slate-900/30 flex items-end justify-around gap-1">
              {[38, 42, 35, 48, 52, 45, 50].map((v, i) => (
                <div
                  key={i}
                  className="w-full max-w-12 rounded-t-lg bg-green-500/50"
                  style={{ height: `${v}%` }}
                />
              ))}
            </div>
            <p className="mt-2 text-sm text-slate-200">
              Avg <span className="font-semibold text-white">45.7%</span>
            </p>
          </div>

          <div className="rounded-2xl border border-white/10 bg-slate-800/40 p-4">
            <p className="text-xs uppercase tracking-wide text-slate-400">Anomaly Score</p>
            <div className="mt-3 space-y-2">
              {[
                { label: "Spike rate", score: 0.34, bar: 34 },
                { label: "Jitter", score: 0.21, bar: 21 },
                { label: "Drift", score: 0.12, bar: 12 },
              ].map((a) => (
                <div key={a.label} className="space-y-1">
                  <div className="flex justify-between text-xs text-slate-300">
                    <span>{a.label}</span>
                    <span className="text-indigo-200">{a.score}</span>
                  </div>
                  <div className="h-2 rounded-full bg-slate-700/50">
                    <div
                      className="h-full rounded-full bg-indigo-500"
                      style={{ width: `${a.bar}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="mt-6 rounded-2xl border border-white/10 bg-slate-800/40 p-4">
          <p className="mb-3 text-xs uppercase tracking-wide text-slate-400">Policy Efficacy (Last 30d)</p>
          <div className="space-y-2">
            {[
              { name: "Rate-limit CPU spikes", stopped: 12, prevented: 8, pct: "95%" },
              { name: "Isolate rogue processes", stopped: 3, prevented: 2, pct: "88%" },
              { name: "Enforce disk quota", stopped: 5, prevented: 4, pct: "92%" },
            ].map((p) => (
              <div
                key={p.name}
                className="flex items-center justify-between rounded-lg border border-white/10 bg-slate-900/30 px-3 py-2 text-sm"
              >
                <span className="text-slate-200">{p.name}</span>
                <div className="flex gap-3 text-xs">
                  <span className="text-emerald-200">Stopped: {p.stopped}</span>
                  <span className="text-amber-200">Prevented: {p.prevented}</span>
                  <span className="font-semibold text-indigo-200">{p.pct}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
