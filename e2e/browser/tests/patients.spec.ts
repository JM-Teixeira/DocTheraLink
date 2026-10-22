import { expect, test } from "@playwright/test";

import { PASSWORD, api, signIn } from "./helpers";

// Patient registration extras (draft auto-save, address parts, emergency
// contact) and the patient history.

const stamp = () => Date.now().toString().slice(-7);

test("an unsaved registration survives a reload and can be discarded", async ({ page, request }) => {
  await signIn(page, request, "cm.rececao1");
  const name = `Rascunho ${stamp()}`;
  await page.goto("/patients/new");
  await page.locator("#name").fill(name);
  await page.locator("#postal_code").fill("4000123");
  await page.locator("#postal_code").press("Tab");
  await expect(page.locator("#postal_code")).toHaveValue("4000-123");
  await page.locator("#city").fill("Porto");
  await page.locator("#emergency_contact_name").fill("Rui Costa");
  await page.locator("#emergency_contact_relation").selectOption("spouse");
  await page.waitForTimeout(800); // the draft is saved half a second after typing

  await page.reload();
  await expect(page.getByText(/Recuperámos o formulário por gravar/)).toBeVisible();
  await expect(page.locator("#name")).toHaveValue(name);
  await expect(page.locator("#emergency_contact_relation")).toHaveValue("spouse");

  // Discard: back to an empty form, and nothing comes back on reload
  await page.getByRole("button", { name: "Descartar rascunho" }).click();
  await expect(page.locator("#name")).toHaveValue("");
  await page.waitForTimeout(800);
  await page.reload();
  await expect(page.locator("#name")).toHaveValue("");
  await expect(page.getByText(/Recuperámos/)).toHaveCount(0);

  // Register with the new fields; the patient page shows them, and the draft is gone
  await page.locator("#name").fill(name);
  await page.locator("#email").fill("ana@");
  await page.getByRole("button", { name: "Registar utente" }).click();
  await expect(page.getByText("E-mail inválido")).toBeVisible();
  await page.locator("#email").fill("rascunho@example.pt");
  await page.locator("#postal_code").fill("4000-123");
  await page.locator("#city").fill("Porto");
  await page.locator("#emergency_contact_name").fill("Rui Costa");
  await page.locator("#emergency_contact_relation").selectOption("spouse");
  await page.locator("#insurer").fill("ADSE");
  await page.getByRole("button", { name: "Registar utente" }).click();
  await expect(page).toHaveURL(/\/patients\/[0-9a-f-]{36}$/);
  await expect(page.getByText("4000-123 Porto")).toBeVisible();
  await expect(page.getByText(/Rui Costa \(Cônjuge ou companheiro\(a\)\)/)).toBeVisible();
  await expect(page.getByRole("region", { name: "Histórico do utente" }).getByText("Utente registado")).toBeVisible();
  await page.goto("/patients/new");
  await expect(page.locator("#name")).toHaveValue("");
});

test("the patient history gathers the referral, its messages and sessions", async ({ page, request }) => {
  const doctorTok = await signIn(page, request, "cm.medico2");
  const physioTok = (await (await request.post("/api/v1/auth/login/", { data: { username: "cm.fisio4", password: PASSWORD } })).json()).access;
  const joao = await (await api(request, physioTok, "GET", "/users/getcurrentuser/")).json();
  const patient = await (await api(request, doctorTok, "POST", "/patient/", { name: `Historico ${stamp()}` })).json();
  const ref = await (await api(request, doctorTok, "POST", "/referrals/", {
    patient_id: patient.id, diagnosis: "Epicondilite lateral", reason: "Dor ao agarrar", therapist_id: joao.id,
  })).json();
  await api(request, physioTok, "POST", `/referrals/${ref.id}/accept/`, {});
  await api(request, physioTok, "POST", `/referrals/${ref.id}/notes/`, { kind: "progress", body: "Menos dor ao agarrar", pain: 4 });

  await page.goto(`/patients/${patient.id}`);
  const history = page.getByRole("region", { name: "Histórico do utente" });
  await expect(history.getByText("Referenciado para fisioterapia")).toBeVisible();
  await expect(history.getByText("Epicondilite lateral")).toBeVisible();
  await expect(history.getByText("Nota de evolução")).toBeVisible();
  await expect(history.getByText("Dor 4/10")).toBeVisible();

  // Only the messages
  await history.getByRole("button", { name: "Mensagens (1)" }).click();
  await expect(history.getByText("Nota de evolução")).toBeVisible();
  await expect(history.getByText("Referenciado para fisioterapia")).toHaveCount(0);

  // An entry opens its referral
  await history.getByRole("link", { name: "Nota de evolução" }).click();
  await expect(page).toHaveURL(new RegExp(`/referrals/${ref.id}$`));
});
