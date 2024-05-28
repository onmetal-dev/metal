"use server";

import { db } from "@db/index";
import { waitlistedEmails } from "@db/schema";
import { ServerActionState } from "@lib/action";
import { headers } from "next/headers";
import { RateLimiterMemory } from "rate-limiter-flexible";
import { z } from "zod";

const rateLimiter = new RateLimiterMemory({
  points: 20, // 20 requests
  duration: 10, // per 10 seconds
});

export async function joinWaitlist(
  _prevState: ServerActionState,
  formData: FormData
): Promise<ServerActionState> {
  const email: string = formData.get("email") as string;
  const ip: string = headers().get("x-forwarded-for") || "unknown-ip";
  try {
    await rateLimiter.consume(ip, 1);
  } catch (e: any) {
    return {
      isError: true,
      message: `Too many requests. Please try again later.`,
    };
  }
  if (!email) {
    return { isError: true, message: "Email is required." };
  }
  try {
    z.string().email().parse(email);
  } catch (e: any) {
    return {
      isError: true,
      message: `Email is invalid: ${e.errors[0].message}`,
    };
  }

  // fire and forget this fetch to keep the response snappy
  fetch(process.env.LOOPS_WAITLIST_FORM_URL!, {
    method: "POST",
    body: `email=${encodeURIComponent(email)}&userGroup=waitlist`,
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
  });

  const result = await db
    .insert(waitlistedEmails)
    .values({ email })
    .returning()
    .onConflictDoNothing();
  if (result.length === 0) {
    return { isError: false, message: "You're already on the waitlist!" };
  }
  return { isError: false, message: "You've been added to the waitlist!" };
}
