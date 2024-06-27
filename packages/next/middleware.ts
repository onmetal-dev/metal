import { clerkMiddleware, createRouteMatcher } from "@clerk/nextjs/server";

const isDashboardRoute = createRouteMatcher(["/dashboard(.*)"]);

// See https://clerk.com/docs/references/nextjs/clerk-middleware
// note: we do not protect API routes since we do that in the API logic itself
// protecting it via Clerk would result in redirects to the login page on 401 responses
export default clerkMiddleware((auth, request) => {
  if (isDashboardRoute(request)) {
    auth().protect();
  }
});

export const config = {
  matcher: [
    // Exclude files with a "." followed by an extension, which are typically static files.
    // Exclude files in the _next directory, which are Next.js internals.
    "/((?!.*\\..*|_next).*)",
    "/",
    // Re-include any files in the api or trpc folders that might have an extension
    "/(api|trpc)(.*)",
  ],
};
