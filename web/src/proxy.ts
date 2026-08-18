import { NextRequest, NextResponse } from "next/server";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

/** Coarse, edge-level route gating: redirect to /setup while unconfigured,
 * redirect away from /setup once configured, and require a session cookie
 * for everything else. Real authorization (role checks, expired tokens)
 * happens per-request against the Go API and is handled by the API client's
 * global 401 handling — this middleware only prevents obviously-wrong page
 * loads, it isn't the source of truth. */
export async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  let setupCompleted = true;
  try {
    const res = await fetch(`${API_BASE_URL}/api/setup/status`, { cache: "no-store" });
    if (res.ok) {
      const data = (await res.json()) as { setupCompleted?: boolean };
      setupCompleted = Boolean(data.setupCompleted);
    }
  } catch {
    // API unreachable — fail open so a briefly-down backend doesn't trap the
    // user in a redirect loop; the page-level API calls will surface the
    // real error instead.
  }

  if (!setupCompleted && pathname !== "/setup") {
    return NextResponse.redirect(new URL("/setup", request.url));
  }

  if (setupCompleted && pathname === "/setup") {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  const isPublicRoute = pathname === "/login" || pathname === "/setup";
  if (!isPublicRoute && !request.cookies.has("session")) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
