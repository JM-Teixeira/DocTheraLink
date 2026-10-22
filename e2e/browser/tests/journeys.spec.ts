import { APIRequestContext, Page, expect, test } from "@playwright/test";

import { PASSWORD, api, signIn } from "./helpers";

// Main journeys of the clinic in the browser. Each test makes its own data
// (patients, shifts, items) through the API and removes what it can, so runs
// don't collide.

const iso = (d: Date) => d.toISOString().slice(0, 10);
const addDays = (date: string, n: number) => {
  const d = new Date(`${date}T00:00:00Z`);
  d.setUTCDate(d.getUTCDate() + n);
  return iso(d);
};
/** "2026-10-21" → "21/10/2026" (the demo users' date format). */
const dmy = (date: string) => date.split("-").reverse().join("/");
const stamp = () => Date.now().toString().slice(-7);

/** A valid Portuguese NIF (mod-11 check digit) unique to this run. */
function freshNif() {
  const base = `2${String(Date.now() % 10_000_000).padStart(7, "0")}`;
  const sum = [...base].reduce((s, c, i) => s + Number(c) * (9 - i), 0);
  const check = 11 - (sum % 11);
  return base + (check >= 10 ? 0 : check);
}

async function token(request: APIRequestContext, username: string) {
  const res = await request.post("/api/v1/auth/login/", { data: { username, password: PASSWORD } });
  return (await res.json()).access as string;
}

async function me(request: APIRequestContext, tok: string) {
  return (await (await api(request, tok, "GET", "/users/getcurrentuser/")).json()) as {
    id: number;
    first_name: string;
    last_name: string;
  };
}

/** Opens the time board on a date (typed in the date field). */
async function boardOn(page: Page, date: string) {
  await page.goto("/time-board");
  const field = page.getByLabel("Data", { exact: true });
  await field.fill(dmy(date));
  await field.press("Enter");
  await expect(field).toHaveValue(dmy(date));
}

type BoardRow = { id: number; hours: { hour: number; appointments: { id: string; patient_id: string }[] }[] };
async function boardRows(request: APIRequestContext, tok: string, date: string) {
  const res = await api(request, tok, "GET", `/schedule/board/?date=${date}`);
  return ((await res.json()).therapists ?? []) as BoardRow[];
}
const sessionsOf = (rows: BoardRow[], patientId: string) =>
  rows.flatMap((r) => r.hours.flatMap((h) => h.appointments.filter((a) => a.patient_id === patientId).map((a) => ({ ...a, therapist: r.id, hour: h.hour }))));

async function divisionWithRoom(request: APIRequestContext, tok: string, room = 2) {
  const divs = (await (await api(request, tok, "GET", "/schedule/divisions/")).json()) as { id: string; capacity: number }[];
  const d = divs.find((x) => x.capacity >= room);
  expect(d, "a division with room").toBeTruthy();
  return d!.id;
}

test("register a patient; duplicates are caught; found by NIF", async ({ page, request }) => {
  await signIn(page, request, "cm.rececao1");
  const name = `Registo Teste ${stamp()}`;
  const nif = freshNif();
  const fill = async (withNif?: string) => {
    await page.goto("/patients/new");
    await page.locator("#name").fill(name);
    await page.locator("#date_of_birth").fill("15/03/1961");
    await page.locator("#date_of_birth").press("Tab");
    await page.locator("label", { hasText: /^Feminino$/ }).click();
    // always set: the form keeps an unsaved draft between visits
    await page.locator("#nif").fill(withNif ?? "");
    await page.getByRole("button", { name: "Registar utente" }).click();
  };

  await fill(nif);
  await expect(page).toHaveURL(/\/patients\/[0-9a-f-]{36}$/);
  const firstUrl = page.url();
  await expect(page.getByText(name).first()).toBeVisible();

  // Same NIF: blocked, no way to register anyway
  await fill(nif);
  const dialog = page.getByRole("dialog");
  await expect(dialog.getByText("Este utente já está registado")).toBeVisible();
  await expect(dialog.getByRole("button", { name: /É outra pessoa/ })).toHaveCount(0);
  await dialog.getByRole("button", { name: "Voltar ao formulário" }).click();

  // Same name and birth date, no NIF: a warning, and "it's someone else" works
  await fill();
  await expect(dialog.getByText("Este utente pode já estar registado")).toBeVisible();
  await expect(dialog.getByRole("button", { name: "Abrir" }).first()).toBeVisible();
  await dialog.getByRole("button", { name: /É outra pessoa/ }).click();
  await expect(page).toHaveURL(/\/patients\/[0-9a-f-]{36}$/);
  expect(page.url()).not.toBe(firstUrl);

  // The NIF finds exactly the first one
  await page.goto("/patients");
  await page.getByRole("searchbox").fill(nif);
  await expect(page.locator("tbody tr")).toHaveCount(1);
  await expect(page.locator("tbody tr").first()).toContainText(name);
});

