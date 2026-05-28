import type { Workspace } from "../types";

export interface GiteeSettings {
  enabled: boolean;
  prSidebar: boolean;
  autoLinkPRs: boolean;
}

export function deriveGiteeSettings(
  workspace: Pick<Workspace, "settings"> | null | undefined,
): GiteeSettings {
  const s = (workspace?.settings ?? {}) as Record<string, unknown>;
  const enabled = s.gitee_enabled !== false;
  return {
    enabled,
    prSidebar: enabled && s.gitee_pr_sidebar_enabled !== false,
    autoLinkPRs: enabled && s.gitee_auto_link_prs_enabled !== false,
  };
}