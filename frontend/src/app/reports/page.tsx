"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { ReportEditor } from "@/components/report-editor";
import { ReportList } from "@/components/report-list";
import { generateReport, listReports, updateReport } from "@/lib/api";
import type { Report, ReportType } from "@/types/report";

const tabs: { type: ReportType; label: string }[] = [
  { type: "daily", label: "日報" },
  { type: "weekly", label: "週報" },
  { type: "monthly", label: "月報" },
];

const generateLabels: Record<ReportType, string> = {
  daily: "日報を生成",
  weekly: "週報を生成",
  monthly: "月報を生成",
};

function today(): string {
  return new Date().toISOString().slice(0, 10);
}

export default function ReportsPage() {
  const [selectedType, setSelectedType] = useState<ReportType>("daily");
  const [reports, setReports] = useState<Report[]>([]);
  const [selected, setSelected] = useState<Report | null>(null);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadReports = useCallback(async (type: ReportType) => {
    setLoading(true);
    setError(null);
    try {
      const list = await listReports(type);
      setReports(list ?? []);
    } catch (e) {
      setError(e instanceof Error ? e.message : "レポートの取得に失敗しました");
      setReports([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    setSelected(null);
    void loadReports(selectedType);
  }, [selectedType, loadReports]);

  const handleGenerate = async () => {
    setGenerating(true);
    setError(null);
    try {
      const report = await generateReport({
        type: selectedType,
        date: today(),
      });
      await loadReports(selectedType);
      setSelected(report);
    } catch (e) {
      setError(e instanceof Error ? e.message : "レポートの生成に失敗しました");
    } finally {
      setGenerating(false);
    }
  };

  const handleSave = async (content: string, status: string) => {
    if (!selected) return;
    setError(null);
    try {
      await updateReport(selected.id, { content, status });
      await loadReports(selectedType);
    } catch (e) {
      setError(e instanceof Error ? e.message : "保存に失敗しました");
    }
  };

  const handleSend = async (content: string) => {
    if (!selected) return;
    setError(null);
    try {
      await updateReport(selected.id, { content, status: "sent" });
      await loadReports(selectedType);
    } catch (e) {
      setError(e instanceof Error ? e.message : "送信に失敗しました");
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="border-b bg-white">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <Link href="/" className="text-xl font-bold text-gray-900">
            Shigoto-Flow
          </Link>
          <nav className="flex gap-6">
            <Link href="/reports" className="font-medium text-blue-600">
              レポート
            </Link>
            <Link href="/settings" className="text-gray-600 hover:text-gray-900">
              設定
            </Link>
          </nav>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-8">
        <div className="mb-6 flex items-center justify-between">
          <div className="flex gap-2">
            {tabs.map((tab) => (
              <button
                key={tab.type}
                onClick={() => setSelectedType(tab.type)}
                className={`rounded-lg px-4 py-2 text-sm font-medium ${
                  selectedType === tab.type
                    ? "bg-blue-600 text-white"
                    : "bg-white text-gray-600 hover:bg-gray-100"
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
          <button
            onClick={handleGenerate}
            disabled={generating}
            className="rounded-lg bg-green-600 px-4 py-2 text-sm text-white hover:bg-green-700 disabled:opacity-50"
          >
            {generating ? "生成中..." : generateLabels[selectedType]}
          </button>
        </div>

        {error && (
          <div
            role="alert"
            className="mb-4 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700"
          >
            {error}
          </div>
        )}

        <div className="grid gap-6 md:grid-cols-[320px_1fr]">
          <div className="rounded-lg border border-gray-200 bg-white p-4">
            {loading ? (
              <p className="py-8 text-center text-gray-500">読み込み中...</p>
            ) : (
              <ReportList
                reports={reports}
                onSelect={setSelected}
                selectedId={selected?.id}
              />
            )}
          </div>

          <div className="rounded-lg border border-gray-200 bg-white p-6">
            {selected ? (
              <ReportEditor
                key={selected.id}
                report={selected}
                onSave={handleSave}
                onSend={handleSend}
              />
            ) : (
              <p className="py-8 text-center text-gray-500">
                レポートを選択するか、「{generateLabels[selectedType]}」で新しく作成してください
              </p>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