test("weekly plan: 3 sessions a week booked from the board", async ({ page, request }) => {
  const coordTok = await signIn(page, request, "coordinator");
  const physio = await me(request, await token(request, "physiotherapist2"));
  const deskTok = await token(request, "receptionist");

  // A Monday far ahead, unique to this run (no seeded shifts there)
  const base = new Date(Date.UTC(2060, 0, 5)); // a Monday
  base.setUTCDate(base.getUTCDate() + 7 * (Number(stamp()) % 400));
  const monday = iso(base);
  const days = [monday, addDays(monday, 2), addDays(monday, 4)];
  const division = await divisionWithRoom(request, coordTok, 1);
  const shifts: string[] = [];
  const patientName = `Plano Teste ${stamp()}`;
  const patient = await (await api(request, deskTok, "POST", "/patient/", { name: patientName })).json();

  try {
    for (const date of days) {
      const r = await api(request, coordTok, "POST", "/schedule/shifts/", {
        date, start_hour: 9, end_hour: 11, therapist_id: physio.id, division_id: division,
      });
      expect(r.status(), `shift ${date}: ${await r.text()}`).toBe(201);
      shifts.push((await r.json()).id);
    }

    await boardOn(page, monday);
    await page.getByRole("button", { name: "Plano semanal" }).click();
    const dialog = page.getByRole("dialog");
    await dialog.getByLabel("Utente").fill(patientName);
    await dialog.getByRole("button", { name: new RegExp(patientName) }).click();
    await dialog.getByLabel("Fisioterapeuta").selectOption({ label: `${physio.first_name} ${physio.last_name}` });
    await dialog.getByLabel("Hora").selectOption({ label: "09:00" });
    await dialog.getByRole("button", { name: "3× por semana" }).click();
    await dialog.getByLabel("Sessões").fill("3");
    await dialog.getByRole("button", { name: "Pré-visualizar" }).click();
    await expect(dialog.getByText("Serão marcadas 3 sessões")).toBeVisible();
    await dialog.getByRole("button", { name: "Marcar 3 sessões" }).click();
    await expect(dialog).toHaveCount(0);

    for (const date of days) {
      const found = sessionsOf(await boardRows(request, coordTok, date), patient.id);
      expect(found, `session on ${date}`).toHaveLength(1);
      expect(found[0].therapist).toBe(physio.id);
      expect(found[0].hour).toBe(9);
    }
  } finally {
    for (const date of days)
      for (const s of sessionsOf(await boardRows(request, coordTok, date), patient.id))
        await api(request, coordTok, "DELETE", `/schedule/appointments/${s.id}/`);
    for (const id of shifts) await api(request, coordTok, "DELETE", `/schedule/shifts/${id}/`);
  }
});

