import { useWebSocket } from "@/hooks/useWebSocket";
import { useSetupState } from "@/hooks/useSetupState";
import { WelcomeStep } from "@/components/steps/WelcomeStep";
import { SummaryStep } from "@/components/steps/SummaryStep";
import { InstallStep } from "@/components/steps/InstallStep";
import { ConfigureStep } from "@/components/steps/ConfigureStep";
import { CompleteStep } from "@/components/steps/CompleteStep";
import { IconWifi, IconWifiOff } from "@tabler/icons-react";

function App() {
  const state = useSetupState();
  const { connected, send } = useWebSocket(state.handleMessage);

  const hasManifest = state.plan !== null;

  return (
    <div className="min-h-screen bg-background">
      {/* Connection indicator */}
      <div className="fixed top-4 right-4 z-50">
        {connected ? (
          <div className="flex items-center gap-1.5 text-xs text-emerald-500">
            <IconWifi className="size-3.5" />
            <span>Connected</span>
          </div>
        ) : (
          <div className="flex items-center gap-1.5 text-xs text-destructive">
            <IconWifiOff className="size-3.5" />
            <span>Disconnected</span>
          </div>
        )}
      </div>

      {/* Step indicator */}
      <div className="flex justify-center pt-8 pb-2">
        <StepIndicator currentStep={state.step} />
      </div>

      {/* Step content */}
      {state.step === "welcome" && (
        <WelcomeStep
          hasManifest={hasManifest}
          onContinue={() => {
            if (state.plan) {
              state.setStep("summary");
            }
          }}
        />
      )}

      {state.step === "summary" && state.plan && (
        <SummaryStep
          plan={state.plan}
          onInstall={() => {
            state.setStep("install");
            send({ type: "confirm", action: "install" });
          }}
          onBack={() => state.setStep("welcome")}
        />
      )}

      {state.step === "install" && (
        <InstallStep
          runtimeStatuses={state.runtimeStatuses}
          logs={state.logs}
          error={state.error}
        />
      )}

      {state.step === "configure" && state.plan && (
        <ConfigureStep
          envVars={state.plan.envVars ?? []}
          configs={state.plan.configs ?? []}
          onSubmit={(env, config) => {
            send({ type: "configure", env, config });
          }}
          onSkip={() => {
            send({ type: "configure", env: {}, config: {} });
          }}
        />
      )}

      {state.step === "complete" && (
        <CompleteStep
          success={state.success}
          message={state.completeMessage}
        />
      )}
    </div>
  );
}

const STEPS = [
  "welcome",
  "summary",
  "install",
  "configure",
  "complete",
] as const;
const STEP_LABELS = ["Welcome", "Summary", "Install", "Configure", "Done"];

function StepIndicator({ currentStep }: { currentStep: string }) {
  const currentIndex = STEPS.indexOf(currentStep as (typeof STEPS)[number]);

  return (
    <div className="flex items-center gap-2">
      {STEPS.map((step, i) => {
        const isActive = i === currentIndex;
        const isCompleted = i < currentIndex;

        return (
          <div key={step} className="flex items-center gap-2">
            <div
              className={`flex items-center justify-center size-7 rounded-full text-xs font-medium transition-colors ${
                isActive
                  ? "bg-primary text-primary-foreground"
                  : isCompleted
                    ? "bg-emerald-500/20 text-emerald-400"
                    : "bg-secondary text-muted-foreground"
              }`}
            >
              {isCompleted ? "\u2713" : i + 1}
            </div>
            <span
              className={`text-xs hidden sm:inline ${
                isActive
                  ? "text-foreground font-medium"
                  : "text-muted-foreground"
              }`}
            >
              {STEP_LABELS[i]}
            </span>
            {i < STEPS.length - 1 && (
              <div
                className={`w-6 h-px ${
                  isCompleted ? "bg-emerald-500/40" : "bg-border"
                }`}
              />
            )}
          </div>
        );
      })}
    </div>
  );
}

export default App;
