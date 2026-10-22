import { defineConfig, devices } from "@playwright/test";

// Runs against an already running stack (make k8s-start, or compose):
// E2E_WEB_URL=http://localhost:8088 by default.
export default defineConfig({
  testDir: "./tests",
  timeout: 30_000,
  fullyParallel: false, // tests share demo users (preferences, filters)
  retries: process.env.CI ? 1 : 0,
  reporter: [["list"], ["html", { open: "never", outputFolder: "report" }]],
  use: {
    baseURL: process.env.E2E_WEB_URL ?? "http://localhost:8088",
    viewport: { width: 1440, height: 900 },
    screenshot: "only-on-failure",
    trace: "retain-on-failure",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"], viewport: { width: 1440, height: 900 } } }],
});
