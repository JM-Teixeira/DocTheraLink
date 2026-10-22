import { APIRequestContext, Page, expect, test } from "@playwright/test";

import { PASSWORD, api, signIn } from "./helpers";

// Moving sessions by drag and drop, and the waiting list, on the demo
// clinic's board in a week of its own far ahead (made and removed here).

async function token(request: APIRequestContext, username: string) {
  const res = await request.post("/api/v1/auth/login/", { data: { username, password: PASSWORD } });
  return (await res.json()).access as string;
}
const iso = (d: Date) => d.toISOString().slice(0, 10);
const dmy = (date: string) => date.split("-").reverse().join("/");
const stamp = () => Date.now().toString().slice(-7);

/** A Monday (far ahead) where the physio has no shift, with shifts 9-12 that week. */
async function freeWeek(request: APIRequestContext, coordTok: string, physioId: number, division: string) {
  const d = new Date(Date.UTC(2080, 0, 1)); // a Monday
  d.setUTCDate(d.getUTCDate() + 7 * (Number(stamp()) % 500));
  for (;;) {
    const board = await (await api(request, coordTok, "GET", `/schedule/board/?date=${iso(d)}`)).json();
    const row = board.therapists.find((r: { id: number }) => r.id === physioId);
    if (!row?.shifts.length) break;
    d.setUTCDate(d.getUTCDate() + 7);
  }
  const shifts: string[] = [];
  for (let i = 0; i < 5; i++) {
    const day = new Date(d);
    day.setUTCDate(day.getUTCDate() + i);
    const r = await api(request, coordTok, "POST", "/schedule/shifts/", {
      date: iso(day), start_hour: 9, end_hour: 12, therapist_id: physioId, division_id: division,
    });
    if (r.status() === 201) shifts.push((await r.json()).id);
  }
  return { monday: iso(d), shifts };
}

async function cleanup(request: APIRequestContext, coordTok: string, monday: string, physioId: number, shifts: string[]) {
  for (let i = 0; i < 5; i++) {
    const d = new Date(`${monday}T00:00:00Z`);
    d.setUTCDate(d.getUTCDate() + i);
    const board = await (await api(request, coordTok, "GET", `/schedule/board/?date=${iso(d)}`)).json();
    for (const r of board.therapists)
      if (r.id === physioId)
        for (const h of r.hours) for (const a of h.appointments) await api(request, coordTok, "DELETE", `/schedule/appointments/${a.id}/`);
  }
  for (const id of shifts) await api(request, coordTok, "DELETE", `/schedule/shifts/${id}/`);
}

async function boardOn(page: Page, date: string) {
  await page.goto("/time-board");
  const field = page.getByLabel("Data", { exact: true });
  await field.fill(dmy(date));
  await field.press("Enter");
  await expect(field).toHaveValue(dmy(date));
}

test("drag a patient to another hour in the physio's day", async ({ page, request }) => {
  const coordTok = await signIn(page, request, "de.marta.ribeiro");
  const tiago = await (await api(request, await token(request, "de.tiago.martins"), "GET", "/users/getcurrentuser/")).json();
  const divs = (await (await api(request, coordTok, "GET", "/schedule/divisions/")).json()) as { id: string; capacity: number }[];
  const gym = divs.find((x) => x.capacity >= 2)!.id;
  const { monday, shifts } = await freeWeek(request, coordTok, tiago.id, gym);
  const patient = await (await api(request, await token(request, "de.rita.lopes"), "POST", "/patient/", { name: `Arrastar ${stamp()}` })).json();
  try {
    expect((await api(request, coordTok, "POST", "/schedule/appointments/", { date: monday, hour: 9, therapist_id: tiago.id, patient_id: patient.id })).status()).toBe(201);

    await boardOn(page, monday);
    await page.locator("[data-physio-row]", { hasText: `${tiago.first_name} ${tiago.last_name}` }).locator("button").first().click();
    const dialog = page.getByRole("dialog");
    // The patient chip (not the "move" button next to it)
    const chip = dialog.locator("button[draggable=true]", { hasText: patient.name });
    await chip.dragTo(dialog.locator('li[data-hour="11"]'));
    await expect(page.getByText("Sessão mudada").first()).toBeVisible();

    const board = await (await api(request, coordTok, "GET", `/schedule/board/?date=${monday}`)).json();
    const row = board.therapists.find((r: { id: number }) => r.id === tiago.id);
    const at = row.hours.find((h: { appointments: { patient_id: string }[] }) => h.appointments.some((a) => a.patient_id === patient.id));
    expect(at.hour).toBe(11);

    // The keyboard way: "move" button and a dialog
    await dialog.getByRole("button", { name: `Mudar a sessão de ${patient.name}` }).click();
    const move = page.getByRole("dialog", { name: "Mudar sessão" });
    await move.getByLabel("Hora").selectOption({ label: "10:00" });
    await move.getByRole("button", { name: "Mudar sessão" }).click();
    await expect(move).toHaveCount(0);
  } finally {
    await cleanup(request, coordTok, monday, tiago.id, shifts);
  }
});

