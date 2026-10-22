import { APIRequestContext, expect, test } from "@playwright/test";

import { PASSWORD, api, signIn } from "./helpers";

// The changes of 0.6: version, stock permissions by floor/category,
// justified misses, clinic people link, clinics page, "today" on the board,
// and per-patient moves for a holiday. Climutua demo data.

async function token(request: APIRequestContext, username: string) {
  const res = await request.post("/api/v1/auth/login/", { data: { username, password: PASSWORD } });
  return (await res.json()).access as string;
}
const stamp = () => Date.now().toString().slice(-7);

test("the version lives on the profile and the About page, not in the menu", async ({ page, request }) => {
  await signIn(page, request, "cm.fisio3");
  await page.goto("/");
  await expect(page.getByText(/Klyniva v0\.\d+\.\d+/)).toHaveCount(0);
  await page.goto("/user/profile");
  await expect(page.getByText("Sobre o Klyniva").first()).toBeVisible();
  await page.getByRole("link", { name: "Abrir a página Sobre" }).click();
  await expect(page).toHaveURL(/\/about$/);
  const versions = await page.locator("dl dd").allInnerTexts();
  expect(versions[0]).toMatch(/^0\.\d+\.\d+$/);
  expect(versions[1]).toBe(versions[0]); // app and server agree
  await expect(page.getByText(/Não altera o servidor, os dados nem outros computadores/)).toBeVisible();
  await page.getByRole("button", { name: "Procurar atualização" }).click();
  await expect(page.getByText(/Já tem a versão mais recente|Há uma atualização/).first()).toBeVisible();
});

test("stock permissions: the coordinator gives Pedro floor 2, Pedro can then restock only there", async ({ browser, page, request }) => {
  const coordTok = await signIn(page, request, "cm.coordenacao");
  const pedroTok = await token(request, "cm.fisio2");
  const pedro = await (await api(request, pedroTok, "GET", "/users/getcurrentuser/")).json();
  await api(request, coordTok, "PUT", `/inventory/grants/${pedro.id}/`, { rules: [] });

  try {
    // Without permission Pedro can only look
    const pedroPage = await browser.newPage();
    await signIn(pedroPage, request, "cm.fisio2");
    await pedroPage.goto("/inventory");
    await expect(pedroPage.getByText(/Pode consultar o stock mas não alterá-lo/)).toBeVisible();
    await expect(pedroPage.getByRole("button", { name: "Repor", exact: true })).toHaveCount(0);

    // The coordinator adds "floor 2, every category" from the people list
    await page.goto("/users");
    await page.getByRole("button", { name: "Permissões de inventário · Pedro Almeida" }).click();
    const dialog = page.getByRole("dialog");
    await dialog.getByRole("button", { name: "Adicionar regra" }).click();
    await dialog.getByRole("group", { name: "Piso" }).getByRole("button", { name: "Piso 2", exact: true }).click();
    await dialog.getByRole("button", { name: "Guardar" }).click();
    await expect(dialog).toHaveCount(0);
    await expect(page.getByRole("row", { name: /Pedro Almeida/ })).toContainText("Piso 2 · todas as categorias");

    // Pedro: TENS electrodes (floor 2) can be restocked, swim caps (floor 1) not
    await pedroPage.reload();
    await pedroPage.getByLabel("Pesquisar artigos").fill("Elétrodos");
    await expect(pedroPage.locator("li", { hasText: "Elétrodos para TENS" }).getByRole("button", { name: "Repor", exact: true })).toBeVisible();
    await pedroPage.getByLabel("Pesquisar artigos").fill("Toucas");
    await expect(pedroPage.locator("li", { hasText: "Toucas de natação" })).toBeVisible();
    await expect(pedroPage.locator("li", { hasText: "Toucas de natação" }).getByRole("button", { name: "Usar", exact: true })).toHaveCount(0);
    await pedroPage.close();
  } finally {
    await api(request, coordTok, "PUT", `/inventory/grants/${pedro.id}/`, { rules: [] });
  }
});

