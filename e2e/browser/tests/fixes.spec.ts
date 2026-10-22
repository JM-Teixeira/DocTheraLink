import { expect, test } from "@playwright/test";

import { signIn } from "./helpers";

// Fixes from the 2026-10-22 review: admin stock items, login layout, spacing.

const stamp = () => Date.now().toString().slice(-7);

test("a state admin adds a stock item by choosing the clinic in the dialog", async ({ page, request }) => {
  await signIn(page, request, "paula.neves");
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
  await signIn(page, request, "paula.neves");
  await page.goto("/facility");
  const button = await page.getByRole("button", { name: "Nova clínica" }).boundingBox();
  const search = await page.getByPlaceholder("Nome ou morada").boundingBox();
  expect(button && search && search.y - (button.y + button.height)).toBeGreaterThanOrEqual(24);
});

test("the coordinator adds a person; the username follows the clinic standard", async ({ page, request }) => {
  await signIn(page, request, "cm.ana.ferreira");
  page.on("pageerror", (e) => {
    throw e;
  });
  await page.goto("/users/add");
  const last = `Teste${stamp().replace(/\d/g, (d) => "abcdefghij"[Number(d)])}`;
  await page.getByLabel(/^Nome próprio/).fill("João");
  await page.getByLabel(/^Apelido/).fill(`Silva ${last}`);
  await expect(page.getByText(`cm.joao.${last.toLowerCase()}`)).toBeVisible();
  await page.getByLabel(/^Email/).fill(`joao.${last.toLowerCase()}@example.test`);
  await page.getByLabel(/^Palavra-passe/).fill("Seguro#2026");
  await page.getByLabel(/^Confirmar palavra-passe/i).fill("Seguro#2026");
  await page.getByRole("button", { name: "Adicionar pessoa" }).click();
  await expect(page).toHaveURL(new RegExp(`/users/cm\.joao\.${last.toLowerCase()}$`));
});
