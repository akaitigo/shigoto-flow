import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { ReportList } from "../report-list";
import type { Report } from "@/types/report";

const mockReports: Report[] = [
  {
    id: "1",
    user_id: "user1",
    type: "daily",
    template_id: "tmpl1",
    content: "テスト日報の内容です。今日はミーティングに参加しました。",
    date: "2026-04-04",
    status: "draft",
    created_at: "2026-04-04T09:00:00Z",
    updated_at: "2026-04-04T09:00:00Z",
  },
  {
    id: "2",
    user_id: "user1",
    type: "daily",
    template_id: "tmpl1",
    content: "別の日報内容",
    date: "2026-04-03",
    status: "sent",
    created_at: "2026-04-03T09:00:00Z",
    updated_at: "2026-04-03T18:00:00Z",
  },
];

describe("ReportList", () => {
  it("renders report items", () => {
    render(<ReportList reports={mockReports} onSelect={() => {}} />);

    expect(screen.getByText("2026-04-04")).toBeInTheDocument();
    expect(screen.getByText("2026-04-03")).toBeInTheDocument();
  });

  it("shows empty state when no reports", () => {
    render(<ReportList reports={[]} onSelect={() => {}} />);

    expect(screen.getByText("レポートがまだありません")).toBeInTheDocument();
  });

  it("calls onSelect when clicking a report", () => {
    const onSelect = vi.fn();
    render(<ReportList reports={mockReports} onSelect={onSelect} />);

    fireEvent.click(screen.getByText("2026-04-04"));

    expect(onSelect).toHaveBeenCalledWith(mockReports[0]);
  });

  it("highlights selected report", () => {
    render(
      <ReportList reports={mockReports} onSelect={() => {}} selectedId="1" />,
    );

    const selectedButton = screen.getByText("2026-04-04").closest("button");
    expect(selectedButton?.className).toContain("border-blue-500");
  });

  it("shows status badges", () => {
    render(<ReportList reports={mockReports} onSelect={() => {}} />);

    expect(screen.getByText("下書き")).toBeInTheDocument();
    expect(screen.getByText("送信済み")).toBeInTheDocument();
  });
});
