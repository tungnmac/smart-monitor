"use client";

import { signIn } from "next-auth/react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useState, FormEvent, Suspense } from "react";
import { useSearchParams } from "next/navigation";

function LoginForm() {
  const router = useRouter();
  const params = useSearchParams();
  const callbackUrl = params.get("callbackUrl") || "/dashboard";
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    const res = await signIn("credentials", {
      redirect: false,
      username,
      password,
      callbackUrl,
    });
    setLoading(false);
    if (res?.error) {
      setError("Invalid credentials");
      return;
    }
    router.push(callbackUrl);
    router.refresh();
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <label className="text-sm text-slate-200" htmlFor="username">
          Username
        </label>
        <input
          id="username"
          name="username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          className="w-full rounded-xl border border-white/10 bg-slate-800/80 px-4 py-3 text-sm text-white outline-none ring-0 transition focus:border-indigo-400 focus:bg-slate-800"
          placeholder="admin"
          autoComplete="username"
          required
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm text-slate-200" htmlFor="password">
          Password
        </label>
        <input
          id="password"
          name="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="w-full rounded-xl border border-white/10 bg-slate-800/80 px-4 py-3 text-sm text-white outline-none ring-0 transition focus:border-indigo-400 focus:bg-slate-800"
          placeholder="••••••••"
          autoComplete="current-password"
          required
        />
      </div>

      {error && (
        <div className="rounded-xl border border-rose-500/30 bg-rose-900/30 px-4 py-3 text-sm text-rose-100">
          {error}
        </div>
      )}

      <button
        type="submit"
        disabled={loading}
        className="w-full rounded-xl bg-indigo-500 px-4 py-3 text-sm font-semibold text-white shadow-lg shadow-indigo-600/30 transition hover:-translate-y-0.5 hover:bg-indigo-400 disabled:cursor-not-allowed disabled:opacity-70"
      >
        {loading ? "Signing in..." : "Sign in"}
      </button>
    </form>
  );
}

export default function LoginPage() {
  return (
    <div className="flex min-h-screen items-center justify-center px-4 py-10 text-slate-50">
      <div className="w-full max-w-lg space-y-8 rounded-3xl border border-white/10 bg-slate-900/70 p-10 shadow-2xl backdrop-blur-lg">
        <div className="space-y-2 text-center">
          <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Smart Monitor</p>
          <h1 className="text-3xl font-semibold text-white">Sign in</h1>
          <p className="text-sm text-slate-300">
            Access centralized agent management, monitoring, and protection.
          </p>
        </div>

        <Suspense fallback={<div className="text-center text-slate-300">Loading...</div>}>
          <LoginForm />
        </Suspense>

        <div className="flex items-center justify-between text-xs text-slate-300">
          <span>Use ADMIN credentials configured via env.</span>
          <Link href="/" className="text-indigo-300 hover:text-indigo-200">
            Back home
          </Link>
        </div>
      </div>
    </div>
  );
}
