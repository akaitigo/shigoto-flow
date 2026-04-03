"use client";

import type { Report } from "@/types/report";

interface ReportListProps {
  reports: Report[];
  onSelect: (report: Report) => void;
  selectedId?: string;
}

export function ReportList({ reports, onSelect, selectedId }: ReportListProps) {
  if (reports.length === 0) {
    return (
      <div className="py-8 text-center text-gray-500">
        レポートがまだありません
      </div>
    );
  }

  return (
    <ul className="space-y-2">
      {reports.map((report) => (
        <li key={report.id}>
          <button
            onClick={() => onSelect(report)}
            className={`w-full rounded-lg border p-3 text-left transition-colors hover:bg-gray-50 ${
              selectedId === report.id
                ? "border-blue-500 bg-blue-50"
                : "border-gray-200"
            }`}
          >
            <div className="flex items-center justify-between">
              <span className="font-medium">{report.date}</span>
              <StatusBadge status={report.status} />
            </div>
            <p className="mt-1 truncate text-sm text-gray-500">
              {report.content.slice(0, 80)}
            </p>
          </button>
        </li>
      ))}
    </ul>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    draft: "bg-yellow-100 text-yellow-800",
    confirmed: "bg-blue-100 text-blue-800",
    sent: "bg-green-100 text-green-800",
  };

  const labels: Record<string, string> = {
    draft: "下書き",
    confirmed: "確認済み",
    sent: "送信済み",
  };

  return (
    <span
      className={`rounded-full px-2 py-0.5 text-xs ${colors[status] ?? "bg-gray-100 text-gray-800"}`}
    >
      {labels[status] ?? status}
    </span>
  );
}
