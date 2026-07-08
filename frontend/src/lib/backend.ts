// Base origin of the Go API server. Used for full-page OAuth redirects where a
// relative URL would incorrectly resolve against the frontend origin.
export const BACKEND_URL =
  process.env.NEXT_PUBLIC_BACKEND_URL ?? "http://localhost:8080";

// oauthConnectUrl builds the absolute URL that starts the OAuth flow for the
// given provider on the backend.
export function oauthConnectUrl(provider: string): string {
  return `${BACKEND_URL}/api/v1/auth/${provider}`;
}
