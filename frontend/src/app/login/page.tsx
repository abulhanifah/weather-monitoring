"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { loginApi, setToken } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const data = await loginApi(username, password);
      setToken(data.token);
      router.push("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login gagal");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-100 dark:bg-black">
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-sm rounded-lg bg-white p-8 shadow dark:bg-zinc-900"
      >
        <h1 className="mb-6 text-center text-2xl font-semibold">Login</h1>
        {error && (
          <p className="mb-4 rounded bg-red-100 px-3 py-2 text-sm text-red-700">
            {error}
          </p>
        )}
        <label className="mb-4 block">
          <span className="mb-1 block text-sm font-medium">Username</span>
          <input
            type="email"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
            autoComplete="username"
            className="w-full rounded border border-zinc-300 px-3 py-2 dark:border-zinc-700 dark:bg-zinc-800"
          />
        </label>
        <label className="mb-6 block">
          <span className="mb-1 block text-sm font-medium">Password</span>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            autoComplete="current-password"
            className="w-full rounded border border-zinc-300 px-3 py-2 dark:border-zinc-700 dark:bg-zinc-800"
          />
        </label>
        <button
          type="submit"
          disabled={loading}
          className="w-full rounded bg-zinc-900 px-4 py-2 font-medium text-white disabled:opacity-50 dark:bg-zinc-50 dark:text-zinc-900"
        >
          {loading ? "Memproses..." : "Login"}
        </button>
      </form>
    </div>
  );
}
