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
  await page.getByRole("button", { name: "Adicionar pessoa", exact: true }).click();
  await expect(page).toHaveURL(new RegExp(`/users/cm\.joao\.${last.toLowerCase()}$`));
});

test("the ? next to the title explains the page in a pop-up", async ({ page, request }) => {
  await signIn(page, request, "cm.ana.ferreira");
  await page.goto("/inventory");
  await expect(page.getByText(/Os artigos descartáveis/)).toHaveCount(0); // no longer on the page
  await page.getByRole("button", { name: "Ajuda: Inventário" }).click();
  const help = page.getByRole("dialog");
  await expect(help.getByRole("heading", { name: "Inventário" })).toBeVisible();
  await expect(help.getByRole("listitem").first()).toBeVisible();
  await expect(help.getByText(/descartáveis/)).toBeVisible();
});

test("an admin views the app as a physio of Climutua, read-only, and comes back", async ({ page, request }) => {
  await signIn(page, request, "paula.neves");
  await page.goto("/");
  await expect(page.getByRole("link", { name: "Sugestões" })).toBeVisible(); // admin menu
  await page.locator("#user-profile-name").click();
  await page.getByRole("menuitem", { name: "Ver como…" }).click();
  const dialog = page.getByRole("dialog");
  await dialog.getByLabel("Função").selectOption("PhysioTherapist");
  await dialog.getByLabel("Clínica").selectOption({ label: "Climutua · Medicina Física e Reabilitação" });
  await dialog.getByRole("button", { name: "Ver", exact: true }).click();

  const banner = page.getByRole("status").filter({ hasText: "A ver como Fisioterapeuta" });
  await expect(banner).toContainText("só leitura");
  await expect(page.getByRole("link", { name: "Sugestões" })).toHaveCount(0); // a physio has no admin menu
  await page.goto("/patients");
  await expect(page.locator("tbody tr").first()).toBeVisible(); // Climutua's patients
  await expect(page.getByRole("button", { name: "Exportar CSV" })).toHaveCount(0);

  await banner.getByRole("button", { name: "Voltar a administrador" }).click();
  await expect(page.getByRole("status").filter({ hasText: "A ver como" })).toHaveCount(0);
  await expect(page.getByRole("link", { name: "Sugestões" })).toBeVisible();
});

test("people: who can change gloves on floor 1, and the work history at a glance", async ({ page, request }) => {
  await signIn(page, request, "cm.ana.ferreira");
  await page.goto("/users");
  await page.getByLabel("Pode alterar stock do piso").selectOption("1");
  const rows = page.locator("tbody");
  // Teresa (all stock), Inês (floor 1) and the coordinator; not Mariana (floors 2-3 hygiene) or Pedro
  await expect(rows.getByText("Teresa Gomes")).toBeVisible();
  await expect(rows.getByText("Inês Moreira")).toBeVisible();
  await expect(rows.getByText("Ana Ferreira")).toBeVisible();
  await expect(rows.getByText("Mariana Costa")).toHaveCount(0);
  await expect(rows.getByText("Pedro Almeida")).toHaveCount(0);

  await page.goto("/users/cm.rui.nogueira");
  const history = page.getByRole("list", { name: "Situação profissional" });
  await expect(history.getByText(/^Entrou a /)).toBeVisible();
  await expect(history.getByText("Saiu a 31/07/2026")).toBeVisible();
});
