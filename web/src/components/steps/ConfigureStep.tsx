import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import type { EnvVarData, ConfigData } from "@/types";
import { IconArrowRight, IconPlayerSkipForward } from "@tabler/icons-react";

interface ConfigureStepProps {
  envVars: EnvVarData[];
  configs: ConfigData[];
  onSubmit: (
    env: Record<string, string>,
    config: Record<string, string>
  ) => void;
  onSkip: () => void;
}

export function ConfigureStep({
  envVars,
  configs,
  onSubmit,
  onSkip,
}: ConfigureStepProps) {
  const [envValues, setEnvValues] = useState<Record<string, string>>(() => {
    const defaults: Record<string, string> = {};
    for (const ev of envVars) {
      defaults[ev.key] = ev.default;
    }
    return defaults;
  });

  const [configValues, setConfigValues] = useState<Record<string, string>>(
    () => {
      const defaults: Record<string, string> = {};
      for (const cfg of configs) {
        for (const field of cfg.fields) {
          defaults[field.path] = field.default;
        }
      }
      return defaults;
    }
  );

  const handleSubmit = () => {
    onSubmit(envValues, configValues);
  };

  return (
    <div className="flex flex-col items-center gap-6 px-4 py-8 max-w-2xl mx-auto">
      <div className="text-center space-y-2">
        <h2 className="text-2xl font-bold">Configure</h2>
        <p className="text-muted-foreground">
          Fill in your template settings. You can change these later.
        </p>
      </div>

      {envVars.length > 0 && (
        <Card className="w-full">
          <CardHeader>
            <CardTitle>Environment Variables</CardTitle>
            <CardDescription>
              These values will be written to your{" "}
              <code className="text-xs bg-secondary px-1 py-0.5 rounded">
                .env
              </code>{" "}
              file
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {envVars.map((ev) => (
              <div key={ev.key} className="space-y-1.5">
                <label className="text-sm font-medium" htmlFor={ev.key}>
                  {ev.label}
                  {ev.required && (
                    <span className="text-destructive ml-1">*</span>
                  )}
                </label>
                {ev.description && (
                  <p className="text-xs text-muted-foreground">
                    {ev.description}
                  </p>
                )}
                <Input
                  id={ev.key}
                  type={
                    ev.type === "secret"
                      ? "password"
                      : ev.type === "number"
                        ? "number"
                        : "text"
                  }
                  placeholder={ev.default || ev.label}
                  value={envValues[ev.key] ?? ""}
                  onChange={(e) =>
                    setEnvValues((prev) => ({
                      ...prev,
                      [ev.key]: e.target.value,
                    }))
                  }
                />
                {ev.docsUrl && (
                  <a
                    href={ev.docsUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-xs text-primary hover:underline"
                  >
                    Documentation
                  </a>
                )}
              </div>
            ))}
          </CardContent>
        </Card>
      )}

      {configs.map((cfg) => (
        <Card key={cfg.file} className="w-full">
          <CardHeader>
            <CardTitle>{cfg.label}</CardTitle>
            <CardDescription>
              {cfg.description}
              <span className="block text-xs mt-1 font-mono">{cfg.file}</span>
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {cfg.fields.map((field) => (
              <div key={field.path} className="space-y-1.5">
                <label className="text-sm font-medium" htmlFor={field.path}>
                  {field.label}
                </label>
                {field.description && (
                  <p className="text-xs text-muted-foreground">
                    {field.description}
                  </p>
                )}
                <Input
                  id={field.path}
                  type={field.type === "number" ? "number" : "text"}
                  placeholder={field.default || field.label}
                  value={configValues[field.path] ?? ""}
                  onChange={(e) =>
                    setConfigValues((prev) => ({
                      ...prev,
                      [field.path]: e.target.value,
                    }))
                  }
                />
              </div>
            ))}
          </CardContent>
        </Card>
      ))}

      <div className="flex gap-3 w-full max-w-sm">
        <Button variant="outline" onClick={onSkip} className="flex-1">
          <IconPlayerSkipForward className="size-4" />
          Skip
        </Button>
        <Button onClick={handleSubmit} className="flex-1" size="lg">
          Save
          <IconArrowRight className="size-4" />
        </Button>
      </div>
    </div>
  );
}
