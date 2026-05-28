import { queryOptions } from "@tanstack/react-query";
import { api } from "../api";

export const giteeKeys = {
  all: (wsId: string) => ["gitee", wsId] as const,
  installations: (wsId: string) => [...giteeKeys.all(wsId), "installations"] as const,
  pullRequests: (issueId: string) => ["gitee", "pull-requests", issueId] as const,
};

export const giteeInstallationsOptions = (wsId: string) =>
  queryOptions({
    queryKey: giteeKeys.installations(wsId),
    queryFn: () => api.listGiteeInstallations(wsId),
    enabled: !!wsId,
  });

export const issueGiteePullRequestsOptions = (issueId: string) =>
  queryOptions({
    queryKey: giteeKeys.pullRequests(issueId),
    queryFn: () => api.listIssuePullRequests(issueId),
    enabled: !!issueId,
  });