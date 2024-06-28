"use client";

import { useEffect } from "react";
import * as Sentry from "@sentry/nextjs";
import type { UserInstrumentation } from "@/app/server/user";

interface InstrumentUserClientSideProps {
  user: UserInstrumentation;
  children: React.ReactNode;
}

export function InstrumentUserClientSide({
  user,
  children,
}: InstrumentUserClientSideProps) {
  useEffect(() => {
    Sentry.setUser(user);
  }, [user.id]);
  return <>{children}</>;
}
