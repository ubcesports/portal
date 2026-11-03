import { test, expect, login, API_URL } from "../fixtures";

test("admin searches for a member and persists a profile update", async ({ page }) => {
  await login(page, "admin");
  await page.goto("/admin/users");
  await page.getByPlaceholder("Search by full name").fill("E2E Member");
  await page.getByRole("link", { name: "E2E Member", exact: true }).click();
  await expect(page).toHaveURL(/\/admin\/users\/30000000-0000-0000-0000-000000000001$/);
  await page.getByRole("button", { name: "Edit", exact: true }).click();
  await page.getByRole("textbox", { name: "Full name", exact: true }).fill("E2E Updated Member");
  const saved = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith("/admin/users/30000000-0000-0000-0000-000000000001"),
  );
  await page.getByRole("button", { name: "Save", exact: true }).click();
  expect((await saved).ok()).toBeTruthy();
  await page.reload();
  await expect(
    page.getByRole("heading", { name: "E2E Updated Member", exact: true }),
  ).toBeVisible();
  const response = await page.request.get(
    `${API_URL}/admin/users/30000000-0000-0000-0000-000000000001`,
  );
  expect(response.ok()).toBeTruthy();
  expect(await response.json()).toMatchObject({ user: { full_name: "E2E Updated Member" } });
});

test("admin grants an offline membership that is visible to the member", async ({
  page,
  browser,
}) => {
  await login(page, "admin");
  await page.goto("/admin/users/30000000-0000-0000-0000-000000000001");
  await page.getByLabel("Eligible membership").selectOption("20000000-0000-0000-0000-000000000003");
  await page.getByLabel("Amount paid").fill("10.00");
  await page.getByRole("combobox", { name: "Payment method", exact: true }).selectOption("cash");
  await page.getByRole("button", { name: "Add membership", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Tier: Day Pass", exact: true })).toBeVisible();
  await page.reload();
  await expect(page.getByRole("heading", { name: "Tier: Day Pass", exact: true })).toBeVisible();
  const memberContext = await browser.newContext({ baseURL: "http://localhost:13000" });
  try {
    const memberPage = await memberContext.newPage();
    await login(memberPage, "member");
    await memberPage.goto("/profile");
    await expect(memberPage.getByText("Day Pass", { exact: true })).toBeVisible();
  } finally {
    await memberContext.close();
  }
});

test("admin exports fixture users as CSV", async ({ page }) => {
  await login(page, "admin");
  await page.goto("/admin/users");
  const downloading = page.waitForEvent("download");
  await page.getByRole("button", { name: "Export CSV", exact: true }).click();
  const download = await downloading;
  expect(await download.failure()).toBeNull();
  const stream = await download.createReadStream();
  const chunks: Buffer[] = [];
  for await (const chunk of stream) chunks.push(Buffer.from(chunk));
  const csv = Buffer.concat(chunks).toString("utf8");
  expect(csv).toContain("member@e2e.invalid");
  expect(csv).toContain("admin@e2e.invalid");
});