test("holiday cover: patients move to the cover with a handover note", async ({ browser, page, request }) => {
  const coordTok = await signIn(page, request, "coordinator");
  const awayTok = await token(request, "physiotherapist");
  const coverTok = await token(request, "physiotherapist2");
  const away = await me(request, awayTok);
  const cover = await me(request, coverTok);
  const deskTok = await token(request, "receptionist");

  // A day within the handover window (3 months), unique per run, late hour
  const date = addDays(iso(new Date()), 82 + (Number(stamp()) % 6));
  const division = await divisionWithRoom(request, coordTok, 2);
  // Leftovers of an interrupted run on that day
  const old = (await (await api(request, coordTok, "GET", `/schedule/absences/?from=${date}&to=${date}`)).json()) as { id: string; therapist_id: number }[];
  for (const a of old) if (a.therapist_id === away.id) await api(request, coordTok, "DELETE", `/schedule/absences/${a.id}/`);
  const leftover = (await (await api(request, coordTok, "GET", `/schedule/board/?date=${date}`)).json()).therapists as (BoardRow & {
    shifts: { id: string; start_hour: number }[];
  })[];
  for (const row of leftover.filter((r) => r.id === away.id || r.id === cover.id)) {
    for (const h of row.hours.filter((x) => x.hour === 19))
      for (const a of h.appointments) await api(request, coordTok, "DELETE", `/schedule/appointments/${a.id}/`);
    for (const sh of row.shifts.filter((x) => x.start_hour === 19)) await api(request, coordTok, "DELETE", `/schedule/shifts/${sh.id}/`);
  }

  const shifts: string[] = [];
  const patientName = `Férias Teste ${stamp()}`;
  const patient = await (await api(request, deskTok, "POST", "/patient/", { name: patientName })).json();
  const handover = `Joelho: progredir carga, gelo no fim (${stamp()})`;

  try {
    for (const who of [away.id, cover.id]) {
      const r = await api(request, coordTok, "POST", "/schedule/shifts/", {
        date, start_hour: 19, end_hour: 20, therapist_id: who, division_id: division,
      });
      expect(r.status(), `shift: ${await r.text()}`).toBe(201);
      shifts.push((await r.json()).id);
    }
    const booked = await api(request, coordTok, "POST", "/schedule/appointments/", {
      date, hour: 19, therapist_id: away.id, patient_id: patient.id,
    });
    expect(booked.status()).toBe(201);

    // The coordinator records the absence and the handover in the app
    await page.goto("/time-board");
    await page.getByRole("button", { name: "Ausências" }).click();
    await page.getByRole("button", { name: "Nova ausência" }).click();
    const dialog = page.getByRole("dialog");
    await dialog.getByLabel("Fisioterapeuta ausente").selectOption({ label: `${away.first_name} ${away.last_name}` });
    await dialog.getByLabel("Fisioterapeuta substituto").selectOption({ label: `${cover.first_name} ${cover.last_name}` });
    for (const label of ["De", "Até"]) {
      await dialog.getByLabel(label, { exact: true }).fill(dmy(date));
      await dialog.getByLabel(label, { exact: true }).press("Tab");
    }
    await dialog.getByLabel("Documento de passagem").fill("Plano geral para a semana.");
    await dialog.getByRole("button", { name: "Pré-visualizar" }).click();
    await expect(dialog.getByText("1 de 1 sessões mudam de fisioterapeuta")).toBeVisible();
    await dialog.getByLabel(patientName, { exact: true }).fill(handover);
    await dialog.getByRole("button", { name: "Confirmar ausência" }).click();
    await expect(page.getByRole("dialog", { name: "Nova ausência" })).toHaveCount(0);

    const moved = sessionsOf(await boardRows(request, coordTok, date), patient.id);
    expect(moved).toHaveLength(1);
    expect(moved[0].therapist).toBe(cover.id);

    // The cover physio finds the note on the patient
    const coverPage = await browser.newPage();
    await signIn(coverPage, request, "physiotherapist2");
    await boardOn(coverPage, date);
    await coverPage.locator("[data-physio-row]", { hasText: `${cover.first_name} ${cover.last_name}` }).locator("button").first().click();
    await coverPage.getByRole("dialog").getByRole("button", { name: new RegExp(patientName) }).click();
    await expect(coverPage.getByText(handover)).toBeVisible();
    await coverPage.close();
  } finally {
    const abs = (await (await api(request, coordTok, "GET", `/schedule/absences/?from=${date}&to=${date}`)).json()) as { id: string; therapist_id: number }[];
    for (const a of abs) if (a.therapist_id === away.id) await api(request, coordTok, "DELETE", `/schedule/absences/${a.id}/`);
    for (const s of sessionsOf(await boardRows(request, coordTok, date), patient.id))
      await api(request, coordTok, "DELETE", `/schedule/appointments/${s.id}/`);
    for (const id of shifts) await api(request, coordTok, "DELETE", `/schedule/shifts/${id}/`);
  }
});

test("inventory: use single items, open and finish containers", async ({ page, request }) => {
  const coordTok = await signIn(page, request, "coordinator");
  const s = stamp();
  const make = async (name: string, tracking: string) => {
    const r = await api(request, coordTok, "POST", "/inventory/items/", {
      name, category: "Teste", unit: tracking === "container" ? "frascos" : "pares", tracking, quantity: 3, min_quantity: 1,
    });
    expect(r.status(), name).toBe(201);
    return (await r.json()).id as string;
  };
  const gloves = `Luvas Teste ${s}`;
  const gel = `Gel Teste ${s}`;
  const ids = [await make(gloves, "unit"), await make(gel, "container")];
  const item = async (id: string) => {
    const list = (await (await api(request, coordTok, "GET", "/inventory/items/?search=Teste")).json()) as { id: string; quantity: number; in_use: number }[];
    return list.find((x) => x.id === id)!;
  };
  const act = async (name: string, action: string) => {
    await page.locator("li", { hasText: name }).getByRole("button", { name: action, exact: true }).click();
    await page.getByRole("dialog").getByRole("button", { name: "Guardar" }).click();
    await expect(page.getByRole("dialog")).toHaveCount(0);
  };

  try {
    await page.goto("/inventory");
    await page.getByLabel("Pesquisar artigos").fill(`Teste ${s}`);
    await expect(page.locator("li", { hasText: gloves })).toBeVisible();

    // Single-use: "Usar" takes it out of stock; there is no "Abrir"
    await expect(page.locator("li", { hasText: gloves }).getByRole("button", { name: "Abrir", exact: true })).toHaveCount(0);
    await act(gloves, "Usar");
    await expect.poll(async () => (await item(ids[0])).quantity).toBe(2);

    // Container: "Abrir" moves one from sealed to in use; "Terminar" ends it
    await expect(page.locator("li", { hasText: gel }).getByRole("button", { name: "Usar", exact: true })).toHaveCount(0);
    await act(gel, "Abrir");
    await expect.poll(async () => item(ids[1])).toMatchObject({ quantity: 2, in_use: 1 });
    await act(gel, "Terminar");
    await expect.poll(async () => item(ids[1])).toMatchObject({ quantity: 2, in_use: 0 });
  } finally {
    for (const id of ids) await api(request, coordTok, "DELETE", `/inventory/items/${id}/`);
  }
});
