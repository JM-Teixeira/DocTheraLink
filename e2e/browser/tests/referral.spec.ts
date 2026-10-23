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
  const deskToken = await signIn(desk, request, "cm.sofia.martins");
  const created = await api(request, deskToken, "POST", "/patient/", { name: patientName });
  expect(created.status()).toBe(201);
  await desk.close();

  // 1. The doctor refers the patient to Pedro Almeida
  const doctor = await browser.newPage();
  await signIn(doctor, request, "cm.miguel.sousa");
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
  await signIn(physio, request, "cm.pedro.almeida");
  await physio.goto(referralUrl);
  await physio.getByRole("button", { name: "Aceitar encaminhamento" }).click();
  await physio.getByRole("button", { name: "Iniciar tratamento" }).click();
  await expect(physio.getByRole("button", { name: "Concluir tratamento" })).toBeVisible();
  await physio.getByLabel("Mensagem", { exact: true }).fill(note);
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
  const outsiderToken = await signIn(outsider, request, "de.sarah.johnson");
  const id = referralUrl.split("/").pop();
  expect((await api(request, outsiderToken, "GET", `/referrals/${id}/`)).status()).toBe(404);
  await outsider.close();
  await doctor.close();
  await physio.close();
});

test("front desk cannot open referrals", async ({ page, request }) => {
  const token = await signIn(page, request, "cm.sofia.martins");
  expect((await api(request, token, "GET", "/referrals/")).status()).toBe(403);
});

// The conversation: the doctor asks, the physio sees "1 por responder" on the
// card and with the filter, answers that question; progress notes with pain
// draw the evolution.
test("questions wait for an answer; progress notes draw the pain", async ({ browser, request }) => {
  const stamp = Date.now().toString().slice(-7);
  const doctor = await browser.newPage();
  const doctorTok = await signIn(doctor, request, "cm.helena.carvalho");
  const physioTok = await (await request.post("/api/v1/auth/login/", { data: { username: "cm.joao.ribeiro", password: "1234@1" } })).json();
  const joao = await (await api(request, physioTok.access, "GET", "/users/getcurrentuser/")).json();
  const patient = await (await api(request, doctorTok, "POST", "/patient/", { name: `Conversa ${stamp}` })).json();
  const ref = await (await api(request, doctorTok, "POST", "/referrals/", {
    patient_id: patient.id, diagnosis: "Tendinopatia rotuliana", reason: "Carga progressiva", therapist_id: joao.id,
  })).json();
  await api(request, physioTok.access, "POST", `/referrals/${ref.id}/accept/`, {});
  const question = `Está a tolerar os agachamentos? (${stamp})`;

  await doctor.goto(`/referrals/${ref.id}`);
  await doctor.getByRole("radio", { name: "Pergunta" }).click();
  await doctor.getByLabel("Mensagem", { exact: true }).fill(question);
  await doctor.getByRole("button", { name: "Enviar" }).click();
  await expect(doctor.getByText(question)).toBeVisible();
  await expect(doctor.getByText("Por responder").first()).toBeVisible();

  const physio = await browser.newPage();
  await signIn(physio, request, "cm.joao.ribeiro");
  await physio.goto("/referrals");
  await physio.getByRole("button", { name: "Por responder" }).click();
  const card = physio.getByRole("link", { name: new RegExp(patient.name) });
  await expect(card).toContainText("1 por responder");
  await card.click();
  await expect(physio.getByText(/1 pergunta está à espera da sua resposta/)).toBeVisible();
  await physio.getByRole("button", { name: "Responder" }).click();
  await physio.getByLabel("Mensagem", { exact: true }).fill("Sim, sem dor no fim.");
  await physio.getByRole("button", { name: "Enviar" }).click();
  await expect(physio.getByText(/Em resposta a:/)).toBeVisible();
  await expect(physio.getByText("Respondida").first()).toBeVisible();

  for (const [pain, text] of [["6", "Primeira semana"], ["3", "Segunda semana"]] as const) {
    await physio.getByRole("radio", { name: "Progresso" }).click();
    await physio.getByLabel("Dor (0-10)").selectOption(pain);
    await physio.getByLabel("Mensagem", { exact: true }).fill(`${text} ${stamp}`);
    await physio.getByRole("button", { name: "Enviar" }).click();
    await expect(physio.getByText(`${text} ${stamp}`)).toBeVisible();
  }
  await expect(physio.getByRole("img", { name: "A dor passou de 6 para 3 em 10" })).toBeVisible();

  // The doctor sees the answer and the evolution; nothing waits any more
  await doctor.reload();
  await expect(doctor.getByText("Respondida").first()).toBeVisible();
  await expect(doctor.getByText("de 6 para 3 em 10")).toBeVisible();
  await doctor.close();
  await physio.close();
});

// The treatment plan: the physio writes the steps of a session; the patient
// page shows them so a colleague can carry on during a holiday.
test("the physio writes the treatment plan and the patient page shows it", async ({ browser, request }) => {
  const stamp = Date.now().toString().slice(-7);
  const physio = await browser.newPage();
  const physioTok = await signIn(physio, request, "cm.joao.ribeiro");
  const joao = await (await api(request, physioTok, "GET", "/users/getcurrentuser/")).json();
  const doctorTok = (await (await request.post("/api/v1/auth/login/", { data: { username: "cm.helena.carvalho", password: "1234@1" } })).json()).access;
  const deskTok = (await (await request.post("/api/v1/auth/login/", { data: { username: "cm.sofia.martins", password: "1234@1" } })).json()).access;
  const patient = await (await api(request, deskTok, "POST", "/patient/", { name: `Plano Ombro ${stamp}` })).json();
  const ref = await (await api(request, doctorTok, "POST", "/referrals/", {
    patient_id: patient.id, diagnosis: "Tendinopatia do ombro", reason: "Dor", therapist_id: joao.id,
  })).json();
  await api(request, physioTok, "POST", `/referrals/${ref.id}/accept/`, {});

  await physio.goto(`/referrals/${ref.id}`);
  const card = physio.getByRole("region", { name: "Plano de tratamento" });
  await card.getByRole("button", { name: "Escrever o plano" }).click();
  await card.getByLabel("Atividade 1").fill("Ultrassons no ombro");
  await card.getByLabel("Min. 1").fill("10");
  await card.getByLabel("Máquina ou sala 1").fill("Máquina 2");
  await card.getByRole("button", { name: "Adicionar passo" }).click();
  await card.getByLabel("Atividade 2").fill("Massagem");
  await card.getByLabel("Min. 2").fill("5");
  await expect(card.getByText("Total: 15 minutos")).toBeVisible();
  await card.getByRole("button", { name: "Guardar" }).click();
  await expect(card.getByText("Ultrassons no ombro")).toBeVisible();
  await expect(card.getByText(/Alterado por João Ribeiro/)).toBeVisible();

  // The coordinator (planning a holiday) sees it on the patient page
  const coord = await browser.newPage();
  await signIn(coord, request, "cm.ana.ferreira");
  await coord.goto(`/patients/${patient.id}`);
  const plan = coord.getByRole("region", { name: "Plano de tratamento" });
  await expect(plan.getByText("Ultrassons no ombro")).toBeVisible();
  await expect(plan.getByText("Massagem")).toBeVisible();
});
