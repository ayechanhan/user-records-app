"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { api } from "@/lib/api";

import { useSession } from "../session-provider";

export default function DashboardPage() {
  const session = useSession();
  const router = useRouter();
  const [loggingOut, setLoggingOut] = useState(false);

  async function handleLogout() {
    setLoggingOut(true);
    await api.logout();
    router.push("/login");
    router.refresh();
  }

  return (
    <main className="mx-auto max-w-2xl px-4 py-10">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
            Welcome, {session.name}
          </h1>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {session.email} · {session.role}
          </p>
        </div>
        <button
          onClick={handleLogout}
          disabled={loggingOut}
          className="rounded-md border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-60 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
        >
          {loggingOut ? "Signing out…" : "Log out"}
        </button>
      </div>

      <p className="mt-8 text-sm text-gray-500 dark:text-gray-400">
        User management is coming in the next phase.
      </p>
    </main>
  );
}
