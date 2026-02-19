import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import type { LogEntry, RuntimeStatus } from "@/types";
import {
  IconCircleCheck,
  IconCircleX,
  IconLoader2,
  IconClock,
} from "@tabler/icons-react";

interface InstallStepProps {
  runtimeStatuses: RuntimeStatus[];
  logs: LogEntry[];
  error: string | null;
}

export function InstallStep({
  runtimeStatuses,
  logs,
  error,
}: InstallStepProps) {
  const completedCount = runtimeStatuses.filter(
    (r) => r.status === "complete"
  ).length;
  const totalCount = runtimeStatuses.length;

  return (
    <div className="flex flex-col items-center gap-6 px-4 py-8 max-w-2xl mx-auto">
      <div className="text-center space-y-2">
        <h2 className="text-2xl font-bold">Installing</h2>
        <p className="text-muted-foreground">
          {completedCount} of {totalCount} runtimes complete
        </p>
      </div>

      {error && (
        <div className="w-full p-4 rounded-lg bg-destructive/10 border border-destructive/20 text-destructive text-sm">
          {error}
        </div>
      )}

      <Card className="w-full">
        <CardHeader>
          <CardTitle>Progress</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {runtimeStatuses.map((rs) => (
            <div key={rs.name} className="space-y-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <StatusIcon status={rs.status} />
                  <span className="text-sm font-medium">{rs.displayName}</span>
                </div>
                <span className="text-xs text-muted-foreground">
                  {rs.status === "downloading" && rs.total && `${rs.total}`}
                  {rs.status === "complete" && rs.version && `v${rs.version}`}
                  {rs.status === "pending" && "Waiting"}
                  {rs.status === "installing" && "Installing..."}
                  {rs.status === "failed" && "Failed"}
                </span>
              </div>
              {(rs.status === "downloading" || rs.status === "installing") && (
                <Progress value={rs.progress} />
              )}
              {rs.status === "complete" && <Progress value={100} />}
            </div>
          ))}
        </CardContent>
      </Card>

      {logs.length > 0 && (
        <Card className="w-full">
          <CardHeader>
            <CardTitle>Log</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="bg-background rounded-lg p-3 max-h-48 overflow-y-auto font-mono text-xs space-y-1">
              {logs.map((log, i) => (
                <div key={i} className={logColor(log.level)}>
                  {log.message}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function StatusIcon({ status }: { status: RuntimeStatus["status"] }) {
  switch (status) {
    case "complete":
      return <IconCircleCheck className="size-4 text-emerald-500" />;
    case "failed":
      return <IconCircleX className="size-4 text-destructive" />;
    case "downloading":
    case "installing":
      return <IconLoader2 className="size-4 text-primary animate-spin" />;
    case "pending":
      return <IconClock className="size-4 text-muted-foreground" />;
  }
}

function logColor(level: string): string {
  switch (level) {
    case "error":
      return "text-destructive";
    case "warn":
      return "text-amber-500";
    default:
      return "text-muted-foreground";
  }
}
