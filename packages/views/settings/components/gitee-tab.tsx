"use client";

import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { ExternalLink, Link2, PanelRight } from "lucide-react";
import { Button } from "@multica/ui/components/ui/button";
import { Card, CardContent } from "@multica/ui/components/ui/card";
import { Label } from "@multica/ui/components/ui/label";
import { Switch } from "@multica/ui/components/ui/switch";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@multica/ui/components/ui/alert-dialog";
import { useAuthStore } from "@multica/core/auth";
import { useWorkspaceId } from "@multica/core/hooks";
import { useCurrentWorkspace } from "@multica/core/paths";
import { memberListOptions, workspaceKeys } from "@multica/core/workspace/queries";
import {
  deriveGiteeSettings,
  giteeInstallationsOptions,
} from "@multica/core/gitee";
import { api } from "@multica/core/api";
import type { Workspace } from "@multica/core/types";
import { useNavigation } from "../../navigation";
import { useT } from "../../i18n";
import { GiteeMark } from "./gitee-mark";

type SettingsKey =
  | "gitee_enabled"
  | "gitee_pr_sidebar_enabled"
  | "gitee_auto_link_prs_enabled";

export function GiteeTab() {
  const { t } = useT("settings");
  const workspace = useCurrentWorkspace();
  const wsId = useWorkspaceId();
  const qc = useQueryClient();
  const navigation = useNavigation();
  const user = useAuthStore((s) => s.user);

  const { data: members = [] } = useQuery(memberListOptions(wsId));
  const currentMember = members.find((m) => m.user_id === user?.id) ?? null;
  const canView = !!currentMember;

  const { data: installationData } = useQuery({
    ...giteeInstallationsOptions(wsId),
    enabled: !!wsId && canView,
  });
  const installations = installationData?.installations ?? [];
  const configured = installationData?.configured ?? false;
  const canManage = installationData?.can_manage === true;
  const connected = installations.length > 0;
  const primaryInstallation = installations[0] ?? null;

  const flags = deriveGiteeSettings(workspace);
  const [savingKey, setSavingKey] = useState<SettingsKey | null>(null);
  const [connecting, setConnecting] = useState(false);
  const [disconnectTarget, setDisconnectTarget] = useState<string | null>(null);
  const [disconnecting, setDisconnecting] = useState(false);

  async function persistSetting(key: SettingsKey, next: boolean) {
    if (!workspace || savingKey) return;
    setSavingKey(key);
    try {
      const merged = {
        ...((workspace.settings as Record<string, unknown>) ?? {}),
        [key]: next,
      };
      const updated = await api.updateWorkspace(workspace.id, { settings: merged });
      qc.setQueryData(workspaceKeys.list(), (old: Workspace[] | undefined) =>
        old?.map((ws) => (ws.id === updated.id ? updated : ws)),
      );
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t(($) => $.gitee.toast_failed));
    } finally {
      setSavingKey(null);
    }
  }

  async function handleConnect() {
    setConnecting(true);
    try {
      const resp = await api.getGiteeConnectURL(wsId);
      if (!resp.configured || !resp.url) {
        toast.error(t(($) => $.gitee.toast_not_configured));
        return;
      }
      window.open(resp.url, "_blank", "noopener");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t(($) => $.gitee.toast_open_failed));
    } finally {
      setConnecting(false);
    }
  }

  async function handleDisconnect() {
    if (!disconnectTarget || disconnecting) return;
    setDisconnecting(true);
    try {
      await api.deleteGiteeInstallation(wsId, disconnectTarget);
      await qc.invalidateQueries({ queryKey: ["gitee", wsId] });
      toast.success(t(($) => $.gitee.toast_disconnected));
      setDisconnectTarget(null);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t(($) => $.gitee.toast_disconnect_failed));
    } finally {
      setDisconnecting(false);
    }
  }

  if (!workspace) return null;

  const repositoriesHref = `${navigation.pathname}?tab=repositories`;

  return (
    <div className="space-y-8">
      <section className="space-y-1">
        <p className="text-sm text-muted-foreground">
          {t(($) => $.gitee.page_description)}
        </p>
      </section>

      <section className="space-y-3">
        <Card>
          <CardContent>
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-start gap-3">
                <div className="rounded-md border bg-muted/50 p-2 text-muted-foreground">
                  <GiteeMark className="h-4 w-4" />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="gitee-master" className="text-sm font-medium">
                    {t(($) => $.gitee.section_master)}
                  </Label>
                  <p className="text-sm text-muted-foreground">
                    {flags.enabled
                      ? t(($) => $.gitee.master_description_on)
                      : t(($) => $.gitee.master_description_off)}
                  </p>
                </div>
              </div>
              <Switch
                id="gitee-master"
                checked={flags.enabled}
                onCheckedChange={(v) => persistSetting("gitee_enabled", v)}
                disabled={!canManage || savingKey === "gitee_enabled"}
              />
            </div>
          </CardContent>
        </Card>
      </section>

      <section className="space-y-3">
        <h2 className="text-sm font-semibold">{t(($) => $.gitee.section_connection)}</h2>
        <Card>
          <CardContent className="space-y-4">
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-start gap-3">
                <GiteeMark className="h-6 w-6 mt-0.5 shrink-0" />
                <div className="space-y-1">
                  <p className="text-sm font-medium">{t(($) => $.gitee.connection_title)}</p>
                  {connected ? (
                    <>
                      <p className="text-xs text-muted-foreground">
                        {t(($) => $.gitee.connected_to, {
                          login: installations.map((i) => i.account_login).join(", "),
                        })}
                      </p>
                      {primaryInstallation?.connected_by && (
                        <p className="text-xs text-muted-foreground">
                          {t(($) => $.gitee.connected_by, {
                            name: primaryInstallation.connected_by!,
                          })}
                        </p>
                      )}
                    </>
                  ) : canManage ? (
                    <p className="text-xs text-muted-foreground">
                      {t(($) => $.gitee.connection_description)}
                    </p>
                  ) : (
                    <p className="text-xs text-muted-foreground">
                      {t(($) => $.gitee.contact_admin_to_connect)}
                    </p>
                  )}
                </div>
              </div>
              {canManage && (
                <div className="flex items-center gap-2">
                  {connected && primaryInstallation ? (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setDisconnectTarget(primaryInstallation.id)}
                    >
                      {t(($) => $.gitee.disconnect)}
                    </Button>
                  ) : (
                    <Button
                      size="sm"
                      onClick={handleConnect}
                      disabled={connecting || !configured}
                      title={
                        !configured
                          ? t(($) => $.gitee.connect_disabled_tooltip)
                          : undefined
                      }
                    >
                      {connecting
                        ? t(($) => $.gitee.connect_opening)
                        : t(($) => $.gitee.connect_gitee)}
                    </Button>
                  )}
                </div>
              )}
            </div>

            {canManage && !configured && (
              <p className="text-xs text-muted-foreground">
                {t(($) => $.gitee.not_configured)}
              </p>
            )}

            {!canManage && connected && (
              <p className="text-xs text-muted-foreground">
                {t(($) => $.gitee.read_only_hint)}
              </p>
            )}
          </CardContent>
        </Card>
      </section>

      <section className="space-y-3">
        <h2 className="text-sm font-semibold">{t(($) => $.gitee.section_features)}</h2>
        <Card>
          <CardContent className="space-y-4">
            <FeatureRow
              id="gitee-pr-sidebar"
              icon={<PanelRight className="h-4 w-4" />}
              label={t(($) => $.gitee.feature_pr_sidebar_label)}
              description={
                <p className="text-sm text-muted-foreground">
                  {t(($) => $.gitee.feature_pr_sidebar_description)}
                </p>
              }
              checked={flags.prSidebar}
              disabled={!canManage || !flags.enabled || savingKey === "gitee_pr_sidebar_enabled"}
              onCheckedChange={(v) => persistSetting("gitee_pr_sidebar_enabled", v)}
            />

            <FeatureRow
              id="gitee-auto-link"
              icon={<Link2 className="h-4 w-4" />}
              label={t(($) => $.gitee.feature_auto_link_label)}
              description={
                <p className="text-sm text-muted-foreground">
                  {t(($) => $.gitee.feature_auto_link_description)}
                </p>
              }
              checked={flags.autoLinkPRs}
              disabled={!canManage || !flags.enabled || savingKey === "gitee_auto_link_prs_enabled"}
              onCheckedChange={(v) => persistSetting("gitee_auto_link_prs_enabled", v)}
            />
          </CardContent>
        </Card>
      </section>

      <section className="space-y-3">
        <h2 className="text-sm font-semibold">{t(($) => $.gitee.section_repositories)}</h2>
        <Card>
          <CardContent>
            <div className="flex flex-wrap items-center justify-between gap-3">
              <p className="text-sm font-medium">
                {t(($) => $.gitee.repositories_shortcut_label)}
              </p>
              <Button
                variant="outline"
                size="sm"
                onClick={() => navigation.push(repositoriesHref)}
              >
                <ExternalLink className="h-3 w-3" />
                {t(($) => $.gitee.repositories_shortcut_link)}
              </Button>
            </div>
          </CardContent>
        </Card>
      </section>

      <AlertDialog
        open={!!disconnectTarget}
        onOpenChange={(v) => {
          if (!v && !disconnecting) setDisconnectTarget(null);
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t(($) => $.gitee.disconnect_confirm_title)}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t(($) => $.gitee.disconnect_confirm_description)}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={disconnecting}>
              {t(($) => $.gitee.disconnect_confirm_cancel)}
            </AlertDialogCancel>
            <AlertDialogAction onClick={handleDisconnect} disabled={disconnecting}>
              {disconnecting
                ? t(($) => $.gitee.disconnecting)
                : t(($) => $.gitee.disconnect_confirm_action)}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function FeatureRow({
  id,
  icon,
  label,
  description,
  checked,
  disabled,
  onCheckedChange,
}: {
  id: string;
  icon: React.ReactNode;
  label: string;
  description: React.ReactNode;
  checked: boolean;
  disabled: boolean;
  onCheckedChange: (v: boolean) => void;
}) {
  return (
    <div className="flex items-start justify-between gap-4">
      <div className="flex items-start gap-3">
        <div className="rounded-md border bg-muted/50 p-2 text-muted-foreground">{icon}</div>
        <div className="space-y-1">
          <Label htmlFor={id} className="text-sm font-medium">
            {label}
          </Label>
          {description}
        </div>
      </div>
      <Switch
        id={id}
        checked={checked}
        disabled={disabled}
        onCheckedChange={onCheckedChange}
      />
    </div>
  );
}