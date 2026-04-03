"use client";

import { DataSourceSettings } from "@/components/datasource-settings";
import type { Provider } from "@/types/report";

export default function SettingsPage() {
  const handleConnect = (provider: Provider) => {
    window.location.href = `/api/v1/auth/${provider}`;
  };

  const handleDisconnect = (_provider: Provider) => {
    // Will be implemented with actual API integration
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="border-b bg-white">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <a href="/" className="text-xl font-bold text-gray-900">
            Shigoto-Flow
          </a>
          <nav className="flex gap-6">
            <a href="/reports" className="text-gray-600 hover:text-gray-900">
              レポート
            </a>
            <a href="/settings" className="font-medium text-blue-600">
              設定
            </a>
          </nav>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-8">
        <h1 className="mb-6 text-2xl font-bold">設定</h1>

        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <DataSourceSettings
            dataSources={[]}
            onConnect={handleConnect}
            onDisconnect={handleDisconnect}
          />
        </div>
      </main>
    </div>
  );
}
