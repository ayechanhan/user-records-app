"use client";

import { useCallback, useEffect, useState } from "react";

import { api, ApiError, type UserLog, type UserRecord } from "@/lib/api";

type Props = {
  user: UserRecord;
  onClose: () => void;
};

const PAGE_SIZE = 20;

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString();
}

export function LogViewerModal({ user, onClose }: Props) {
  const [logs, setLogs] = useState<UserLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const fetchLogs = useCallback(
    async (targetPage: number) => {
      setLoading(true);
      setError(null);
      try {
        const res = await api.listUserLogs(user.id, targetPage, PAGE_SIZE);
        setLogs(res.logs);
        setTotal(res.total);
        setPage(res.page);
      } catch (err) {
        setError(err instanceof ApiError ? err.message : "Failed to load log history.");
      } finally {
        setLoading(false);
      }
    },
    [user.id],
  );

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchLogs(1);
  }, [fetchLogs]);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4">
      <div className="w-full max-w-2xl rounded-lg bg-white p-6 shadow-lg dark:bg-gray-900">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Log history</h2>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {user.name} ({user.email})
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-md border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
          >
            Close
          </button>
        </div>

        {error && (
          <div className="mt-4 flex items-center justify-between rounded-md bg-red-50 px-4 py-3 text-sm text-red-700 dark:bg-red-950 dark:text-red-400">
            <span>{error}</span>
            <button type="button" onClick={() => fetchLogs(page)} className="font-medium underline">
              Retry
            </button>
          </div>
        )}

        <div className="mt-4 max-h-96 overflow-y-auto rounded-lg border border-gray-200 dark:border-gray-800">
          {loading ? (
            <p className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">Loading…</p>
          ) : logs.length === 0 ? (
            <p className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No log events yet.</p>
          ) : (
            <ul className="divide-y divide-gray-100 dark:divide-gray-800">
              {logs.map((log, i) => (
                <li key={i} className="px-4 py-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-gray-900 dark:text-gray-100">{log.event}</span>
                    <span className="text-xs text-gray-500 dark:text-gray-400">{formatDateTime(log.created_at)}</span>
                  </div>
                  {log.data && Object.keys(log.data).length > 0 && (
                    <pre className="mt-1 overflow-x-auto rounded bg-gray-50 px-2 py-1 text-xs text-gray-600 dark:bg-gray-800 dark:text-gray-400">
                      {JSON.stringify(log.data, null, 2)}
                    </pre>
                  )}
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className="mt-4 flex items-center justify-between text-sm text-gray-500 dark:text-gray-400">
          <span>
            Page {page} of {pageCount} · {total} total
          </span>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => fetchLogs(page - 1)}
              disabled={page <= 1 || loading}
              className="rounded-md border border-gray-300 px-3 py-1.5 disabled:opacity-40 dark:border-gray-700"
            >
              Previous
            </button>
            <button
              type="button"
              onClick={() => fetchLogs(page + 1)}
              disabled={page >= pageCount || loading}
              className="rounded-md border border-gray-300 px-3 py-1.5 disabled:opacity-40 dark:border-gray-700"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
