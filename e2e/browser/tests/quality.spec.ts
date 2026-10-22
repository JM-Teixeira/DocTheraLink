import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";

import { api, signIn } from "./helpers";

// Cross-cutting checks: notifications, printing, small screens, accessibility.

test("notifications bell: new referral shows up, opens it, marks read", async ({ page, request }) => {
  const doctorTok = await (await request.post("/api/v1/auth/login/", { data: { username: "cm.helena.carvalho", password: "1234@1" } })).json();
  const patient = await (await api(request, doctorTok.access, "POST", "/patient/", { name: `Sino Teste ${Date.now() % 1e7}` })).json();
  const ref = await (await api(request, doctorTok.access, "POST", "/referrals/", {
    patient_id: patient.id, diagnosis: "Cervicalgia", reason: "Mobilidade",
  })).json();

  await signIn(page, request, "cm.mariana.costa");
  await page.goto("/");
  // (desktop and mobile menus each have a bell; use the visible one and the open panel)
  await page.locator("#notification-slide-btn:visible").first().click();
  const panel = page.getByRole("dialog");
  const item = panel.getByRole("button", { name: new RegExp(patient.name) });
  await expect(item).toBeVisible();
  await item.click();
  await expect(page).toHaveURL(new RegExp(`/referrals/${ref.id}$`));

  // Marked read: it no longer appears under "Por ler"
  await page.locator("#notification-slide-btn:visible").first().click();
  await panel.getByRole("tab", { name: "Por ler" }).click();
  await expect(panel.getByRole("button", { name: new RegExp(patient.name) })).toHaveCount(0);
});

test("the letter prints as one clean A4 page of content", async ({ page, request }) => {
  const tok = await signIn(page, request, "cm.ana.ferreira");
  const refs = await (await api(request, tok, "GET", "/referrals/?status=completed&limit=1")).json();
  test.skip(!refs.results?.length, "no completed referral in the demo data");
  await page.goto(`/referrals/${refs.results[0].id}/letter`);
  await expect(page.getByRole("heading", { name: "Relatório de alta de fisioterapia" }).first()).toBeVisible();

  // In print, only the letter shows: no sidebar, no buttons
  await page.emulateMedia({ media: "print" });
  await expect(page.locator("nav").first()).toBeHidden();
  await expect(page.getByRole("button", { name: /Imprimir/ })).toBeHidden();
  const pdf = await page.pdf({ format: "A4" });
  expect(pdf.byteLength).toBeGreaterThan(5_000);
  expect(pdf.subarray(0, 4).toString()).toBe("%PDF");
});

test.describe("phone screen", () => {
  test.use({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true });

  for (const path of ["/", "/patients", "/referrals", "/time-board", "/inventory", "/users"]) {
    test(`${path} fits the screen`, async ({ page, request }) => {
      await signIn(page, request, "cm.ana.ferreira");
      await page.goto(path);
      await page.waitForLoadState("networkidle");
      // The page itself never scrolls sideways (wide tables scroll inside their box)
      const overflow = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
      expect(overflow, `${path} horizontal overflow`).toBeLessThanOrEqual(1);
      await expect(page.locator("main, #pages").first()).toBeVisible();
    });
  }
});

// Serious and critical WCAG A/AA problems on the main pages
for (const path of ["/", "/patients", "/patients/new", "/referrals", "/referrals/new", "/time-board", "/inventory", "/users", "/user/preferences", "/about"]) {
  test(`accessibility: ${path}`, async ({ page, request }) => {
    await signIn(page, request, "cm.ana.ferreira");
    await page.goto(path);
    await page.waitForLoadState("networkidle");
    const { violations } = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa"]).analyze();
    const bad = violations
      .filter((v) => v.impact === "serious" || v.impact === "critical")
      .map((v) => `${v.id} (${v.impact}): ${v.nodes.length}× e.g. ${v.nodes[0]?.target.join(" ")}`);
    expect(bad, bad.join("\n")).toEqual([]);
  });
}
