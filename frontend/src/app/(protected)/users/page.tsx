"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { createColumnHelper, tableFeatures, useTable } from "@tanstack/react-table";

import { api, ApiError, type UserRecord } from "@/lib/api";

import { useSession } from "../session-provider";
import { UserFormModal } from "./user-form-modal";
import { DeleteConfirmDialog } from "./delete-confirm-dialog";

const features = tableFeatures({});
const columnHelper = createColumnHelper<typeof features, UserRecord>();
const PAGE_SIZE = 20;
const EMPTY_USERS: UserRecord[] = [];

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

export default function UsersPage() {
  const session = useSession();
  const isAdmin = session.role === "admin";
  const router = useRouter();

  const [users, setUsers] = useState<UserRecord[]>(EMPTY_USERS);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [loggingOut, setLoggingOut] = useState(false);

  const [creating, setCreating] = useState(false);
  const [editingUser, setEditingUser] = useState<UserRecord | null>(null);
  const [deletingUser, setDeletingUser] = useState<UserRecord | null>(null);

  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const fetchUsers = useCallback(async (targetPage: number) => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.listUsers(targetPage, PAGE_SIZE);
      setUsers(res.users);
      setTotal(res.total);
      setPage(res.page);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to load users.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    // Initial data fetch on mount — loading already starts true, so this is
    // the standard fetch-on-mount pattern, not a cascading-render chain.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchUsers(1);
  }, [fetchUsers]);

  function handleMutationSuccess() {
    const wasLastRowOnPage = deletingUser !== null && users.length === 1 && page > 1;
    setCreating(false);
    setEditingUser(null);
    setDeletingUser(null);
    fetchUsers(wasLastRowOnPage ? page - 1 : page);
  }

  async function handleLogout() {
    setLoggingOut(true);
    await api.logout();
    router.push("/login");
    router.refresh();
  }

  const columns = useMemo(
    () =>
      columnHelper.columns([
        columnHelper.accessor("name", { header: "Name" }),
        columnHelper.accessor("email", { header: "Email" }),
        columnHelper.accessor("created_at", {
          header: "Created",
          cell: (info) => formatDate(info.getValue()),
        }),
        ...(isAdmin
          ? [
              columnHelper.display({
                id: "actions",
                header: "Actions",
                cell: ({ row }) => (
                  <div className="flex gap-3">
                    <button
                      type="button"
                      onClick={() => setEditingUser(row.original)}
                      className="text-sm font-medium text-gray-700 hover:underline dark:text-gray-300"
                    >
                      Edit
                    </button>
                    <button
                      type="button"
                      onClick={() => setDeletingUser(row.original)}
                      className="text-sm font-medium text-red-600 hover:underline"
                    >
                      Delete
                    </button>
                  </div>
                ),
              }),
            ]
          : []),
      ]),
    [isAdmin],
  );

  const table = useTable({ features, columns, data: users });

  return (
    <main className="mx-auto max-w-4xl px-4 py-10">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-gray-900 dark:text-gray-100">Users</h1>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Signed in as {session.name} ({session.email}) · {session.role}
          </p>
        </div>
        <div className="flex items-center gap-2">
          {isAdmin && (
            <button
              type="button"
              onClick={() => setCreating(true)}
              className="rounded-md bg-gray-900 px-3 py-2 text-sm font-medium text-white hover:bg-gray-800 dark:bg-gray-100 dark:text-gray-900 dark:hover:bg-white"
            >
              New user
            </button>
          )}
          <button
            type="button"
            onClick={handleLogout}
            disabled={loggingOut}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-60 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
          >
            {loggingOut ? "Signing out…" : "Log out"}
          </button>
        </div>
      </div>

      {error && (
        <div className="mt-6 flex items-center justify-between rounded-md bg-red-50 px-4 py-3 text-sm text-red-700 dark:bg-red-950 dark:text-red-400">
          <span>{error}</span>
          <button type="button" onClick={() => fetchUsers(page)} className="font-medium underline">
            Retry
          </button>
        </div>
      )}

      <div className="mt-6 overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-800">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-gray-900">
            {table.getHeaderGroups().map((group) => (
              <tr key={group.id}>
                {group.headers.map((header) => (
                  <th key={header.id} className="px-4 py-3 font-medium text-gray-600 dark:text-gray-400">
                    {header.isPlaceholder ? null : <table.FlexRender header={header} />}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={columns.length} className="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
                  Loading…
                </td>
              </tr>
            ) : users.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
                  No users yet.
                </td>
              </tr>
            ) : (
              table.getRowModel().rows.map((row) => (
                <tr key={row.id} className="border-b border-gray-100 last:border-0 dark:border-gray-800">
                  {row.getAllCells().map((cell) => (
                    <td key={cell.id} className="px-4 py-3 text-gray-900 dark:text-gray-100">
                      <table.FlexRender cell={cell} />
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <div className="mt-4 flex items-center justify-between text-sm text-gray-500 dark:text-gray-400">
        <span>
          Page {page} of {pageCount} · {total} total
        </span>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => fetchUsers(page - 1)}
            disabled={page <= 1 || loading}
            className="rounded-md border border-gray-300 px-3 py-1.5 disabled:opacity-40 dark:border-gray-700"
          >
            Previous
          </button>
          <button
            type="button"
            onClick={() => fetchUsers(page + 1)}
            disabled={page >= pageCount || loading}
            className="rounded-md border border-gray-300 px-3 py-1.5 disabled:opacity-40 dark:border-gray-700"
          >
            Next
          </button>
        </div>
      </div>

      {creating && (
        <UserFormModal mode="create" onClose={() => setCreating(false)} onSuccess={handleMutationSuccess} />
      )}
      {editingUser && (
        <UserFormModal
          mode="edit"
          user={editingUser}
          onClose={() => setEditingUser(null)}
          onSuccess={handleMutationSuccess}
        />
      )}
      {deletingUser && (
        <DeleteConfirmDialog
          user={deletingUser}
          onCancel={() => setDeletingUser(null)}
          onSuccess={handleMutationSuccess}
        />
      )}
    </main>
  );
}
