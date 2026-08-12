import { redirect } from "next/navigation";
import type { ReactNode } from "react";

import { getSession } from "@/lib/session";

import { SessionProvider } from "./session-provider";

// Runs on the server for every route under this group: no session, no
// render — the redirect happens before any protected UI or data ever
// reaches the client.
export default async function ProtectedLayout({ children }: { children: ReactNode }) {
  const session = await getSession();
  if (!session) {
    redirect("/login");
  }

  return <SessionProvider session={session}>{children}</SessionProvider>;
}
