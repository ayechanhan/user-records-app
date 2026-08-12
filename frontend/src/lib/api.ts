const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

export type Role = "admin" | "user";

export type Identity = {
  id: string;
  name: string;
  email: string;
  role: Role;
};

export type UserRecord = {
  id: string;
  name: string;
  email: string;
  created_at: string;
  updated_at: string;
};

export type UserListResponse = {
  users: UserRecord[];
  total: number;
  page: number;
  page_size: number;
};

export type CreateUserInput = {
  name: string;
  email: string;
  password: string;
};

export type UpdateUserInput = {
  name: string;
  email: string;
  password?: string;
};

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    credentials: "include",
    headers: { "Content-Type": "application/json", ...options.headers },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}) as { error?: string });
    throw new ApiError(res.status, body.error ?? `request failed with status ${res.status}`);
  }
  if (res.status === 204) {
    return undefined as T;
  }
  return res.json();
}

export const api = {
  login: (email: string, password: string) =>
    request<Identity>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  logout: () => request<void>("/auth/logout", { method: "POST" }),
  me: () => request<Identity>("/auth/me"),
  listUsers: (page: number, pageSize: number) =>
    request<UserListResponse>(`/users?page=${page}&page_size=${pageSize}`),
  createUser: (input: CreateUserInput) =>
    request<UserRecord>("/users", { method: "POST", body: JSON.stringify(input) }),
  updateUser: (id: string, input: UpdateUserInput) =>
    request<UserRecord>(`/users/${id}`, { method: "PUT", body: JSON.stringify(input) }),
  deleteUser: (id: string) => request<void>(`/users/${id}`, { method: "DELETE" }),
};
