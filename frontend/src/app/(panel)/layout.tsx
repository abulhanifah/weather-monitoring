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
      <aside className="flex w-60 shrink-0 flex-col bg-zinc-900 text-zinc-100">
        <div className="px-5 py-6 text-lg font-semibold">Weather Panel</div>
        <nav className="flex flex-1 flex-col gap-1 px-3">
          {MENU.map((item) => {
            const active = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
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
      <main className="flex-1 p-6">{children}</main>
    </div>
  );
}