test("inventory filters: single-use vs containers, and by floor", async ({ page, request }) => {
  await signIn(page, request, "cm.coordenacao");
  await page.goto("/inventory");
  await page.getByRole("button", { name: "Embalagens (abrir e terminar)" }).click();
  await expect(page.locator("li", { hasText: "Gel de ultrassons" })).toBeVisible();
  await expect(page.locator("li", { hasText: "Luvas de nitrilo (M)" })).toHaveCount(0);
  await page.getByRole("button", { name: "Descartáveis (usar)" }).click();
  await expect(page.locator("li", { hasText: "Luvas de nitrilo (M)" })).toBeVisible();
  await expect(page.locator("li", { hasText: "Gel de ultrassons" })).toHaveCount(0);
  await page.getByRole("group", { name: "Tipo" }).getByRole("button", { name: "Todos", exact: true }).click().catch(() => undefined);
  await page.getByRole("group", { name: "Piso" }).getByRole("button", { name: "Piso 1", exact: true }).click();
  await expect(page.locator("li", { hasText: "Toucas de natação" })).toBeVisible();
  await expect(page.locator("li", { hasText: "Elétrodos para TENS" })).toHaveCount(0);
  // The explanation of per-session use opens
  await page.getByText("Como funciona o uso por sessão?").first().click();
  await expect(page.getByText(/retira essa quantidade do stock sozinho/).first()).toBeVisible();
});

test("a missed session is justified or not, and the label says which", async ({ page, request }) => {
  const physioTok = await signIn(page, request, "cm.fisio3");
  const reason = `Doente ${stamp()}`;
  await page.goto("/");
  const missButtons = page.getByRole("button", { name: "Marcar falta" });
  test.skip((await missButtons.count()) === 0, "no session today for this physio");
  await missButtons.first().click();
  const dialog = page.getByRole("dialog");
  await dialog.getByRole("radio", { name: "Justificada" }).click();
  await dialog.getByLabel("Motivo").fill(reason);
  await dialog.getByRole("button", { name: "Marcar falta" }).click();
  await expect(dialog).toHaveCount(0);
  await expect(page.getByText("Faltou (justificada)").first()).toBeVisible();

  // Back to booked so the demo day stays as it was
  const today = new Date().toISOString().slice(0, 10);
  const board = await (await api(request, physioTok, "GET", `/schedule/board/?date=${today}`)).json();
  for (const row of board.therapists)
    for (const h of row.hours)
      for (const a of h.appointments)
        if (a.miss_reason === reason) await api(request, physioTok, "PATCH", `/schedule/appointments/${a.id}/`, { status: "booked" });
});

test("a clinic's people open as the people list filtered by clinic", async ({ page, request }) => {
  await signIn(page, request, "stateadmin");
  await page.goto("/facility");
  await page.locator("a[href=\"/users?clinic=6b1f3c1e-5a8e-4c1d-9a0e-000000000002\"]").click();
  await expect(page).toHaveURL(/\/users\?clinic=6b1f3c1e-5a8e-4c1d-9a0e-000000000002$/);
  await expect(page.getByRole("heading", { name: "Pessoas" })).toBeVisible();
  await expect(page.getByLabel("Clínica")).toHaveValue("6b1f3c1e-5a8e-4c1d-9a0e-000000000002");
  await expect(page.getByText("Ana Ferreira")).toBeVisible();
  // The old address still works
  await page.goto("/facility/6b1f3c1e-5a8e-4c1d-9a0e-000000000002/users");
  await expect(page).toHaveURL(/\/users\?clinic=/);
});

test("clinics page: search and chips instead of advanced filters", async ({ page, request }) => {
  await signIn(page, request, "stateadmin");
  await page.goto("/facility");
  await expect(page.getByRole("button", { name: /Filtros avançados|Advanced filters/ })).toHaveCount(0);
  await page.getByRole("searchbox", { name: "Pesquisar clínicas" }).fill("cedofeita");
  await expect(page.getByText("1 clínica")).toBeVisible();
  await expect(page.getByText(/Climutua/).first()).toBeVisible();
});

test("time board: today is marked, and the Today button pulses it in the week view", async ({ page, request }) => {
  await signIn(page, request, "cm.coordenacao");
  await page.goto("/time-board");
  // Day view on today: bold label with a "Hoje" badge; gone on another day
  await expect(page.getByRole("button", { name: "Hoje" })).toHaveAttribute("aria-pressed", "true");
  await page.getByRole("button", { name: "Seguinte" }).click();
  await expect(page.getByRole("button", { name: "Hoje" })).toHaveAttribute("aria-pressed", "false");

  await page.getByRole("tab", { name: "Semana" }).click();
  await page.getByRole("button", { name: "Hoje" }).click();
  const todayCell = page.locator('[aria-current="date"]').first();
  await expect(todayCell).toBeVisible();
  await expect(todayCell).toHaveClass(/today-flash/);
  await expect(todayCell).not.toHaveClass(/today-flash/, { timeout: 4000 });
});

