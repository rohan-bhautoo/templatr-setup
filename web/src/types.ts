// Server → Client message types (matches Go ServerMessage)
export interface ServerMessage {
  type: string;
  step?: string;
  status?: string;
  name?: string;
  version?: string;
  action?: string;
  runtime?: string;
  progress?: number;
  speed?: string;
  total?: string;
  level?: string;
  message?: string;
  success?: boolean;
  plan?: PlanData;
}

export interface PlanData {
  template: TemplateData;
  runtimes: RuntimeData[];
  packages?: PackageData;
  envVars?: EnvVarData[];
  configs?: ConfigData[];
}

export interface TemplateData {
  name: string;
  version: string;
  tier: string;
  category: string;
}

export interface RuntimeData {
  name: string;
  displayName: string;
  requiredVersion: string;
  installedVersion: string;
  action: "skip" | "install" | "upgrade";
}

export interface PackageData {
  manager: string;
  installCommand: string;
  managerFound: boolean;
}

export interface EnvVarData {
  key: string;
  label: string;
  description: string;
  default: string;
  required: boolean;
  type: "text" | "url" | "email" | "secret" | "number" | "boolean";
  docsUrl?: string;
}

export interface ConfigData {
  file: string;
  label: string;
  description: string;
  fields: ConfigFieldData[];
}

export interface ConfigFieldData {
  path: string;
  label: string;
  description: string;
  type: string;
  default: string;
}

// Client → Server message types (matches Go ClientMessage)
export interface ClientMessage {
  type: "load_manifest" | "confirm" | "configure" | "cancel";
  action?: string;
  env?: Record<string, string>;
  config?: Record<string, string>;
  manifestContent?: string;
  manifestPath?: string;
}

// Wizard step
export type WizardStep =
  | "welcome"
  | "summary"
  | "install"
  | "configure"
  | "complete";

// Log entry for the log viewer
export interface LogEntry {
  level: string;
  message: string;
  timestamp: number;
}

// Runtime install status for the install step
export interface RuntimeStatus {
  name: string;
  displayName: string;
  status: "pending" | "installing" | "downloading" | "complete" | "failed";
  progress: number;
  version?: string;
  total?: string;
}
