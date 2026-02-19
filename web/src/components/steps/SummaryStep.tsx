import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { PlanData } from "@/types";
import {
  IconCircleCheck,
  IconDownload,
  IconArrowUp,
  IconArrowLeft,
} from "@tabler/icons-react";

interface SummaryStepProps {
  plan: PlanData;
  onInstall: () => void;
  onBack: () => void;
}

export function SummaryStep({ plan, onInstall, onBack }: SummaryStepProps) {
  const needsAction = plan.runtimes.some((r) => r.action !== "skip");

  return (
    <div className="flex flex-col items-center gap-6 px-4 py-8 max-w-2xl mx-auto">
      <div className="text-center space-y-2">
        <h2 className="text-2xl font-bold">System Summary</h2>
        <p className="text-muted-foreground">
          {plan.template.name}
          {plan.template.tier && (
            <span className="text-xs ml-2">({plan.template.tier})</span>
          )}
        </p>
      </div>

      <Card className="w-full">
        <CardHeader>
          <CardTitle>Runtimes</CardTitle>
          <CardDescription>
            Required runtimes and their status on your system
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {plan.runtimes.map((runtime) => (
              <div
                key={runtime.name}
                className="flex items-center justify-between p-3 rounded-lg bg-secondary/50"
              >
                <div className="flex items-center gap-3">
                  <RuntimeIcon action={runtime.action} />
                  <div>
                    <p className="font-medium text-sm">
                      {runtime.displayName}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      Required: {runtime.requiredVersion}
                      {runtime.installedVersion && (
                        <>
                          {" "}
                          &middot; Installed: {runtime.installedVersion}
                        </>
                      )}
                    </p>
                  </div>
                </div>
                <ActionBadge action={runtime.action} />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {plan.packages && (
        <Card className="w-full">
          <CardHeader>
            <CardTitle>Packages</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between p-3 rounded-lg bg-secondary/50">
              <div>
                <p className="font-medium text-sm">
                  Package Manager: {plan.packages.manager}
                </p>
                <p className="text-xs text-muted-foreground">
                  {plan.packages.installCommand}
                </p>
              </div>
              <Badge
                variant={plan.packages.managerFound ? "secondary" : "outline"}
                className={
                  plan.packages.managerFound
                    ? "bg-emerald-500/20 text-emerald-400"
                    : "bg-amber-500/20 text-amber-400 border-amber-500/30"
                }
              >
                {plan.packages.managerFound ? "Found" : "Not found"}
              </Badge>
            </div>
          </CardContent>
        </Card>
      )}

      <div className="flex gap-3 w-full max-w-sm">
        <Button variant="outline" onClick={onBack} className="flex-1">
          <IconArrowLeft className="size-4" />
          Back
        </Button>
        <Button onClick={onInstall} className="flex-1" size="lg">
          {needsAction ? "Install" : "Continue"}
        </Button>
      </div>
    </div>
  );
}

function RuntimeIcon({ action }: { action: string }) {
  switch (action) {
    case "skip":
      return <IconCircleCheck className="size-5 text-emerald-500" />;
    case "install":
      return <IconDownload className="size-5 text-primary" />;
    case "upgrade":
      return <IconArrowUp className="size-5 text-amber-500" />;
    default:
      return null;
  }
}

function ActionBadge({ action }: { action: string }) {
  switch (action) {
    case "skip":
      return (
        <Badge
          variant="secondary"
          className="bg-emerald-500/20 text-emerald-400"
        >
          OK
        </Badge>
      );
    case "install":
      return <Badge variant="default">Install</Badge>;
    case "upgrade":
      return (
        <Badge variant="secondary" className="bg-amber-500/20 text-amber-400">
          Upgrade
        </Badge>
      );
    default:
      return <Badge variant="outline">{action}</Badge>;
  }
}
