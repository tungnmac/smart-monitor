import Link from "next/link";
import { redirect } from "next/navigation";
import { getServerSession } from "next-auth";
import { authOptions } from "@/lib/auth";

const navItems = [
  { href: "/dashboard", label: "Overview" },
  { href: "/dashboard/agents", label: "Agents" },
  { href: "/dashboard/monitor", label: "Monitor" },
  { href: "/dashboard/analytics", label: "Analytics" },
  { href: "/dashboard/detect", label: "Detect" },
  { href: "/dashboard/protect", label: "Protect" },
  { href: "/dashboard/prevent", label: "Prevent" },
];

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const session = await getServerSession(authOptions);
  if (!session) {
    redirect("/login");
  }

  return (
    <div className="min-h-screen">
      <header className="border-b border-white/10 bg-slate-900/60 backdrop-blur-lg">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
          <div>
            <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Smart Monitor</p>
            <h1 className="text-lg font-semibold text-white">Unified Operations Console</h1>
          </div>
          <div className="flex items-center gap-3 text-sm text-slate-200">
            <div className="rounded-full bg-emerald-500/20 px-3 py-1 text-emerald-200">Secure</div>
            <div className="rounded-full bg-white/10 px-3 py-1">{session.user?.email}</div>
            <Link
              href="/api/auth/signout"
              className="rounded-full bg-white/10 px-3 py-1 text-slate-50 hover:bg-white/20"
            >
              Sign out
            </Link>
          </div>
        </div>
      </header>

      <div className="mx-auto flex max-w-7xl gap-6 px-6 py-8">
        <aside className="hidden w-64 space-y-3 md:block">
          {navItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="block rounded-xl border border-white/10 bg-slate-900/60 px-4 py-3 text-sm font-semibold text-slate-100 transition hover:-translate-y-0.5 hover:border-indigo-400/50 hover:text-white"
            >
              {item.label}
            </Link>
          ))}
        </aside>

        <main className="flex-1 space-y-6">
          <div className="grid gap-4 sm:grid-cols-3">
            {["Agents", "Health", "Security"].map((item, idx) => (
              <div
                key={item}
                className="rounded-2xl border border-white/10 bg-slate-900/60 p-4 text-sm text-slate-200 shadow-lg"
              >
                <p className="text-xs uppercase tracking-wide text-indigo-300">{item}</p>
                <p className="mt-2 text-2xl font-semibold text-white">
                  {idx === 0 ? "128" : idx === 1 ? "99.97%" : "24h clean"}
                </p>
                <p className="text-xs text-slate-400">Fleet status</p>
              </div>
            ))}
          </div>
          {children}
        </main>
      </div>
    </div>
  );
}
