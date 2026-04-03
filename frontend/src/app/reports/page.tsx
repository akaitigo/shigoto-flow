"use client";

import { useState } from "react";
import Link from "next/link";
import type { ReportType } from "@/types/report";

export default function ReportsPage() {
  const [selectedType, setSelectedType] = useState<ReportType>("daily");

  const tabs: { type: ReportType; label: string }[] = [
    { type: "daily", label: "日報" },
    { type: "weekly", label: "週報" },
    { type: "monthly", label: "月報" },
  ];

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
          <button className="rounded-lg bg-green-600 px-4 py-2 text-sm text-white hover:bg-green-700">
            {selectedType === "daily"
              ? "日報を生成"
              : selectedType === "weekly"
                ? "週報を生成"
                : "月報を生成"}
          </button>
        </div>

        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <p className="text-center text-gray-500">
            データソースを接続してレポートを生成してください
          </p>
        </div>
      </main>
    </div>
  );
}
