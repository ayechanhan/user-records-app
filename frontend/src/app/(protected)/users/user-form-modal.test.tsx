import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { api, ApiError, type UserRecord } from "@/lib/api";

import { UserFormModal } from "./user-form-modal";

jest.mock("@/lib/api", () => ({
  ...jest.requireActual("@/lib/api"),
  api: { createUser: jest.fn(), updateUser: jest.fn() },
}));

const mockedApi = jest.mocked(api);

const existingUser: UserRecord = {
  id: "u1",
  name: "Ada Lovelace",
  email: "ada@example.com",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

beforeEach(() => {
  jest.clearAllMocks();
});

test("create mode renders empty fields and requires a password", () => {
  render(<UserFormModal mode="create" onClose={jest.fn()} onSuccess={jest.fn()} />);

  expect(screen.getByLabelText("Name")).toHaveValue("");
  expect(screen.getByLabelText("Email")).toHaveValue("");
  expect(screen.getByLabelText(/password/i)).toBeRequired();
});

test("create mode submits name/email/password and calls onSuccess", async () => {
  mockedApi.createUser.mockResolvedValue({ ...existingUser, id: "new-id" });
  const onSuccess = jest.fn();
  const user = userEvent.setup();
  render(<UserFormModal mode="create" onClose={jest.fn()} onSuccess={onSuccess} />);

  await user.type(screen.getByLabelText("Name"), "Grace Hopper");
  await user.type(screen.getByLabelText("Email"), "grace@example.com");
  await user.type(screen.getByLabelText(/password/i), "grace-password");
  await user.click(screen.getByRole("button", { name: /save/i }));

  await waitFor(() =>
    expect(mockedApi.createUser).toHaveBeenCalledWith({
      name: "Grace Hopper",
      email: "grace@example.com",
      password: "grace-password",
    }),
  );
  expect(onSuccess).toHaveBeenCalled();
});

test("create mode rejects a short password without calling the API", async () => {
  const user = userEvent.setup();
  render(<UserFormModal mode="create" onClose={jest.fn()} onSuccess={jest.fn()} />);

  await user.type(screen.getByLabelText("Name"), "Grace Hopper");
  await user.type(screen.getByLabelText("Email"), "grace@example.com");
  await user.type(screen.getByLabelText(/password/i), "short");
  await user.click(screen.getByRole("button", { name: /save/i }));

  expect(await screen.findByRole("alert")).toHaveTextContent("at least 8 characters");
  expect(mockedApi.createUser).not.toHaveBeenCalled();
});

test("edit mode pre-fills name and email, password optional", () => {
  render(<UserFormModal mode="edit" user={existingUser} onClose={jest.fn()} onSuccess={jest.fn()} />);

  expect(screen.getByLabelText("Name")).toHaveValue("Ada Lovelace");
  expect(screen.getByLabelText("Email")).toHaveValue("ada@example.com");
  expect(screen.getByLabelText(/password/i)).not.toBeRequired();
});

test("edit mode submits with password omitted when left blank", async () => {
  mockedApi.updateUser.mockResolvedValue(existingUser);
  const onSuccess = jest.fn();
  const user = userEvent.setup();
  render(<UserFormModal mode="edit" user={existingUser} onClose={jest.fn()} onSuccess={onSuccess} />);

  await user.clear(screen.getByLabelText("Name"));
  await user.type(screen.getByLabelText("Name"), "Ada L.");
  await user.click(screen.getByRole("button", { name: /save/i }));

  await waitFor(() =>
    expect(mockedApi.updateUser).toHaveBeenCalledWith("u1", {
      name: "Ada L.",
      email: "ada@example.com",
      password: undefined,
    }),
  );
  expect(onSuccess).toHaveBeenCalled();
});

test("edit mode submits a new password when provided", async () => {
  mockedApi.updateUser.mockResolvedValue(existingUser);
  const user = userEvent.setup();
  render(<UserFormModal mode="edit" user={existingUser} onClose={jest.fn()} onSuccess={jest.fn()} />);

  await user.type(screen.getByLabelText(/password/i), "new-password");
  await user.click(screen.getByRole("button", { name: /save/i }));

  await waitFor(() =>
    expect(mockedApi.updateUser).toHaveBeenCalledWith(
      "u1",
      expect.objectContaining({ password: "new-password" }),
    ),
  );
});

test("shows the API error message on failure", async () => {
  mockedApi.createUser.mockRejectedValue(new ApiError(409, "email already exists"));
  const user = userEvent.setup();
  render(<UserFormModal mode="create" onClose={jest.fn()} onSuccess={jest.fn()} />);

  await user.type(screen.getByLabelText("Name"), "Grace Hopper");
  await user.type(screen.getByLabelText("Email"), "grace@example.com");
  await user.type(screen.getByLabelText(/password/i), "grace-password");
  await user.click(screen.getByRole("button", { name: /save/i }));

  expect(await screen.findByRole("alert")).toHaveTextContent("email already exists");
});

test("cancel calls onClose without submitting", async () => {
  const onClose = jest.fn();
  const user = userEvent.setup();
  render(<UserFormModal mode="create" onClose={onClose} onSuccess={jest.fn()} />);

  await user.click(screen.getByRole("button", { name: /cancel/i }));

  expect(onClose).toHaveBeenCalled();
  expect(mockedApi.createUser).not.toHaveBeenCalled();
});
