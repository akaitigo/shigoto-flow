import { describe, it, expect } from "vitest";
import { BACKEND_URL, oauthConnectUrl } from "../backend";

describe("oauthConnectUrl", () => {
  it("builds an absolute backend URL for each provider", () => {
    expect(oauthConnectUrl("google")).toBe(`${BACKEND_URL}/api/v1/auth/google`);
    expect(oauthConnectUrl("slack")).toBe(`${BACKEND_URL}/api/v1/auth/slack`);
    expect(oauthConnectUrl("github")).toBe(`${BACKEND_URL}/api/v1/auth/github`);
  });

  it("is absolute, not a relative path that would hit the frontend origin", () => {
    const url = oauthConnectUrl("google");
    expect(url.startsWith("http://") || url.startsWith("https://")).toBe(true);
  });

  it("defaults to localhost:8080 when NEXT_PUBLIC_BACKEND_URL is unset", () => {
    expect(BACKEND_URL).toBe("http://localhost:8080");
  });
});
