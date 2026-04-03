"use client";

import type { Activity, Provider } from "@/types/report";

interface ActivityListProps {
  activities: Activity[];
}

export function ActivityList({ activities }: ActivityListProps) {
  if (activities.length === 0) {
    return (
      <div className="py-4 text-center text-gray-500">
        今日の活動データはまだありません
      </div>
    );
  }

  const grouped = groupBySource(activities);

  return (
    <div className="space-y-4">
      {Object.entries(grouped).map(([source, acts]) => (
        <div key={source}>
          <h3 className="mb-2 flex items-center gap-2 font-semibold">
            <SourceIcon provider={source as Provider} />
            {sourceLabel(source as Provider)}
          </h3>
          <ul className="space-y-1">
            {acts.map((activity) => (
              <li
                key={activity.id}
                className="flex items-start gap-2 rounded p-2 hover:bg-gray-50"
              >
                <span className="mt-1 text-xs text-gray-400">
                  {new Date(activity.timestamp).toLocaleTimeString("ja-JP", {
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                </span>
                <span className="text-sm">{activity.title}</span>
              </li>
            ))}
          </ul>
        </div>
      ))}
    </div>
  );
}

function SourceIcon({ provider }: { provider: Provider }) {
  const icons: Record<Provider, string> = {
    google: "📅",
    slack: "💬",
    github: "🔧",
    gmail: "📧",
  };
  return <span>{icons[provider] ?? "📋"}</span>;
}

function sourceLabel(provider: Provider): string {
  const labels: Record<Provider, string> = {
    google: "Google Calendar",
    slack: "Slack",
    github: "GitHub",
    gmail: "Gmail",
  };
  return labels[provider] ?? provider;
}

function groupBySource(activities: Activity[]): Record<Provider, Activity[]> {
  const grouped: Record<string, Activity[]> = {};
  for (const activity of activities) {
    if (!grouped[activity.source]) {
      grouped[activity.source] = [];
    }
    grouped[activity.source].push(activity);
  }
  return grouped as Record<Provider, Activity[]>;
}
