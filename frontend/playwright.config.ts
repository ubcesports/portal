import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  timeout: 60_000,
  expect: {
    timeout: 10_000,
  },

  fullyParallel: true,

  use: {
    baseURL: "http://localhost:3000",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },

  projects: [
    {
      name: "chromium",
      use: devices["Desktop Chrome"],
    },
  ],

  webServer: [
    {
      command: "make -C .. be",
      url: "http://localhost:8080/health",
      timeout: 120_000,
      reuseExistingServer: true,
    },
    {
      command: "npm run dev",
      url: "http://localhost:3000",
      timeout: 120_000,
      reuseExistingServer: true,
    },
  ],
});