test("holiday: each patient can go to a different physio", async ({ page, request }) => {
  const coordTok = await signIn(page, request, "cm.coordenacao");
  const ids = {} as Record<string, { id: number; first_name: string; last_name: string }>;
  for (const u of ["cm.fisio1", "cm.fisio2", "cm.fisio3"])
    ids[u] = await (await api(request, await token(request, u), "GET", "/users/getcurrentuser/")).json();
  const [away, y, z] = [ids["cm.fisio1"], ids["cm.fisio2"], ids["cm.fisio3"]];
  const name = (p: { first_name: string; last_name: string }) => `${p.first_name} ${p.last_name}`;
  const d = new Date();
  d.setUTCDate(d.getUTCDate() + 84 + (Number(stamp()) % 5));
  const date = d.toISOString().slice(0, 10);
  const dmy = date.split("-").reverse().join("/");
  const divs = (await (await api(request, coordTok, "GET", "/schedule/divisions/")).json()) as { id: string; capacity: number; name: string }[];
  const gym = divs.find((x) => x.name === "Ginásio")!.id;
  const deskTok = await token(request, "cm.rececao1");
  const shifts: string[] = [];
  const patients: { id: string; name: string }[] = [];

  try {
    for (const who of [away, y, z]) {
      const r = await api(request, coordTok, "POST", "/schedule/shifts/", { date, start_hour: 19, end_hour: 20, therapist_id: who.id, division_id: gym });
      expect(r.status(), await r.text()).toBe(201);
      shifts.push((await r.json()).id);
    }
    for (const n of ["A", "B"]) {
      const p = await (await api(request, deskTok, "POST", "/patient/", { name: `Divisão ${n} ${stamp()}` })).json();
      patients.push(p);
      expect((await api(request, coordTok, "POST", "/schedule/appointments/", { date, hour: 19, therapist_id: away.id, patient_id: p.id })).status()).toBe(201);
    }

    await page.goto("/time-board");
    await page.getByRole("button", { name: "Ausências" }).click();
    await page.getByRole("button", { name: "Nova ausência" }).click();
    const dialog = page.getByRole("dialog");
    await dialog.getByLabel("Fisioterapeuta ausente").selectOption({ label: name(away) });
    await dialog.getByLabel("Fisioterapeuta substituto").selectOption({ label: name(y) });
    for (const label of ["De", "Até"]) {
      await dialog.getByLabel(label, { exact: true }).fill(dmy);
      await dialog.getByLabel(label, { exact: true }).press("Tab");
    }
    await dialog.getByRole("button", { name: "Pré-visualizar" }).click();
    await expect(dialog.getByText("2 de 2 sessões mudam de fisioterapeuta")).toBeVisible();

    // Patient B goes to z instead of the cover; the summary follows
    const pickB = dialog.getByLabel(`Quem fica com ${patients[1].name} às 19:00`);
    const zOption = (await pickB.locator("option").allInnerTexts()).find((t) => t.startsWith(name(z)))!;
    await pickB.selectOption({ label: zOption });
    const summary = dialog.getByRole("complementary", { name: "Quem recebe quem" });
    await expect(summary.getByRole("listitem").filter({ hasText: name(z) }).first()).toContainText(patients[1].name);
    await expect(summary.getByRole("listitem").filter({ hasText: name(y) }).first()).toContainText(patients[0].name);
    await dialog.getByRole("button", { name: "Confirmar ausência" }).click();
    await expect(page.getByRole("dialog", { name: "Nova ausência" })).toHaveCount(0);

    const board = await (await api(request, coordTok, "GET", `/schedule/board/?date=${date}`)).json();
    const owner = (pid: string) =>
      board.therapists.find((r: { hours: { appointments: { patient_id: string }[] }[] }) =>
        r.hours.some((h) => h.appointments.some((a) => a.patient_id === pid)),
      )?.id;
    expect(owner(patients[0].id)).toBe(y.id);
    expect(owner(patients[1].id)).toBe(z.id);
  } finally {
    const abs = (await (await api(request, coordTok, "GET", `/schedule/absences/?from=${date}&to=${date}`)).json()) as { id: string; therapist_id: number }[];
    for (const a of abs) if (a.therapist_id === away.id) await api(request, coordTok, "DELETE", `/schedule/absences/${a.id}/`);
    const board = await (await api(request, coordTok, "GET", `/schedule/board/?date=${date}`)).json();
    for (const r of board.therapists)
      for (const h of r.hours)
        for (const a of h.appointments)
          if (patients.some((p) => p.id === a.patient_id)) await api(request, coordTok, "DELETE", `/schedule/appointments/${a.id}/`);
    for (const id of shifts) await api(request, coordTok, "DELETE", `/schedule/shifts/${id}/`);
  }
});
