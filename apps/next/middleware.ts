import { clerkMiddleware, createRouteMatcher } from "@clerk/nextjs/server";

// For more info, see:
// https://clerk.com/docs/upgrade-guides/core-2/nextjs#migrating-to-clerk-middleware
const isPublicRoute = createRouteMatcher([
  // Allow signed out users to access these routes:
  "/",
  "/register",
  "/login",
  "/login-to-cli",
  "/api/doc",
  "__nextjs_original-stack-frame",
]);

export default clerkMiddleware((auth, request) => {
  if (!isPublicRoute(request)) {
    auth().protect();
  }
});

export const config = {
  matcher: [
    // Exclude files with a "." followed by an extension, which are typically static files.
    // Exclude files in the _next directory, which are Next.js internals.
    "/((?!.+\\.[\\w]+$|_next).*)",
    // Re-include any files in the api or trpc folders that might have an extension
    "/(api|trpc)(.*)",
  ],
};
