import { useCallback, useState } from "react";
import type {
  LogEntry,
  PlanData,
  RuntimeStatus,
  ServerMessage,
  WizardStep,
} from "../types";

interface SetupState {
  step: WizardStep;
  plan: PlanData | null;
  runtimeStatuses: RuntimeStatus[];
  logs: LogEntry[];
  error: string | null;
  completeMessage: string | null;
  success: boolean;
}

interface UseSetupStateReturn extends SetupState {
  setStep: (step: WizardStep) => void;
  handleMessage: (msg: ServerMessage) => void;
}

export function useSetupState(): UseSetupStateReturn {
  const [state, setState] = useState<SetupState>({
    step: "welcome",
    plan: null,
    runtimeStatuses: [],
    logs: [],
    error: null,
    completeMessage: null,
    success: false,
  });

  const setStep = useCallback((step: WizardStep) => {
    setState((prev) => ({ ...prev, step }));
  }, []);

  const handleMessage = useCallback((msg: ServerMessage) => {
    setState((prev) => {
      switch (msg.type) {
        case "plan": {
          const statuses: RuntimeStatus[] =
            msg.plan?.runtimes.map((r) => ({
              name: r.name,
              displayName: r.displayName,
              status:
                r.action === "skip"
                  ? ("complete" as const)
                  : ("pending" as const),
              progress: r.action === "skip" ? 100 : 0,
              version: r.installedVersion || undefined,
            })) ?? [];

          return {
            ...prev,
            plan: msg.plan ?? null,
            runtimeStatuses: statuses,
            step: "summary",
            error: null,
          };
        }

        case "step": {
          if (msg.step === "configure" && msg.status === "ready") {
            return { ...prev, step: "configure" };
          }
          return prev;
        }

        case "runtime": {
          if (msg.status === "installing") {
            return {
              ...prev,
              runtimeStatuses: prev.runtimeStatuses.map((rs) =>
                rs.name === msg.name
                  ? { ...rs, status: "installing" as const }
                  : rs
              ),
            };
          }
          if (msg.action === "skip") {
            return {
              ...prev,
              runtimeStatuses: prev.runtimeStatuses.map((rs) =>
                rs.name === msg.name
                  ? {
                      ...rs,
                      status: "complete" as const,
                      progress: 100,
                      version: msg.version,
                    }
                  : rs
              ),
            };
          }
          return prev;
        }

        case "download": {
          return {
            ...prev,
            runtimeStatuses: prev.runtimeStatuses.map((rs) =>
              rs.name === msg.runtime
                ? {
                    ...rs,
                    status: "downloading" as const,
                    progress: msg.progress ?? 0,
                    total: msg.total,
                  }
                : rs
            ),
          };
        }

        case "install": {
          if (msg.status === "complete") {
            return {
              ...prev,
              runtimeStatuses: prev.runtimeStatuses.map((rs) =>
                rs.name === msg.runtime
                  ? {
                      ...rs,
                      status: "complete" as const,
                      progress: 100,
                      version: msg.version,
                    }
                  : rs
              ),
            };
          }
          return prev;
        }

        case "log": {
          return {
            ...prev,
            logs: [
              ...prev.logs,
              {
                level: msg.level ?? "info",
                message: msg.message ?? "",
                timestamp: Date.now(),
              },
            ],
          };
        }

        case "error": {
          return {
            ...prev,
            error: msg.message ?? "An unknown error occurred",
          };
        }

        case "complete": {
          return {
            ...prev,
            step: "complete",
            success: msg.success ?? false,
            completeMessage: msg.message ?? null,
          };
        }

        default:
          return prev;
      }
    });
  }, []);

  return {
    ...state,
    setStep,
    handleMessage,
  };
}
