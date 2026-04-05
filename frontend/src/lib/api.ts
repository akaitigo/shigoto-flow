import type {
  Report,
  Activity,
  Template,
  DataSource,
  ReportType,
  ApiError,
} from "@/types/report";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

async function fetchAPI<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  if (!res.ok) {
    const err: ApiError = await res.json();
    throw new Error(err.error || `API error: ${res.status}`);
  }

  return res.json() as Promise<T>;
}

export async function listReports(
  type: ReportType,
  limit = 20,
  offset = 0,
): Promise<Report[]> {
  return fetchAPI<Report[]>(
    `/reports?type=${type}&limit=${limit}&offset=${offset}`,
  );
}

export async function getReport(id: string): Promise<Report> {
  return fetchAPI<Report>(`/reports/${id}`);
}

export async function createReport(data: {
  type: ReportType;
  template_id: string;
  content: string;
  date: string;
}): Promise<Report> {
  return fetchAPI<Report>("/reports", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateReport(
  id: string,
  data: { content: string; status: string },
): Promise<{ status: string }> {
  return fetchAPI<{ status: string }>(`/reports/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function generateReport(data: {
  type: ReportType;
  date: string;
}): Promise<Report> {
  return fetchAPI<Report>("/reports/generate", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function listActivities(date: string): Promise<Activity[]> {
  return fetchAPI<Activity[]>(`/activities?date=${date}`);
}

export async function collectActivities(): Promise<{
  status: string;
  message: string;
}> {
  return fetchAPI<{ status: string; message: string }>("/activities/collect", {
    method: "POST",
  });
}

export async function listTemplates(): Promise<Template[]> {
  return fetchAPI<Template[]>("/templates");
}

export async function createTemplate(data: {
  name: string;
  type: ReportType;
  sections: { title: string; order: number }[];
  is_default: boolean;
}): Promise<Template> {
  return fetchAPI<Template>("/templates", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function listDataSources(): Promise<DataSource[]> {
  return fetchAPI<DataSource[]>("/datasources");
}

export async function deleteDataSource(
  provider: string,
): Promise<{ status: string }> {
  return fetchAPI<{ status: string }>(`/datasources/${provider}`, {
    method: "DELETE",
  });
}
