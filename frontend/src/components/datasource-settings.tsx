"use client";

import type { DataSource, Provider } from "@/types/report";

interface DataSourceSettingsProps {
  dataSources: DataSource[];
  onConnect: (provider: Provider) => void;
  onDisconnect: (provider: Provider) => void;
}

const PROVIDERS: { provider: Provider; label: string; description: string }[] =
  [
    {
      provider: "google",
      label: "Google Calendar",
      description: "予定・会議を自動集約",
    },
    {
      provider: "slack",
      label: "Slack",
      description: "投稿・メッセージを自動集約",
    },
    {
      provider: "github",
      label: "GitHub",
      description: "コミット・PR・Issueを自動集約",
    },
    {
      provider: "gmail",
      label: "Gmail",
      description: "メール送受信を自動集約",
    },
  ];

export function DataSourceSettings({
  dataSources,
  onConnect,
  onDisconnect,
}: DataSourceSettingsProps) {
  const connectedProviders = new Set(dataSources.map((ds) => ds.provider));

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-bold">データソース設定</h2>
      <div className="grid gap-4 sm:grid-cols-2">
        {PROVIDERS.map(({ provider, label, description }) => {
          const isConnected = connectedProviders.has(provider);

          return (
            <div
              key={provider}
              className="rounded-lg border border-gray-200 p-4"
            >
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="font-medium">{label}</h3>
                  <p className="text-sm text-gray-500">{description}</p>
                </div>
                {isConnected ? (
                  <button
                    onClick={() => onDisconnect(provider)}
                    className="rounded bg-red-100 px-3 py-1 text-sm text-red-700 hover:bg-red-200"
                  >
                    切断
                  </button>
                ) : (
                  <button
                    onClick={() => onConnect(provider)}
                    className="rounded bg-blue-100 px-3 py-1 text-sm text-blue-700 hover:bg-blue-200"
                  >
                    接続
                  </button>
                )}
              </div>
              <div className="mt-2">
                <span
                  className={`text-xs ${isConnected ? "text-green-600" : "text-gray-400"}`}
                >
                  {isConnected ? "接続済み" : "未接続"}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
