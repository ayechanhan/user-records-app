import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { api, ApiError } from "@/lib/api";

import LoginPage from "./page";

const pushMock = jest.fn();
const refreshMock = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock, refresh: refreshMock }),
}));

// Keep the real ApiError class (so `instanceof` and .message work as in
// production) and only replace the api object's methods.
jest.mock("@/lib/api", () => ({
  ...jest.requireActual("@/lib/api"),
  api: { login: jest.fn() },
}));

const mockedApi = jest.mocked(api);

beforeEach(() => {
  jest.clearAllMocks();
});

async function fillAndSubmit(email: string, password: string) {
  const user = userEvent.setup();
  await user.type(screen.getByLabelText("Email"), email);
  await user.type(screen.getByLabelText("Password"), password);
  await user.click(screen.getByRole("button", { name: /sign in/i }));
}

test("renders email and password fields", () => {
  render(<LoginPage />);
  expect(screen.getByLabelText("Email")).toBeInTheDocument();
  expect(screen.getByLabelText("Password")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: /sign in/i })).toBeInTheDocument();
});

test("submits credentials and redirects to /users on success", async () => {
  mockedApi.login.mockResolvedValue({ id: "admin", name: "Admin", email: "admin@example.com", role: "admin" });
  render(<LoginPage />);

  await fillAndSubmit("admin@example.com", "admin-password");

  await waitFor(() => expect(mockedApi.login).toHaveBeenCalledWith("admin@example.com", "admin-password"));
  expect(pushMock).toHaveBeenCalledWith("/users");
  expect(refreshMock).toHaveBeenCalled();
});

test("shows the API error message on invalid credentials", async () => {
  mockedApi.login.mockRejectedValue(new ApiError(401, "invalid credentials"));
  render(<LoginPage />);

  await fillAndSubmit("admin@example.com", "wrong-password");

  expect(await screen.findByRole("alert")).toHaveTextContent("invalid credentials");
  expect(pushMock).not.toHaveBeenCalled();
});

test("shows a generic message for a non-API error", async () => {
  mockedApi.login.mockRejectedValue(new Error("network down"));
  render(<LoginPage />);

  await fillAndSubmit("admin@example.com", "admin-password");

  expect(await screen.findByRole("alert")).toHaveTextContent("Something went wrong");
});

test("disables the submit button while the request is in flight", async () => {
  let resolveLogin: (value: Awaited<ReturnType<typeof api.login>>) => void;
  mockedApi.login.mockReturnValue(
    new Promise((resolve) => {
      resolveLogin = resolve;
    }),
  );
  render(<LoginPage />);

  await fillAndSubmit("admin@example.com", "admin-password");

  const button = screen.getByRole("button", { name: /signing in/i });
  expect(button).toBeDisabled();

  resolveLogin!({ id: "admin", name: "Admin", email: "admin@example.com", role: "admin" });
  await waitFor(() => expect(pushMock).toHaveBeenCalled());
});
