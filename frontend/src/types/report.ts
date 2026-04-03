export type ReportType = "daily" | "weekly" | "monthly";

export type ReportStatus = "draft" | "confirmed" | "sent";

export interface Report {
  id: string;
  user_id: string;
  type: ReportType;
  template_id: string;
  content: string;
  date: string;
  status: ReportStatus;
  created_at: string;
  updated_at: string;
}

export interface Activity {
  id: string;
  user_id: string;
  source: Provider;
  title: string;
  body: string;
  timestamp: string;
  metadata: string;
  created_at: string;
}

export type Provider = "google" | "slack" | "github" | "gmail";

export interface Template {
  id: string;
  user_id: string;
  name: string;
  type: ReportType;
  sections: TemplateSection[];
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

export interface TemplateSection {
  title: string;
  order: number;
}

export interface DataSource {
  id: string;
  user_id: string;
  provider: Provider;
  expires_at: string;
  created_at: string;
  updated_at: string;
}

export interface ApiError {
  error: string;
  code: string;
}
