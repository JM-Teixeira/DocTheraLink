// Prints every serious/critical accessibility problem per page with the
// offending HTML, to fix them. Usage: node tests/a11y-report.mjs [baseUrl]
import AxeBuilder from "@axe-core/playwright";
import { chromium } from "@playwright/test";

const base = process.argv[2] ?? "http://localhost:8088";
const pages = ["/", "/patients", "/patients/new", "/referrals", "/referrals/new", "/time-board", "/inventory", "/users", "/user/preferences"];

const login = await fetch(`${base}/api/v1/auth/login/`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ username: "cm.coordenacao", password: "1234@1" }),
}).then((r) => r.json());

const browser = await chromium.launch();
const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
await ctx.addInitScript(([a, r]) => {
  localStorage.setItem("care_access_token", a);
  localStorage.setItem("care_refresh_token", r);
  localStorage.setItem("i18nextLng", "pt");
}, [login.access, login.refresh]);
const page = await ctx.newPage();
const seen = new Map();
for (const path of pages) {
  await page.goto(base + path);
  await page.waitForLoadState("networkidle");
  const { violations } = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa"]).analyze();
  for (const v of violations.filter((x) => x.impact === "serious" || x.impact === "critical")) {
    for (const n of v.nodes) {
      const key = `${v.id}|${n.html.slice(0, 160)}`;
      if (!seen.has(key)) seen.set(key, { id: v.id, impact: v.impact, html: n.html.slice(0, 220), pages: [] });
      seen.get(key).pages.push(path);
    }
  }
}
for (const s of seen.values()) console.log(`${s.id} [${s.impact}] on ${[...new Set(s.pages)].join(" ")}\n    ${s.html}`);
console.log(`${seen.size} distinct problems`);
await browser.close();
