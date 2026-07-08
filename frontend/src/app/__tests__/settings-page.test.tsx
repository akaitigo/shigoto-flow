import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/api", () => ({
  listDataSources: vi.fn(),
  deleteDataSource: vi.fn(),
}));

import SettingsPage from "../settings/page";
import { listDataSources, deleteDataSource } from "@/lib/api";
import type { DataSource } from "@/types/report";

const mockSource: DataSource = {
  id: "1",
  user_id: "u1",
  provider: "google",
  expires_at: "2026-12-31T00:00:00Z",
  created_at: "2026-04-01T00:00:00Z",
  updated_at: "2026-04-01T00:00:00Z",
};

describe("SettingsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("fetches data sources on mount and shows connected state", async () => {
    vi.mocked(listDataSources).mockResolvedValue([mockSource]);

    render(<SettingsPage />);

    expect(listDataSources).toHaveBeenCalledTimes(1);
    expect(
      await screen.findByRole("button", { name: "切断" }),
    ).toBeInTheDocument();
    expect(await screen.findByText("接続済み")).toBeInTheDocument();
  });

  it("calls deleteDataSource and refetches when disconnect is clicked", async () => {
    vi.mocked(listDataSources).mockResolvedValue([mockSource]);
    vi.mocked(deleteDataSource).mockResolvedValue({ status: "deleted" });

    render(<SettingsPage />);

    const disconnect = await screen.findByRole("button", { name: "切断" });
    fireEvent.click(disconnect);

    await waitFor(() =>
      expect(deleteDataSource).toHaveBeenCalledWith("google"),
    );
    // one initial load + one refetch after disconnect
    await waitFor(() => expect(listDataSources).toHaveBeenCalledTimes(2));
  });

  it("surfaces an error when the fetch fails", async () => {
    vi.mocked(listDataSources).mockRejectedValue(new Error("認証が必要です"));

    render(<SettingsPage />);

    expect(await screen.findByRole("alert")).toHaveTextContent("認証が必要です");
  });
});
