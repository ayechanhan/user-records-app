"use client";

import { createContext, useContext, type ReactNode } from "react";

import type { Session } from "@/lib/session";

const SessionContext = createContext<Session | null>(null);

export function SessionProvider({ session, children }: { session: Session; children: ReactNode }) {
  return <SessionContext.Provider value={session}>{children}</SessionContext.Provider>;
}

// useSession is only valid inside the (protected) layout, which guarantees a
// session exists before rendering children — see layout.tsx's redirect.
export function useSession(): Session {
  const session = useContext(SessionContext);
  if (!session) {
    throw new Error("useSession must be used within the (protected) route group");
  }
  return session;
}
