/**
 * Core Engine lifecycle manager.
 * Starts, health-checks, and stops the Go Core Engine binary.
 */
import { spawn, type ChildProcess } from "node:child_process";
import { existsSync } from "node:fs";
import { join } from "node:path";

export class EngineManager {
  private process: ChildProcess | null = null;
  private readonly baseUrl: string;
  private readonly binaryPath: string;
  private readonly reflectorDir: string;
  private started = false;

  constructor(port: number, basePath: string) {
    this.baseUrl = `http://127.0.0.1:${port}`;
    this.reflectorDir = join(basePath, ".reflector");
    // Binary is expected at .reflector/bin/reflector
    this.binaryPath = join(this.reflectorDir, "bin", "reflector");
  }

  getBaseUrl(): string {
    return this.baseUrl;
  }

  getReflectorDir(): string {
    return this.reflectorDir;
  }

  /**
   * Ensure the Core Engine is running.
   * If already running (health check passes), reuse it.
   * Otherwise, start it with retries.
   */
  async ensureRunning(): Promise<boolean> {
    // First check if already running
    if (await this.healthCheck()) {
      console.log("[reflector] Core Engine already running, reusing");
      this.started = true;
      return true;
    }

    // Check binary exists
    if (!existsSync(this.binaryPath)) {
      console.error(`[reflector] Core Engine binary not found at ${this.binaryPath}`);
      console.error("[reflector] Run 'reflector build' or build manually");
      return false;
    }

    // Start with retries (3 attempts, exponential backoff)
    for (let attempt = 1; attempt <= 3; attempt++) {
      console.log(`[reflector] Starting Core Engine (attempt ${attempt}/3)...`);

      try {
        this.startProcess();
        // Wait for health check with 10s timeout
        const healthy = await this.waitForHealth(10_000);
        if (healthy) {
          console.log("[reflector] Core Engine started successfully");
          this.started = true;
          return true;
        }
      } catch (err) {
        console.warn(`[reflector] Start attempt ${attempt} failed:`, err);
      }

      // Stop if started but unhealthy
      this.stopProcess();

      // Exponential backoff: 1s, 2s
      if (attempt < 3) {
        const delay = Math.pow(2, attempt - 1) * 1000;
        await sleep(delay);
      }
    }

    console.error("[reflector] Core Engine failed to start after 3 attempts, degrading gracefully");
    return false;
  }

  /**
   * Gracefully stop the Core Engine.
   */
  async shutdown(): Promise<void> {
    if (!this.process) return;

    console.log("[reflector] Shutting down Core Engine...");
    this.process.kill("SIGTERM");

    // Wait up to 5 seconds for graceful shutdown
    const deadline = Date.now() + 5000;
    while (this.process && Date.now() < deadline) {
      await sleep(200);
    }

    // Force kill if still running
    if (this.process) {
      console.warn("[reflector] Force killing Core Engine");
      this.process.kill("SIGKILL");
      this.process = null;
    }

    this.started = false;
  }

  isRunning(): boolean {
    return this.started;
  }

  /**
   * Call the Core Engine's reflect API.
   */
  async reflect(body: {
    trigger_type: string;
    trigger_detail: string;
    sessions: unknown[];
    config: { sentiment_enabled: boolean; sentiment_source: string };
  }): Promise<{ status: string; sessions_analyzed: number; report_path: string; tokens_consumed: number; duration_ms: number } | null> {
    if (!this.started) {
      console.warn("[reflector] Core Engine not running, skipping reflect");
      return null;
    }

    try {
      const resp = await fetch(`${this.baseUrl}/api/v1/reflect`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      if (!resp.ok) {
        console.error(`[reflector] Reflect API returned ${resp.status}`);
        return null;
      }

      return (await resp.json()) as {
        status: string;
        sessions_analyzed: number;
        report_path: string;
        tokens_consumed: number;
        duration_ms: number;
      };
    } catch (err) {
      console.error("[reflector] Reflect API call failed:", err);
      return null;
    }
  }

  /**
   * Call the Core Engine's cleanup API.
   */
  async cleanup(days: number): Promise<{ status: string; days: string } | null> {
    if (!this.started) return null;

    try {
      const resp = await fetch(`${this.baseUrl}/api/v1/cleanup`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ days }),
      });

      if (!resp.ok) {
        console.error(`[reflector] Cleanup API returned ${resp.status}`);
        return null;
      }

      return (await resp.json()) as { status: string; days: string };
    } catch (err) {
      console.error("[reflector] Cleanup API call failed:", err);
      return null;
    }
  }

  private startProcess(): void {
    this.process = spawn(this.binaryPath, [], {
      cwd: this.reflectorDir,
      env: { ...process.env },
      stdio: ["ignore", "pipe", "pipe"],
    });

    this.process.stdout?.on("data", (data: Buffer) => {
      const lines = data.toString().trim().split("\n");
      for (const line of lines) {
        console.log(`[reflector:engine] ${line}`);
      }
    });

    this.process.stderr?.on("data", (data: Buffer) => {
      const lines = data.toString().trim().split("\n");
      for (const line of lines) {
        console.error(`[reflector:engine] ${line}`);
      }
    });

    this.process.on("exit", (code) => {
      console.log(`[reflector] Core Engine exited with code ${code}`);
      this.process = null;
      this.started = false;
    });
  }

  private stopProcess(): void {
    if (this.process) {
      this.process.kill("SIGTERM");
      this.process = null;
    }
    this.started = false;
  }

  private async healthCheck(): Promise<boolean> {
    try {
      const resp = await fetch(`${this.baseUrl}/api/v1/health`, {
        signal: AbortSignal.timeout(2000),
      });
      if (resp.ok) {
        const body = (await resp.json()) as { status: string };
        return body.status === "ok";
      }
      return false;
    } catch {
      return false;
    }
  }

  private async waitForHealth(timeoutMs: number): Promise<boolean> {
    const deadline = Date.now() + timeoutMs;
    while (Date.now() < deadline) {
      if (await this.healthCheck()) return true;
      await sleep(500);
    }
    return false;
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
