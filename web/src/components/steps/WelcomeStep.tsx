import { useState, useRef, useCallback } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import {
  IconUpload,
  IconFileDescription,
  IconCheck,
} from "@tabler/icons-react";

interface WelcomeStepProps {
  hasManifest: boolean;
  onContinue: () => void;
  onLoadManifest: (content: string) => void;
}

export function WelcomeStep({
  hasManifest,
  onContinue,
  onLoadManifest,
}: WelcomeStepProps) {
  const [dragging, setDragging] = useState(false);
  const [fileName, setFileName] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragging(true);
  }, []);

  const handleDragLeave = useCallback(() => {
    setDragging(false);
  }, []);

  const handleFile = useCallback(
    (file: File) => {
      setFileName(file.name);
      const reader = new FileReader();
      reader.onload = () => {
        if (typeof reader.result === "string") {
          onLoadManifest(reader.result);
        }
      };
      reader.readAsText(file);
    },
    [onLoadManifest]
  );

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragging(false);
      const file = e.dataTransfer.files[0];
      if (file) handleFile(file);
    },
    [handleFile]
  );

  return (
    <div className="flex flex-col items-center justify-center min-h-[70vh] gap-8 px-4">
      <div className="text-center space-y-3">
        <h1 className="text-3xl font-bold tracking-tight">templatr-setup</h1>
        <p className="text-muted-foreground text-lg max-w-md">
          Set up your template in minutes. We'll install everything you need.
        </p>
      </div>

      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <IconFileDescription className="size-5" />
            Manifest File
          </CardTitle>
          <CardDescription>
            The{" "}
            <code className="text-xs bg-secondary px-1 py-0.5 rounded">
              .templatr.toml
            </code>{" "}
            file describes what your template needs.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {hasManifest ? (
            <div className="flex items-center gap-3 p-4 rounded-lg bg-emerald-500/10 border border-emerald-500/20">
              <IconCheck className="size-5 text-emerald-500" />
              <div>
                <p className="font-medium text-sm">Manifest detected</p>
                <p className="text-xs text-muted-foreground">
                  Found in the current directory
                </p>
              </div>
            </div>
          ) : fileName ? (
            <div className="flex items-center gap-3 p-4 rounded-lg bg-primary/10 border border-primary/20">
              <IconFileDescription className="size-5 text-primary" />
              <div>
                <p className="font-medium text-sm">Uploaded</p>
                <p className="text-xs text-muted-foreground">{fileName}</p>
              </div>
            </div>
          ) : (
            <div
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
              className={`flex flex-col items-center gap-3 p-8 rounded-lg border-2 border-dashed cursor-pointer transition-colors ${
                dragging
                  ? "border-primary bg-primary/5"
                  : "border-border hover:border-primary/50"
              }`}
            >
              <IconUpload className="size-8 text-muted-foreground" />
              <p className="text-sm text-muted-foreground text-center">
                Drop your{" "}
                <code className="text-xs bg-secondary px-1 py-0.5 rounded">
                  .templatr.toml
                </code>{" "}
                here or click to browse
              </p>
              <input
                ref={fileInputRef}
                type="file"
                accept=".toml"
                className="hidden"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) handleFile(file);
                }}
              />
            </div>
          )}

          <Button
            onClick={onContinue}
            disabled={!hasManifest}
            className="w-full"
            size="lg"
          >
            Continue
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
