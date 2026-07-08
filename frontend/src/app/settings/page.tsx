"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { DataSourceSettings } from "@/components/datasource-settings";
import { deleteDataSource, listDataSources } from "@/lib/api";
import { oauthConnectUrl } from "@/lib/backend";
import type { DataSource, Provider } from "@/types/report";

export default function SettingsPage() {
  const [dataSources, setDataSources] = useState<DataSource[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadDataSources = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const sources = await listDataSources();
      setDataSources(sources ?? []);
    } catch (e) {
      setError(
        e instanceof Error ? e.message : "データソースの取得に失敗しました",
      );
      setDataSources([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadDataSources();
  }, [loadDataSources]);

  const handleConnect = (provider: Provider) => {
    // Absolute URL to the backend origin — a relative URL would resolve against
    // the frontend origin (:3000) and 404.
    window.location.href = oauthConnectUrl(provider);
  };

  const handleDisconnect = async (provider: Provider) => {
    setError(null);
    try {
      await deleteDataSource(provider);
      await loadDataSources();
    } catch (e) {
      setError(e instanceof Error ? e.message : "切断に失敗しました");
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
            <Link href="/reports" className="text-gray-600 hover:text-gray-900">
              レポート
            </Link>
            <Link href="/settings" className="font-medium text-blue-600">
              設定
            </Link>
          </nav>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-8">
        <h1 className="mb-6 text-2xl font-bold">設定</h1>

        {error && (
          <div
            role="alert"
            className="mb-4 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700"
          >
            {error}
          </div>
        )}

        <div className="rounded-lg border border-gray-200 bg-white p-6">
          {loading ? (
            <p className="py-8 text-center text-gray-500">読み込み中...</p>
          ) : (
            <DataSourceSettings
              dataSources={dataSources}
              onConnect={handleConnect}
              onDisconnect={handleDisconnect}
            />
          )}
        </div>
      </main>
    </div>
  );
}
