export type GiteePullRequestState = "open" | "closed" | "merged" | "draft";

export interface GiteeInstallation {
  id: string;
  workspace_id: string;
  installation_id?: number;
  account_login: string;
  account_type: "User" | "Organization";
  account_avatar_url: string | null;
  created_at: string;
  connected_by?: string;
}

export interface GiteePullRequest {
  id: string;
  workspace_id: string;
  repo_owner: string;
  repo_name: string;
  number: number;
  title: string;
  state: GiteePullRequestState;
  html_url: string;
  branch: string | null;
  author_login: string | null;
  author_avatar_url: string | null;
  merged_at: string | null;
  closed_at: string | null;
  pr_created_at: string;
  pr_updated_at: string;
  mergeable_state?: string | null;
  checks_conclusion?: "passed" | "failed" | "pending" | null;
}

export interface ListGiteeInstallationsResponse {
  installations: GiteeInstallation[];
  configured: boolean;
  can_manage?: boolean;
}

export interface GiteeConnectResponse {
  url?: string;
  configured: boolean;
}