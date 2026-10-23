import { expect, test } from "@playwright/test";

import { PASSWORD, api, signIn } from "./helpers";

// Settings follow the person (stored on the account): theme, text size,
// start page; suggestions go to the admins and come back with an answer.

test("settings: dark theme, large text and the start page", async ({ page, request }) => {
  const token = await signIn(page, request, "cm.pedro.almeida");
  const reset = { theme: "light", text_size: "normal", start_page: "home" };
  try {
    await page.goto("/settings");
    const appearance = page.getByRole("region", { name: "Aparência" });
    await appearance.getByRole("radio", { name: "Escuro" }).click();
    await expect(page.locator("html")).toHaveClass(/theme-dark/);
    // the whole page is inverted (colours turned back), pictures are not
    const filter = await page.evaluate(() => getComputedStyle(document.documentElement).filter);
    expect(filter).toContain("invert(0.9)");
    await appearance.getByRole("radio", { name: "Grande" }).click();
    await expect(page.locator("html")).toHaveClass(/text-large/);
    await page.getByRole("region", { name: "Página inicial" }).getByRole("radio", { name: "Quadro horário" }).click();
    await expect(page.getByRole("radio", { name: "Quadro horário" })).toHaveAttribute("aria-checked", "true");

    // A new browser session starts on the chosen page, still dark
    const fresh = await page.context().browser()!.newPage();
    await signIn(fresh, request, "cm.pedro.almeida");
    await fresh.goto("/");
    await expect(fresh).toHaveURL(/\/time-board$/);
    await expect(fresh.locator("html")).toHaveClass(/theme-dark/);
    await fresh.close();
  } finally {
    await api(request, token, "PATCH", "/users/cm.pedro.almeida/preferences/", reset);
  }
});

test("suggestions: a physio sends one, the admin answers, the physio sees it", async ({ browser, request }) => {
  const text = `Imprimir o dia de cada fisioterapeuta ${Date.now()}`;
  const physio = await browser.newPage();
  await signIn(physio, request, "cm.pedro.almeida");
  await physio.goto("/settings");
  const card = physio.getByRole("region", { name: "Sugestões" });
  await card.getByRole("radio", { name: "Ideia" }).click();
  await card.getByLabel("A sua sugestão").fill(text);
  await card.getByRole("button", { name: "Enviar sugestão" }).click();
  const mine = card.getByRole("listitem").filter({ hasText: text });
  await expect(mine).toBeVisible();
  await expect(mine.getByText("Nova", { exact: true })).toBeVisible();

  const admin = await browser.newPage();
  await signIn(admin, request, "admin.user");
  await admin.goto("/suggestions");
  const item = admin.getByRole("listitem").filter({ hasText: text });
  await item.getByLabel("Resposta").fill("Boa ideia, fica planeada.");
  await item.getByRole("button", { name: "Guardar" }).click();
  await item.getByLabel("Estado").selectOption("planned");
  await expect(admin.getByText(text)).toHaveCount(0); // left the "new" filter

  await physio.reload();
  await expect(mine.getByText("Boa ideia, fica planeada.")).toBeVisible();
  await expect(mine.getByText("Planeada", { exact: true })).toBeVisible();

  // Nobody else reads all suggestions
  const tok = (await (await request.post("/api/v1/auth/login/", { data: { username: "cm.pedro.almeida", password: PASSWORD } })).json()).access;
  expect((await api(request, tok, "GET", "/suggestions/")).status()).toBe(403);
});
