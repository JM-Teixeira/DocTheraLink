import { expect, test } from "@playwright/test";

import { signIn } from "./helpers";

// Fixes from the 2026-10-22 review: admin stock items, login layout, spacing.

const stamp = () => Date.now().toString().slice(-7);

test("a state admin adds a stock item by choosing the clinic in the dialog", async ({ page, request }) => {
  await signIn(page, request, "stateadmin");
  await page.goto("/inventory");
  await page.getByRole("button", { name: "Novo artigo" }).click();
  const dialog = page.getByRole("dialog");
  const save = dialog.getByRole("button", { name: "Guardar" });
  const name = `Luvas teste ${stamp()}`;
  await dialog.getByLabel(/^Nome/).fill(name);
  await expect(save).toBeDisabled(); // no clinic yet
  await dialog.getByLabel(/^Clínica/).selectOption({ index: 1 });
  await save.click();
  await expect(dialog).toBeHidden();
  await page.getByLabel("Pesquisar artigos").fill(name);
  await expect(page.getByText(name)).toBeVisible();
});

test("the login title stays at the same height in Portuguese and English", async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const titleY = async (lang: string) => {
    await page.addInitScript((l) => localStorage.setItem("i18nextLng", l), lang);
    await page.goto("/");
    const box = await page.getByRole("heading", { level: 1, name: "Klyniva" }).boundingBox();
    return box?.y;
  };
  const pt = await titleY("pt");
  const en = await titleY("en");
  expect(pt).toBeDefined();
  expect(en).toBe(pt);
});

test("page headers keep a gap above the first card (clinics)", async ({ page, request }) => {
  await signIn(page, request, "stateadmin");
  await page.goto("/facility");
  const button = await page.getByRole("button", { name: "Nova clínica" }).boundingBox();
  const search = await page.getByPlaceholder("Nome ou morada").boundingBox();
  expect(button && search && search.y - (button.y + button.height)).toBeGreaterThanOrEqual(24);
});
