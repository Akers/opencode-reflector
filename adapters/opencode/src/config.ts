/**
 * Configuration reader for opencode-reflector.
 * Reads from .reflector/reflector.yaml (shared with Core Engine).
 */
import { readFileSync, existsSync } from "node:fs";
import { join } from "node:path";
import { parse as parseYaml } from "yaml";

export interface ReflectorConfig {
  port: number;
  trigger: {
    time: {
      enabled: boolean;
      schedule: string; // e.g. "00:00"
    };
    events: {
      enabled: boolean;
      types: string[];
      messageInterval: number;
    };
  };
  model: {
    id: string;           // 兼容旧配置
    override: string;     // 强制覆盖宿主默认模型，为空则自动获取
  };
  sentiment: {
    enabled: boolean;
    skipOnMessageTrigger: boolean;
    mode: "agent" | "builtin" | "off";  // agent=SDK Prompt, builtin=规则引擎, off=关闭
  };
  classification: {
    enabled: boolean;               // 是否启用 L2 LLM 分类
    confidenceThreshold: number;    // L2 置信度阈值
  };
  retention: {
    days: number;
  };
  report: {
    template: string;
  };
  logLevel: string;
}

const DEFAULT_CONFIG: ReflectorConfig = {
  port: 19870,
  trigger: {
    time: { enabled: true, schedule: "00:00" },
    events: { enabled: true, types: ["TASK_FINISHED", "N_MESSAGES"], messageInterval: 10 },
  },
  model: { id: "", override: "" },
  sentiment: { enabled: true, skipOnMessageTrigger: true, mode: "agent" },
  classification: { enabled: true, confidenceThreshold: 0.7 },
  retention: { days: 90 },
  report: { template: "default" },
  logLevel: "info",
};

/**
 * Load reflector configuration from a base directory.
 * Looks for `.reflector/reflector.yaml` under basePath.
 * Falls back to defaults for missing fields.
 */
export function loadConfig(basePath: string): ReflectorConfig {
  const configPath = join(basePath, ".reflector", "reflector.yaml");

  if (!existsSync(configPath)) {
    return { ...DEFAULT_CONFIG };
  }

  try {
    const raw = readFileSync(configPath, "utf-8");
    const parsed = parseYaml(raw);
    if (!parsed || typeof parsed !== "object") {
      return { ...DEFAULT_CONFIG };
    }
    return deepMerge(DEFAULT_CONFIG as unknown as Record<string, unknown>, parsed as Record<string, unknown>) as unknown as ReflectorConfig;
  } catch (err) {
    console.warn("[reflector] Failed to read config, using defaults:", err);
    return { ...DEFAULT_CONFIG };
  }
}

function deepMerge(target: Record<string, unknown>, source: Record<string, unknown>): Record<string, unknown> {
  const result = { ...target };
  for (const key of Object.keys(source)) {
    const srcVal = source[key];
    const tgtVal = target[key];
    if (
      srcVal &&
      typeof srcVal === "object" &&
      !Array.isArray(srcVal) &&
      tgtVal &&
      typeof tgtVal === "object" &&
      !Array.isArray(tgtVal)
    ) {
      result[key] = deepMerge(tgtVal as Record<string, unknown>, srcVal as Record<string, unknown>);
    } else {
      result[key] = srcVal;
    }
  }
  return result;
}
