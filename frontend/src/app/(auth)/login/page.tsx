"use client";

import { signIn } from "next-auth/react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useState, FormEvent, Suspense } from "react";
import { useSearchParams } from "next/navigation";

function SignupForm() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    if (password !== confirmPassword) {
      setError("Passwords do not match");
      setLoading(false);
      return;
    }

    try {
      const res = await fetch("/api/auth/signup", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, username, password }),
      });

      if (!res.ok) {
        const data = await res.json();
        setError(data.message || "Signup failed");
        setLoading(false);
        return;
      }

      // Auto sign in after successful signup
      const signInRes = await signIn("credentials", {
        redirect: false,
        email,
        password,
        callbackUrl: "/dashboard",
      });

      setLoading(false);

      if (signInRes?.error) {
        setError("Account created! Please sign in.");
        // Switch to login tab
        setTimeout(() => window.location.reload(), 2000);
      } else {
        router.push("/dashboard");
        router.refresh();
      }
    } catch (err) {
      setError("An error occurred during signup");
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <label className="text-sm text-slate-200" htmlFor="signup-email">
          Email
        </label>
        <input
          id="signup-email"
          name="email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="w-full rounded-xl border border-white/10 bg-slate-800/80 px-4 py-3 text-sm text-white outline-none ring-0 transition focus:border-indigo-400 focus:bg-slate-800"
          placeholder="your@email.com"
          autoComplete="email"
          required
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm text-slate-200" htmlFor="signup-username">
          Username
        </label>
        <input
          id="signup-username"
          name="username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          className="w-full rounded-xl border border-white/10 bg-slate-800/80 px-4 py-3 text-sm text-white outline-none ring-0 transition focus:border-indigo-400 focus:bg-slate-800"
          placeholder="Choose a username"
          autoComplete="username"
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm text-slate-200" htmlFor="signup-password">
          Password
        </label>
        <input
          id="signup-password"
          name="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="w-full rounded-xl border border-white/10 bg-slate-800/80 px-4 py-3 text-sm text-white outline-none ring-0 transition focus:border-indigo-400 focus:bg-slate-800"
          placeholder="••••••••"
          autoComplete="new-password"
          required
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm text-slate-200" htmlFor="confirm-password">
          Confirm Password
        </label>
        <input
          id="confirm-password"
          name="confirmPassword"
          type="password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          className="w-full rounded-xl border border-white/10 bg-slate-800/80 px-4 py-3 text-sm text-white outline-none ring-0 transition focus:border-indigo-400 focus:bg-slate-800"
          placeholder="••••••••"
          autoComplete="new-password"
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
        {loading ? "Creating account..." : "Sign up"}
      </button>
    </form>
  );
}

function LoginForm() {
  const router = useRouter();
  const params = useSearchParams();
  const callbackUrl = params.get("callbackUrl") || "/dashboard";
  const [email, setEmail] = useState("admin@smartmonitor.com");
  const [password, setPassword] = useState("adminpassword");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    const res = await signIn("credentials", {
      redirect: false,
      email,
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
        <label className="text-sm text-slate-200" htmlFor="email">
          Email
        </label>
        <input
          id="email"
          name="email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="w-full rounded-xl border border-white/10 bg-slate-800/80 px-4 py-3 text-sm text-white outline-none ring-0 transition focus:border-indigo-400 focus:bg-slate-800"
          placeholder="admin@smartmonitor.com"
          autoComplete="email"
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
  const [activeTab, setActiveTab] = useState<"login" | "signup">("login");

  return (
    <div className="flex min-h-screen items-center justify-center px-4 py-10 text-slate-50">
      <div className="w-full max-w-lg space-y-8 rounded-3xl border border-white/10 bg-slate-900/70 p-10 shadow-2xl backdrop-blur-lg">
        <div className="space-y-2 text-center">
          <p className="text-xs uppercase tracking-[0.25em] text-indigo-300">Smart Monitor</p>
          <h1 className="text-3xl font-semibold text-white">
            {activeTab === "login" ? "Sign in" : "Sign up"}
          </h1>
          <p className="text-sm text-slate-300">
            Access centralized agent management, monitoring, and protection.
          </p>
        </div>

        {/* Tabs */}
        <div className="flex gap-2 rounded-xl bg-slate-800/50 p-1">
          <button
            onClick={() => setActiveTab("login")}
            className={`flex-1 rounded-lg px-4 py-2 text-sm font-medium transition ${
              activeTab === "login"
                ? "bg-indigo-500 text-white shadow-lg shadow-indigo-600/30"
                : "text-slate-300 hover:text-white"
            }`}
          >
            Sign in
          </button>
          <button
            onClick={() => setActiveTab("signup")}
            className={`flex-1 rounded-lg px-4 py-2 text-sm font-medium transition ${
              activeTab === "signup"
                ? "bg-indigo-500 text-white shadow-lg shadow-indigo-600/30"
                : "text-slate-300 hover:text-white"
            }`}
          >
            Sign up
          </button>
        </div>

        <Suspense fallback={<div className="text-center text-slate-300">Loading...</div>}>
          {activeTab === "login" ? <LoginForm /> : <SignupForm />}
        </Suspense>

        <div className="flex items-center justify-between text-xs text-slate-300">
          <span>
            {activeTab === "login"
              ? "Use ADMIN credentials configured via env."
              : "Create a new account to get started."}
          </span>
          <Link href="/" className="text-indigo-300 hover:text-indigo-200">
            Back home
          </Link>
        </div>
      </div>
    </div>
  );
}
