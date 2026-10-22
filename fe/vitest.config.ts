import path from "node:path";
import { defineConfig } from "vitest/config";

// Unit/component tests (Vitest + Testing Library). Cypress covers end-to-end.
export default defineConfig({
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    include: ["src/**/*.test.{ts,tsx}"],
    css: false,
    coverage: {
      provider: "v8",
      include: [
        "src/components/Referral/**/*.{ts,tsx}",
        "src/components/Patient/**/*.{ts,tsx}",
      ],
      exclude: ["**/*.test.{ts,tsx}"],
      reporter: ["text", "html"],
      // The patient flow must stay at least 80 % covered (CI runs test:coverage)
      thresholds: {
        "src/components/Patient/**": { lines: 80, statements: 80 },
      },
    },
  },
  // Keep in sync with resolve.alias in vite.config.mts
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "@careConfig": path.resolve(__dirname, "./care.config.ts"),
      "@core": path.resolve(__dirname, "./src"),
    },
  },
});
