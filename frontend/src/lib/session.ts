import "server-only";
import { cookies } from "next/headers";

import type { Identity } from "./api";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";
const AUTH_COOKIE_NAME = "auth_token";

export type Session = Identity;

// getSession runs on the server and forwards the incoming request's auth
// cookie to the backend explicitly — server-side fetch has no browser
// cookie jar, so credentials:"include" has no effect here.
export async function getSession(): Promise<Session | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get(AUTH_COOKIE_NAME)?.value;
  if (!token) {
    return null;
  }

  const res = await fetch(`${API_BASE_URL}/auth/me`, {
    headers: { Cookie: `${AUTH_COOKIE_NAME}=${token}` },
    cache: "no-store",
  });
  if (!res.ok) {
    return null;
  }
  return res.json();
}
