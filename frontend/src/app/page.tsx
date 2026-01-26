import Link from "next/link";
import { redirect } from "next/navigation";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/auth";

export default async function Home() {
  const session = await getServerSession(authOptions);
  if (session) {
    redirect("/dashboard");
  }

  return (
    <main className="mx-auto flex min-h-screen max-w-5xl flex-col items-center justify-center px-6 py-12 text-slate-50">
      <div className="w-full rounded-3xl border border-white/10 bg-slate-900/60 p-10 shadow-2xl backdrop-blur-md">
        <div className="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
          <div className="max-w-2xl space-y-3">
            <p className="text-sm uppercase tracking-[0.3em] text-indigo-300">Smart Monitor</p>
            <h1 className="text-4xl font-semibold leading-tight text-white md:text-5xl">
              Central command for agents, monitoring, analytics, and protection
            </h1>
            <p className="text-slate-300">
              Log in to manage agents, inspect live telemetry, detect anomalies, and orchestrate protective actions across your fleet.
            </p>
          </div>
          <div className="flex flex-col gap-3 text-sm text-slate-200">
            <Link
              href="/login"
              className="inline-flex items-center justify-center rounded-full bg-indigo-500 px-6 py-3 text-base font-semibold text-white shadow-lg shadow-indigo-600/30 transition hover:-translate-y-0.5 hover:bg-indigo-400"
            >
              Continue to console
            </Link>
            <div className="rounded-full bg-white/5 px-4 py-2 text-xs text-slate-300">
              Auth required • Role-based access • Secure sessions
            </div>
          </div>
        </div>
        <div className="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {["Agents", "Monitor", "Analytics", "Detect", "Protect", "Prevent"].map((item) => (
            <div
              key={item}
              className="glass rounded-2xl p-4 text-sm text-slate-200 transition hover:-translate-y-1 hover:border-indigo-400/50"
            >
              <p className="text-xs uppercase tracking-wide text-indigo-300">{item}</p>
              <p className="mt-2 text-slate-100">Deep visibility and controls tailored for {item.toLowerCase()}.</p>
            </div>
          ))}
        </div>
      </div>
    </main>
  );
}
