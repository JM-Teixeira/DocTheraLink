import { expect, test } from "@playwright/test";

import { api, signIn } from "./helpers";

// The core journey, as people do it in the app (Climutua demo data):
// the doctor refers a patient to a physio; the physio accepts, starts,
// writes a progress note the doctor reads, and completes with an outcome;
// the discharge letter carries it. Another clinic sees none of it.
test("referral journey: refer, accept, treat, talk, complete, letter", async ({ browser, request }) => {
  const stamp = Date.now().toString().slice(-7);
  const patientName = `Navegador Teste ${stamp}`;
  const note = `Boa evolução, dor 3/10 (${stamp})`;
  const outcome = `Objetivos atingidos, alta com exercícios em casa (${stamp})`;

  // A fresh patient, registered by the front desk
  const desk = await browser.newPage();
  const deskToken = await signIn(desk, request, "cm.rececao1");
  const created = await api(request, deskToken, "POST", "/patient/", { name: patientName });
  expect(created.status()).toBe(201);
  await desk.close();

  // 1. The doctor refers the patient to Pedro Almeida
  const doctor = await browser.newPage();
  await signIn(doctor, request, "cm.medico1");
  await doctor.goto("/referrals/new");
  await doctor.getByLabel("Pesquisar utente").fill(patientName);
  await doctor.getByRole("button", { name: new RegExp(patientName) }).click();
  await doctor.locator("#diagnosis").fill("Lombalgia mecânica crónica");
  await doctor.locator("#reason").fill("Dor lombar há 4 meses; fortalecimento e reeducação postural.");
  await doctor.locator("#therapist").selectOption({ label: "Pedro Almeida" });
  await doctor.getByRole("button", { name: "Enviar encaminhamento" }).click();
  await expect(doctor).toHaveURL(/\/referrals\/[0-9a-f-]{36}$/);
  const referralUrl = new URL(doctor.url()).pathname;
  await expect(doctor.getByText("Lombalgia mecânica crónica").first()).toBeVisible();

  // 2. The physio accepts, starts and writes a progress note
  const physio = await browser.newPage();
  await signIn(physio, request, "cm.fisio2");
  await physio.goto(referralUrl);
  await physio.getByRole("button", { name: "Aceitar encaminhamento" }).click();
  await physio.getByRole("button", { name: "Iniciar tratamento" }).click();
  await expect(physio.getByRole("button", { name: "Concluir tratamento" })).toBeVisible();
  await physio.getByPlaceholder(/Escreva uma mensagem/).fill(note);
  await physio.getByRole("button", { name: "Enviar" }).click();
  await expect(physio.getByText(note)).toBeVisible();

  // 3. The doctor reads it on the referral
  await doctor.reload();
  await expect(doctor.getByText(note)).toBeVisible();
  await expect(doctor.getByRole("button", { name: "Aceitar encaminhamento" })).toHaveCount(0);

  // 4. The physio completes with the outcome
  await physio.getByRole("button", { name: "Concluir tratamento" }).click();
  const dialog = physio.getByRole("dialog");
  await dialog.locator("textarea").fill(outcome);
  await dialog.getByRole("button", { name: "Concluir tratamento" }).click();
  await expect(dialog).toHaveCount(0);
  // (shown in the outcome box and in the thread's "completed" entry)
  await expect(physio.getByText(outcome).first()).toBeVisible();
  await expect(physio.getByRole("button", { name: "Concluir tratamento" })).toHaveCount(0);

  // 5. The discharge letter has the patient and the outcome
  await doctor.goto(`${referralUrl}/letter`);
  // (the sheet is rendered twice: on screen and in the print copy)
  await expect(doctor.getByRole("heading", { name: "Relatório de alta de fisioterapia" }).first()).toBeVisible();
  await expect(doctor.getByText(patientName).first()).toBeVisible();
  await expect(doctor.getByText(outcome).first()).toBeVisible();

  // 6. Nobody at the other clinic can open it
  const outsider = await browser.newPage();
  const outsiderToken = await signIn(outsider, request, "physiotherapist");
  const id = referralUrl.split("/").pop();
  expect((await api(request, outsiderToken, "GET", `/referrals/${id}/`)).status()).toBe(404);
  await outsider.close();
  await doctor.close();
  await physio.close();
});

test("front desk cannot open referrals", async ({ page, request }) => {
  const token = await signIn(page, request, "cm.rececao1");
  expect((await api(request, token, "GET", "/referrals/")).status()).toBe(403);
});
