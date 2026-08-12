import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { api, ApiError, type UserRecord, type Identity } from "@/lib/api";

import { SessionProvider } from "../session-provider";
import UsersPage from "./page";

const pushMock = jest.fn();
const refreshMock = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock, refresh: refreshMock }),
}));

jest.mock("@/lib/api", () => ({
  ...jest.requireActual("@/lib/api"),
  api: {
    listUsers: jest.fn(),
    createUser: jest.fn(),
    updateUser: jest.fn(),
    deleteUser: jest.fn(),
    listUserLogs: jest.fn(),
    logout: jest.fn(),
  },
}));

const mockedApi = jest.mocked(api);

const adminSession: Identity = { id: "admin", name: "Admin", email: "admin@example.com", role: "admin" };
const userSession: Identity = { id: "u1", name: "Grace", email: "grace@example.com", role: "user" };

const users: UserRecord[] = [
  { id: "1", name: "Ada Lovelace", email: "ada@example.com", created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z" },
  { id: "2", name: "Grace Hopper", email: "grace@example.com", created_at: "2026-01-02T00:00:00Z", updated_at: "2026-01-02T00:00:00Z" },
];

function renderAs(session: Identity) {
  return render(
    <SessionProvider session={session}>
      <UsersPage />
    </SessionProvider>,
  );
}

beforeEach(() => {
  jest.clearAllMocks();
});

test("shows a loading state before the initial fetch resolves", async () => {
  mockedApi.listUsers.mockReturnValue(new Promise(() => {})); // never resolves
  renderAs(adminSession);

  expect(screen.getByText("Loading…")).toBeInTheDocument();
});

test("renders fetched rows for an admin, including the actions column", async () => {
  mockedApi.listUsers.mockResolvedValue({ users, total: 2, page: 1, page_size: 20 });
  renderAs(adminSession);

  expect(await screen.findByText("Ada Lovelace")).toBeInTheDocument();
  expect(screen.getByText("Grace Hopper")).toBeInTheDocument();
  expect(mockedApi.listUsers).toHaveBeenCalledWith(1, 20);

  const rows = screen.getAllByRole("row");
  expect(within(rows[1]).getByRole("button", { name: /edit/i })).toBeInTheDocument();
  expect(within(rows[1]).getByRole("button", { name: /delete/i })).toBeInTheDocument();
  expect(within(rows[1]).getByRole("button", { name: /logs/i })).toBeInTheDocument();
});

test("hides admin-only controls for a non-admin identity", async () => {
  mockedApi.listUsers.mockResolvedValue({ users, total: 2, page: 1, page_size: 20 });
  renderAs(userSession);

  await screen.findByText("Ada Lovelace");

  expect(screen.queryByRole("button", { name: /new user/i })).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: /edit/i })).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: /delete/i })).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: /logs/i })).not.toBeInTheDocument();
});

test("shows an empty state when there are no users", async () => {
  mockedApi.listUsers.mockResolvedValue({ users: [], total: 0, page: 1, page_size: 20 });
  renderAs(adminSession);

  expect(await screen.findByText("No users yet.")).toBeInTheDocument();
});

test("shows an error banner with a working retry on fetch failure", async () => {
  mockedApi.listUsers.mockRejectedValueOnce(new ApiError(500, "internal error"));
  renderAs(adminSession);

  expect(await screen.findByText("internal error")).toBeInTheDocument();

  mockedApi.listUsers.mockResolvedValueOnce({ users, total: 2, page: 1, page_size: 20 });
  const user = userEvent.setup();
  await user.click(screen.getByRole("button", { name: /retry/i }));

  expect(await screen.findByText("Ada Lovelace")).toBeInTheDocument();
  expect(mockedApi.listUsers).toHaveBeenCalledTimes(2);
});

test("admin can open the create modal via New user", async () => {
  mockedApi.listUsers.mockResolvedValue({ users, total: 2, page: 1, page_size: 20 });
  const user = userEvent.setup();
  renderAs(adminSession);

  await screen.findByText("Ada Lovelace");
  await user.click(screen.getByRole("button", { name: /new user/i }));

  expect(screen.getByRole("heading", { name: "New user" })).toBeInTheDocument();
});

test("admin can open the edit modal pre-filled for a row", async () => {
  mockedApi.listUsers.mockResolvedValue({ users, total: 2, page: 1, page_size: 20 });
  const user = userEvent.setup();
  renderAs(adminSession);

  await screen.findByText("Ada Lovelace");
  const rows = screen.getAllByRole("row");
  await user.click(within(rows[1]).getByRole("button", { name: /edit/i }));

  expect(screen.getByRole("heading", { name: "Edit user" })).toBeInTheDocument();
  expect(screen.getByLabelText("Name")).toHaveValue("Ada Lovelace");
});

test("admin can open the log viewer for a row", async () => {
  mockedApi.listUsers.mockResolvedValue({ users, total: 2, page: 1, page_size: 20 });
  mockedApi.listUserLogs.mockResolvedValue({
    logs: [{ user_id: "1", event: "user.created", data: {}, created_at: "2026-01-01T00:00:00Z" }],
    total: 1,
    page: 1,
    page_size: 20,
  });
  const user = userEvent.setup();
  renderAs(adminSession);

  await screen.findByText("Ada Lovelace");
  const rows = screen.getAllByRole("row");
  await user.click(within(rows[1]).getByRole("button", { name: /logs/i }));

  expect(screen.getByRole("heading", { name: "Log history" })).toBeInTheDocument();
  expect(await screen.findByText("user.created")).toBeInTheDocument();
  expect(mockedApi.listUserLogs).toHaveBeenCalledWith("1", 1, 20);
});

test("admin can open the delete confirmation for a row", async () => {
  mockedApi.listUsers.mockResolvedValue({ users, total: 2, page: 1, page_size: 20 });
  const user = userEvent.setup();
  renderAs(adminSession);

  await screen.findByText("Ada Lovelace");
  const rows = screen.getAllByRole("row");
  await user.click(within(rows[1]).getByRole("button", { name: /delete/i }));

  const dialogHeading = screen.getByRole("heading", { name: "Delete user" });
  expect(dialogHeading).toBeInTheDocument();
  expect(within(dialogHeading.parentElement!).getByText(/ada lovelace/i)).toBeInTheDocument();
});

test("pagination fetches the next page and disables Previous on page 1", async () => {
  mockedApi.listUsers.mockResolvedValueOnce({ users, total: 40, page: 1, page_size: 20 });
  const user = userEvent.setup();
  renderAs(adminSession);

  await screen.findByText("Ada Lovelace");
  expect(screen.getByRole("button", { name: /previous/i })).toBeDisabled();
  expect(screen.getByText(/page 1 of 2/i)).toBeInTheDocument();

  mockedApi.listUsers.mockResolvedValueOnce({ users: [], total: 40, page: 2, page_size: 20 });
  await user.click(screen.getByRole("button", { name: /^next$/i }));

  await waitFor(() => expect(mockedApi.listUsers).toHaveBeenCalledWith(2, 20));
});

test("logging out calls the API and redirects to /login", async () => {
  mockedApi.listUsers.mockResolvedValue({ users, total: 2, page: 1, page_size: 20 });
  mockedApi.logout.mockResolvedValue(undefined);
  const user = userEvent.setup();
  renderAs(adminSession);

  await screen.findByText("Ada Lovelace");
  await user.click(screen.getByRole("button", { name: /log out/i }));

  await waitFor(() => expect(mockedApi.logout).toHaveBeenCalled());
  expect(pushMock).toHaveBeenCalledWith("/login");
  expect(refreshMock).toHaveBeenCalled();
});
