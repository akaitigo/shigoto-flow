import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { ActivityList } from "../activity-list";
import type { Activity } from "@/types/report";

const mockActivities: Activity[] = [
  {
    id: "1",
    user_id: "user1",
    source: "google",
    title: "チームミーティング",
    body: "",
    timestamp: "2026-04-04T10:00:00Z",
    metadata: "",
    created_at: "2026-04-04T10:00:00Z",
  },
  {
    id: "2",
    user_id: "user1",
    source: "github",
    title: "repo/project にプッシュ",
    body: "",
    timestamp: "2026-04-04T14:00:00Z",
    metadata: "PushEvent",
    created_at: "2026-04-04T14:00:00Z",
  },
  {
    id: "3",
    user_id: "user1",
    source: "slack",
    title: "#dev での投稿",
    body: "Hello world",
    timestamp: "2026-04-04T11:00:00Z",
    metadata: "",
    created_at: "2026-04-04T11:00:00Z",
  },
];

describe("ActivityList", () => {
  it("renders activities grouped by source", () => {
    render(<ActivityList activities={mockActivities} />);

    expect(screen.getByText("Google Calendar")).toBeInTheDocument();
    expect(screen.getByText("GitHub")).toBeInTheDocument();
    expect(screen.getByText("Slack")).toBeInTheDocument();
  });

  it("renders activity titles", () => {
    render(<ActivityList activities={mockActivities} />);

    expect(screen.getByText("チームミーティング")).toBeInTheDocument();
    expect(screen.getByText("repo/project にプッシュ")).toBeInTheDocument();
    expect(screen.getByText("#dev での投稿")).toBeInTheDocument();
  });

  it("shows empty state when no activities", () => {
    render(<ActivityList activities={[]} />);

    expect(
      screen.getByText("今日の活動データはまだありません"),
    ).toBeInTheDocument();
  });
});
