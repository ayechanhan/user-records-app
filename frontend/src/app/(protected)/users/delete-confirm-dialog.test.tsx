import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { api, ApiError, type UserRecord } from "@/lib/api";

import { DeleteConfirmDialog } from "./delete-confirm-dialog";

jest.mock("@/lib/api", () => ({
  ...jest.requireActual("@/lib/api"),
  api: { deleteUser: jest.fn() },
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

test("confirming calls deleteUser with the right id and calls onSuccess", async () => {
  mockedApi.deleteUser.mockResolvedValue(undefined);
  const onSuccess = jest.fn();
  const ue = userEvent.setup();
  render(<DeleteConfirmDialog user={user} onCancel={jest.fn()} onSuccess={onSuccess} />);

  await ue.click(screen.getByRole("button", { name: /^delete$/i }));

  await waitFor(() => expect(mockedApi.deleteUser).toHaveBeenCalledWith("u1"));
  expect(onSuccess).toHaveBeenCalled();
});

test("cancel calls onCancel without deleting", async () => {
  const onCancel = jest.fn();
  const ue = userEvent.setup();
  render(<DeleteConfirmDialog user={user} onCancel={onCancel} onSuccess={jest.fn()} />);

  await ue.click(screen.getByRole("button", { name: /cancel/i }));

  expect(onCancel).toHaveBeenCalled();
  expect(mockedApi.deleteUser).not.toHaveBeenCalled();
});

test("shows the API error message on failure and does not call onSuccess", async () => {
  mockedApi.deleteUser.mockRejectedValue(new ApiError(500, "internal error"));
  const onSuccess = jest.fn();
  const ue = userEvent.setup();
  render(<DeleteConfirmDialog user={user} onCancel={jest.fn()} onSuccess={onSuccess} />);

  await ue.click(screen.getByRole("button", { name: /^delete$/i }));

  expect(await screen.findByRole("alert")).toHaveTextContent("internal error");
  expect(onSuccess).not.toHaveBeenCalled();
});
