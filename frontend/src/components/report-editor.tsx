"use client";

import { useState, useCallback } from "react";
import type { Report } from "@/types/report";

interface ReportEditorProps {
  report: Report;
  onSave: (content: string, status: string) => void;
  onSend: (content: string) => void;
}

export function ReportEditor({ report, onSave, onSend }: ReportEditorProps) {
  const [content, setContent] = useState(report.content);
  const [isSaving, setIsSaving] = useState(false);

  const handleSave = useCallback(async () => {
    setIsSaving(true);
    try {
      onSave(content, "confirmed");
    } finally {
      setIsSaving(false);
    }
  }, [content, onSave]);

  const handleSend = useCallback(async () => {
    setIsSaving(true);
    try {
      onSend(content);
    } finally {
      setIsSaving(false);
    }
  }, [content, onSend]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">
          {reportTypeLabel(report.type)} — {report.date}
        </h2>
        <span className="rounded-full bg-gray-200 px-3 py-1 text-sm">
          {statusLabel(report.status)}
        </span>
      </div>

      <textarea
        className="h-96 w-full rounded-lg border border-gray-300 p-4 font-mono text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="レポート内容を入力..."
      />

      <div className="flex gap-3">
        <button
          onClick={handleSave}
          disabled={isSaving}
          className="rounded-lg bg-blue-600 px-6 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
        >
          {isSaving ? "保存中..." : "保存"}
        </button>
        <button
          onClick={handleSend}
          disabled={isSaving}
          className="rounded-lg bg-green-600 px-6 py-2 text-white hover:bg-green-700 disabled:opacity-50"
        >
          送信
        </button>
      </div>
    </div>
  );
}

function reportTypeLabel(type: string): string {
  switch (type) {
    case "daily":
      return "日報";
    case "weekly":
      return "週報";
    case "monthly":
      return "月報";
    default:
      return type;
  }
}

function statusLabel(status: string): string {
  switch (status) {
    case "draft":
      return "下書き";
    case "confirmed":
      return "確認済み";
    case "sent":
      return "送信済み";
    default:
      return status;
  }
}
