import { expect, test } from "@playwright/test";

import { PASSWORD, api, signIn, visibleText } from "./helpers";

// Runs against the demo data of the local stack (Climutua, see README).

test("login form signs in and refuses a wrong password", async ({ page }) => {
  await page.goto("/");
  await page.locator("#username").fill("cm.ana.ferreira");
  await page.locator("input[name=password]").first().fill("wrong-password");
  await page.locator("button[type=submit]").first().click();
  await expect(page).toHaveURL(/login|\/$/);
  await expect(page.locator("#username")).toBeVisible();

  await page.locator("input[name=password]").first().fill(PASSWORD);
  await page.locator("button[type=submit]").first().click();
  await expect(page.getByRole("link", { name: /Quadro horário|Time board/ }).first()).toBeVisible();
});

test("a person who left cannot sign in", async ({ request }) => {
  const res = await request.post("/api/v1/auth/login/", { data: { username: "cm.rui.nogueira", password: PASSWORD } });
  expect(res.ok()).toBeFalsy();
});

test("people: status chips find who left the clinic", async ({ page, request }) => {
  await signIn(page, request, "cm.ana.ferreira");
  await page.goto("/users");
  await expect(page.getByRole("heading", { name: "Pessoas" })).toBeVisible();
  await expect(page.getByText("Rui Nogueira")).toHaveCount(0); // active by default
  await page.getByRole("button", { name: "Saiu", exact: true }).click();
  await expect(page.getByText("Rui Nogueira")).toBeVisible();
  await page.getByText("Rui Nogueira").click();
  await expect(page).toHaveURL(/\/users\/cm\.rui\.nogueira/);
});

test("patients: filters narrow the list and the metrics agree", async ({ page, request }) => {
  await signIn(page, request, "cm.ana.ferreira");
  await page.goto("/patients");
  await page.getByRole("button", { name: "65-79 anos" }).click();
  await page.getByRole("button", { name: "Nível de cuidados 4" }).click();
  const metrics = page.getByRole("region", { name: "Métricas dos utentes" });
  await expect(metrics).toBeVisible();
  const total = Number(await metrics.locator("p.text-2xl").innerText());
  const rows = page.locator("tbody tr");
  await expect(rows).toHaveCount(Math.min(total, 20));
  for (const ages of await rows.locator("td:nth-child(2)").allInnerTexts()) {
    expect(Number(ages)).toBeGreaterThanOrEqual(65);
    expect(Number(ages)).toBeLessThanOrEqual(79);
  }
});

test("dates follow the person's preference", async ({ page, request }) => {
  const token = await signIn(page, request, "cm.ana.ferreira");
  try {
    expect((await api(request, token, "PATCH", "/users/cm.ana.ferreira/preferences/", { date_format: "ymd" })).ok()).toBeTruthy();
    await page.goto("/time-board");
    const today = new Date().toISOString().slice(0, 10);
    await expect(page.getByLabel("Data", { exact: true })).toHaveValue(today);

    // Change it from the profile, as a person would
    await page.goto("/settings");
    await page.getByRole("radio", { name: /Dia \/ mês \/ ano/ }).click();
    await expect(page.getByRole("radio", { name: /Dia \/ mês \/ ano/ })).toHaveAttribute("aria-checked", "true");
    await page.goto("/time-board");
    const [y, m, d] = today.split("-");
    await expect(page.getByLabel("Data", { exact: true })).toHaveValue(`${d}/${m}/${y}`);
  } finally {
    await api(request, token, "PATCH", "/users/cm.ana.ferreira/preferences/", { date_format: "dmy" });
  }
});

test("dropdown arrows never sit on top of the text", async ({ page, request }) => {
  await signIn(page, request, "cm.ana.ferreira");
  for (const path of ["/referrals", "/users", "/patients"]) {
    await page.goto(path);
    await page.waitForLoadState("networkidle");
    for (const select of await page.locator("select").all()) {
      const padRight = await select.evaluate((el) => parseFloat(getComputedStyle(el).paddingRight));
      expect(padRight, `${path} select padding`).toBeGreaterThanOrEqual(32);
    }
  }
});

test("no long dashes in the interface", async ({ page, request }) => {
  await signIn(page, request, "cm.ana.ferreira");
  for (const path of ["/", "/patients", "/users", "/referrals", "/time-board", "/inventory", "/users/cm.rui.nogueira/facilities"]) {
    await page.goto(path);
    await page.waitForLoadState("networkidle");
    expect(await visibleText(page), path).not.toMatch(/[—–]/);
  }
});

test("another clinic's people stay invisible in the app", async ({ page, request }) => {
  const token = await signIn(page, request, "de.marta.ribeiro", "en");
  await page.goto("/users");
  await page.getByRole("searchbox").fill("cm.");
  await expect(page.getByText(/Nobody matches/)).toBeVisible();
  expect((await api(request, token, "GET", "/users/cm.catarina.lopes/")).status()).toBe(404);
  await page.goto("/users/cm.catarina.lopes");
  await expect(page.getByText("Catarina Lopes")).toHaveCount(0);
});

test("time board: filter by floor and open a physio's day", async ({ page, request }) => {
  await signIn(page, request, "cm.ana.ferreira");
  await page.goto("/time-board");
  await page.getByRole("button", { name: "Piso 3", exact: true }).click();
  // Floor 3 is neuro rehab only; the gym and hydrotherapy (floor 1) go away
  await expect(page.getByRole("button", { name: /Reabilitação Neurológica/ }).first()).toBeVisible();
  await expect(page.getByRole("button", { name: /Hidroterapia/ })).toHaveCount(0);
  await page.getByRole("button", { name: "Todos", exact: true }).first().click();
  // The physio's name in the grid (not the filter chip) opens their day
  const row = page.locator("[data-physio-row]").first();
  const name = (await row.locator("button").first().innerText()).split("\n")[0];
  await row.locator("button").first().click();
  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  await expect(dialog.getByRole("heading", { name })).toBeVisible();
});
