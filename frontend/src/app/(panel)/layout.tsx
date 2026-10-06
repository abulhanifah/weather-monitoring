"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { clearToken, getToken } from "@/lib/api";

const MENU = [
  { label: "Dashboard", href: "/dashboard" },
  { label: "Manajemen Device", href: "/devices" },
  { label: "Manajemen Sensor", href: "/sensors" },
  { label: "Manajemen User", href: "/users" },
];

export default function PanelLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const [ready, setReady] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
    } else {
      setReady(true);
    }
  }, [router]);

  function handleLogout() {
    clearToken();
    router.push("/login");
  }

  if (!ready) return null;

  return (
    <div className="flex min-h-screen bg-zinc-100 dark:bg-black">
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-30 bg-black/50 md:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}
      <aside
        className={`fixed inset-y-0 left-0 z-40 flex w-60 shrink-0 transform flex-col bg-zinc-900 text-zinc-100 transition-transform md:static md:translate-x-0 ${
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div className="flex items-center justify-between px-5 py-6">
          <span className="text-lg font-semibold">Weather Panel</span>
          <button
            onClick={() => setSidebarOpen(false)}
            aria-label="Tutup menu"
            className="rounded px-2 py-1 text-xl leading-none hover:bg-zinc-800 md:hidden"
          >
            ×
          </button>
        </div>
        <nav className="flex flex-1 flex-col gap-1 overflow-y-auto px-3">
          {MENU.map((item) => {
            const active = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={() => setSidebarOpen(false)}
                className={`rounded px-3 py-2 text-sm ${
                  active ? "bg-zinc-700 font-medium" : "hover:bg-zinc-800"
                }`}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>
        <div className="border-t border-zinc-700 p-3">
          <button
            onClick={handleLogout}
            className="w-full rounded px-3 py-2 text-left text-sm hover:bg-zinc-800"
          >
            Logout
          </button>
        </div>
      </aside>
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-20 flex items-center gap-3 bg-white px-4 py-3 shadow-sm dark:bg-zinc-900 md:hidden">
          <button
            onClick={() => setSidebarOpen(true)}
            aria-label="Buka menu"
            className="rounded px-2 py-1 text-xl leading-none hover:bg-zinc-200 dark:hover:bg-zinc-800"
          >
            ☰
          </button>
          <span className="font-semibold">Weather Panel</span>
        </header>
        <main className="min-w-0 flex-1 p-4 md:p-6">{children}</main>
      </div>
    </div>
  );
}
