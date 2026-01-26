import Link from "next/link";

const streams = [
  { title: "Agent Heartbeats", value: "128 online", trend: "+2" },
  { title: "Monitors", value: "642 signals/min", trend: "stable" },
  { title: "Detections", value: "3 alerts", trend: "-1" },
];

const panels = [
  {
    title: "Agent Ops",
    body: "Manage enrollment, tokens, and process control from a single pane.",
    href: "/dashboard/agents",
  },
  {
    title: "Monitoring",
    body: "Live CPU/RAM/Disk telemetry with fleet-wide baselines and drift.",
    href: "/dashboard/monitor",
  },
  {
    title: "Analytics",
    body: "Trend lines, anomaly bands, and policy impact across environments.",
    href: "/dashboard/analytics",
  },
  {
    title: "Detect",
    body: "Surface spikes, rogue processes, and lateral movement patterns.",
    href: "/dashboard/detect",
  },
  {
    title: "Protect",
    body: "Apply policies to isolate, throttle, or terminate risky workloads.",
    href: "/dashboard/protect",
  },
  {
    title: "Prevent",
    body: "Proactive hardening with guardrails and continuous posture checks.",
    href: "/dashboard/prevent",
  },
];

export default function DashboardHome() {
  return (
    <div className="space-y-6">
      <div className="rounded-3xl border border-white/10 bg-gradient-to-br from-indigo-600/30 via-slate-900 to-slate-950 p-6 text-white shadow-2xl">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="text-xs uppercase tracking-[0.25em] text-indigo-200">Overview</p>
            <h2 className="text-2xl font-semibold">Unified fleet visibility</h2>
            <p className="text-sm text-slate-200">
              Real-time observability with built-in detect, protect, and prevent flows.
            </p>
          </div>
          <Link
            href="/dashboard/agents"
            className="rounded-full bg-white/15 px-4 py-2 text-sm font-semibold text-white transition hover:-translate-y-0.5 hover:bg-white/25"
          >
            Manage agents
          </Link>
        </div>
        <div className="mt-6 grid gap-4 sm:grid-cols-3">
          {streams.map((item) => (
            <div
              key={item.title}
              className="rounded-2xl border border-white/20 bg-slate-900/50 p-4 text-sm text-slate-100"
            >
              <p className="text-xs uppercase tracking-wide text-indigo-200">{item.title}</p>
              <p className="mt-2 text-2xl font-semibold text-white">{item.value}</p>
              <p className="text-xs text-emerald-200">{item.trend}</p>
            </div>
          ))}
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {panels.map((panel) => (
          <Link
            key={panel.title}
            href={panel.href}
            className="rounded-2xl border border-white/10 bg-slate-900/60 p-5 text-slate-100 shadow-lg transition hover:-translate-y-1 hover:border-indigo-400/50"
          >
            <p className="text-xs uppercase tracking-wide text-indigo-300">{panel.title}</p>
            <p className="mt-2 text-sm text-slate-200">{panel.body}</p>
          </Link>
        ))}
      </div>
    </div>
  );
}
