import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { api, ApiError, type UserRecord } from "@/lib/api";

import { LogViewerModal } from "./log-viewer-modal";

jest.mock("@/lib/api", () => ({
  ...jest.requireActual("@/lib/api"),
  api: { listUserLogs: jest.fn() },
}));

const mockedApi = jest.mocked(api);

const user: UserRecord = {
  id: "u1",
  name: "Ada Lovelace",
  email: "ada@example.com",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

beforeEach(() => {
  jest.clearAllMocks();
});

test("shows a loading state before the fetch resolves", () => {
  mockedApi.listUserLogs.mockReturnValue(new Promise(() => {}));
  render(<LogViewerModal user={user} onClose={jest.fn()} />);

  expect(screen.getByText("Loading…")).toBeInTheDocument();
});

test("fetches logs for the given user and renders them", async () => {
  mockedApi.listUserLogs.mockResolvedValue({
    logs: [
      { user_id: "u1", event: "user.created", data: { name: "Ada" }, created_at: "2026-01-01T00:00:00Z" },
      { user_id: "u1", event: "user.login", data: {}, created_at: "2026-01-02T00:00:00Z" },
    ],
    total: 2,
    page: 1,
    page_size: 20,
  });
  render(<LogViewerModal user={user} onClose={jest.fn()} />);

  expect(await screen.findByText("user.created")).toBeInTheDocument();
  expect(screen.getByText("user.login")).toBeInTheDocument();
  expect(mockedApi.listUserLogs).toHaveBeenCalledWith("u1", 1, 20);
});

test("shows an empty state when there is no log history", async () => {
  mockedApi.listUserLogs.mockResolvedValue({ logs: [], total: 0, page: 1, page_size: 20 });
  render(<LogViewerModal user={user} onClose={jest.fn()} />);

  expect(await screen.findByText("No log events yet.")).toBeInTheDocument();
});

test("shows an error banner with a working retry", async () => {
  mockedApi.listUserLogs.mockRejectedValueOnce(new ApiError(500, "internal error"));
  render(<LogViewerModal user={user} onClose={jest.fn()} />);

  expect(await screen.findByText("internal error")).toBeInTheDocument();

  mockedApi.listUserLogs.mockResolvedValueOnce({
    logs: [{ user_id: "u1", event: "user.login", data: {}, created_at: "2026-01-01T00:00:00Z" }],
    total: 1,
    page: 1,
    page_size: 20,
  });
  const ue = userEvent.setup();
  await ue.click(screen.getByRole("button", { name: /retry/i }));

  expect(await screen.findByText("user.login")).toBeInTheDocument();
  expect(mockedApi.listUserLogs).toHaveBeenCalledTimes(2);
});

test("close button calls onClose", async () => {
  mockedApi.listUserLogs.mockResolvedValue({ logs: [], total: 0, page: 1, page_size: 20 });
  const onClose = jest.fn();
  const ue = userEvent.setup();
  render(<LogViewerModal user={user} onClose={onClose} />);

  await screen.findByText("No log events yet.");
  await ue.click(screen.getByRole("button", { name: /close/i }));

  expect(onClose).toHaveBeenCalled();
});

test("pagination fetches the next page", async () => {
  mockedApi.listUserLogs.mockResolvedValueOnce({
    logs: [{ user_id: "u1", event: "user.created", data: {}, created_at: "2026-01-01T00:00:00Z" }],
    total: 40,
    page: 1,
    page_size: 20,
  });
  const ue = userEvent.setup();
  render(<LogViewerModal user={user} onClose={jest.fn()} />);

  await screen.findByText("user.created");
  expect(screen.getByRole("button", { name: /previous/i })).toBeDisabled();

  mockedApi.listUserLogs.mockResolvedValueOnce({ logs: [], total: 40, page: 2, page_size: 20 });
  await ue.click(screen.getByRole("button", { name: /^next$/i }));

  await waitFor(() => expect(mockedApi.listUserLogs).toHaveBeenCalledWith("u1", 2, 20));
});