test("waiting list: add a patient, book their plan, they leave the list", async ({ page, request }) => {
  const coordTok = await signIn(page, request, "de.marta.ribeiro");
  const tiago = await (await api(request, await token(request, "de.tiago.martins"), "GET", "/users/getcurrentuser/")).json();
  const divs = (await (await api(request, coordTok, "GET", "/schedule/divisions/")).json()) as { id: string; capacity: number }[];
  const gym = divs.find((x) => x.capacity >= 2)!.id;
  const { monday, shifts } = await freeWeek(request, coordTok, tiago.id, gym);
  const name = `Lista Espera ${stamp()}`;
  await api(request, await token(request, "de.rita.lopes"), "POST", "/patient/", { name });
  try {
    await boardOn(page, monday);
    await page.getByRole("button", { name: /Lista de espera/ }).click();
    const dialog = page.getByRole("dialog", { name: "Lista de espera" });
    await dialog.getByRole("button", { name: "Adicionar à lista de espera" }).first().click();
    await dialog.getByLabel("Utente").fill(name);
    await dialog.getByRole("button", { name: new RegExp(name) }).click();
    // Tuesdays and Thursdays
    for (const day of [/^ter/i, /^qui/i]) await dialog.getByRole("group", { name: "Dias" }).getByRole("button", { name: day }).click();
    await dialog.getByLabel(/Urgente/).check();
    await dialog.locator("form").getByRole("button", { name: "Adicionar à lista de espera" }).click();
    const row = dialog.getByRole("listitem").filter({ hasText: name });
    await expect(row).toContainText("Urgente");
    await expect(row).toContainText("à espera há 0 dias");

    // Book the plan from the list
    await row.getByRole("button", { name: "Marcar plano" }).click();
    const plan = page.getByRole("dialog", { name: "Plano semanal" });
    await expect(plan).toContainText(name);
    await plan.getByLabel("Fisioterapeuta").selectOption({ label: `${tiago.first_name} ${tiago.last_name}` });
    await plan.getByLabel("Hora").selectOption({ label: "09:00" });
    await plan.getByLabel("Sessões").fill("2");
    await plan.getByRole("button", { name: "Pré-visualizar" }).click();
    await plan.getByRole("button", { name: /^Marcar 2 sessões$/ }).click();
    await expect(plan).toHaveCount(0);

    await page.getByRole("button", { name: /Lista de espera/ }).click();
    await expect(page.getByRole("dialog", { name: "Lista de espera" }).getByText(name)).toHaveCount(0);
  } finally {
    await cleanup(request, coordTok, monday, tiago.id, shifts);
    const list = (await (await api(request, coordTok, "GET", "/schedule/waiting/")).json()) as { id: string; patient?: { name: string } }[];
    for (const e of list) if (e.patient?.name === name) await api(request, coordTok, "PATCH", `/schedule/waiting/${e.id}/`, { status: "removed" });
  }
});

test("waiting list: an entry with a free slot stands out and books into it", async ({ page, request }) => {
  const coordTok = await signIn(page, request, "cm.ana.ferreira");
  const patient = await (await api(request, coordTok, "POST", "/patient/", { name: `Vaga Livre ${stamp()}` })).json();
  const entry = await (await api(request, coordTok, "POST", "/schedule/waiting/", { patient_id: patient.id, period: "any" })).json();
  try {
    await page.goto("/time-board");
    await page.getByRole("button", { name: /Lista de espera/ }).click();
    const item = page.getByRole("dialog").getByRole("listitem").filter({ hasText: patient.name });
    await expect(item.getByText(/^Vaga disponível: /)).toBeVisible();
    await item.getByRole("button", { name: "Marcar plano" }).click();
    await expect(page.getByRole("dialog").getByText(patient.name).first()).toBeVisible();
  } finally {
    await api(request, coordTok, "PATCH", `/schedule/waiting/${entry.id}/`, { status: "removed" });
  }
});
