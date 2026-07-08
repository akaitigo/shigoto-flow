import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/api", () => ({
  listReports: vi.fn(),
  generateReport: vi.fn(),
  updateReport: vi.fn(),
}));

import ReportsPage from "../reports/page";
import { listReports, generateReport } from "@/lib/api";
import type { Report } from "@/types/report";

const mockReport: Report = {
  id: "r1",
  user_id: "u1",
  type: "daily",
  template_id: "t1",
  content: "生成された日報の内容",
  date: "2026-04-05",
  status: "draft",
  created_at: "2026-04-05T09:00:00Z",
  updated_at: "2026-04-05T09:00:00Z",
};

describe("ReportsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(listReports).mockResolvedValue([]);
  });

  it("fetches daily reports on mount", async () => {
    render(<ReportsPage />);

    await waitFor(() => expect(listReports).toHaveBeenCalledWith("daily"));
  });

  it("invokes generateReport when the generate button is clicked", async () => {
    vi.mocked(generateReport).mockResolvedValue(mockReport);

    render(<ReportsPage />);
    await waitFor(() => expect(listReports).toHaveBeenCalled());

    fireEvent.click(screen.getByRole("button", { name: "日報を生成" }));

    await waitFor(() =>
      expect(generateReport).toHaveBeenCalledWith(
        expect.objectContaining({ type: "daily" }),
      ),
    );
  });

  it("refetches when switching report type", async () => {
    render(<ReportsPage />);
    await waitFor(() => expect(listReports).toHaveBeenCalledWith("daily"));

    fireEvent.click(screen.getByRole("button", { name: "週報" }));

    await waitFor(() => expect(listReports).toHaveBeenCalledWith("weekly"));
  });

  it("renders fetched reports in the list", async () => {
    vi.mocked(listReports).mockResolvedValue([mockReport]);

    render(<ReportsPage />);

    expect(await screen.findByText("2026-04-05")).toBeInTheDocument();
  });
});
