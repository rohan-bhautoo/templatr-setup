import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  IconCircleCheck,
  IconCircleX,
  IconCopy,
  IconCheck,
} from "@tabler/icons-react";

interface CompleteStepProps {
  success: boolean;
  message: string | null;
  logFilePath?: string;
}

export function CompleteStep({
  success,
  message,
  logFilePath,
}: CompleteStepProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-[70vh] gap-8 px-4">
      <div className="text-center space-y-4">
        {success ? (
          <IconCircleCheck className="size-16 text-emerald-500 mx-auto" />
        ) : (
          <IconCircleX className="size-16 text-destructive mx-auto" />
        )}
        <h2 className="text-2xl font-bold">
          {success ? "Setup Complete!" : "Setup Failed"}
        </h2>
      </div>

      {message && (
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>{success ? "Next Steps" : "Error Details"}</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="text-sm text-muted-foreground whitespace-pre-wrap leading-relaxed">
              {message.trim()}
            </pre>
          </CardContent>
        </Card>
      )}

      {success && (
        <div className="w-full max-w-md">
          <button
            onClick={() => handleCopy("npm run dev")}
            className="w-full flex items-center justify-between p-3 rounded-lg bg-secondary/50 hover:bg-secondary/80 transition-colors cursor-pointer"
          >
            <code className="text-sm font-mono">npm run dev</code>
            {copied ? (
              <IconCheck className="size-4 text-emerald-500" />
            ) : (
              <IconCopy className="size-4 text-muted-foreground" />
            )}
          </button>
        </div>
      )}

      {logFilePath && (
        <p className="text-xs text-muted-foreground">
          Log file: {logFilePath}
        </p>
      )}

      <Button
        variant="outline"
        onClick={() => window.close()}
        className="mt-4"
      >
        Close
      </Button>
    </div>
  );
}
