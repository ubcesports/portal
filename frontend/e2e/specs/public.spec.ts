import { test, expect, API_URL } from "../fixtures";

test("public landing navigates to a database-backed catalog", async ({ page, request }) => {
  const response = await request.get(`${API_URL}/membership/tiers`);
  expect(response.ok()).toBeTruthy();
  expect(await response.json()).toEqual(
    expect.arrayContaining([expect.objectContaining({ title: "Basic Tier" })]),
  );
  await page.goto("/");
  await expect(page.getByRole("heading", { level: 1 })).toContainText("gaming community");
  await page.getByRole("link", { name: "Become a member", exact: true }).click();
  await expect(page).toHaveURL(/\/pricing$/);
  await expect(page.getByRole("heading", { name: "Choose your UBCEA pass" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Basic Tier", exact: true })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Lounge Tier", exact: true })).toBeVisible();
  await page.reload();
  await expect(page.getByRole("heading", { name: "Basic Tier", exact: true })).toBeVisible();
});

test("anonymous protected API requests are rejected and profile starts login", async ({
  page,
  request,
}) => {
  expect((await request.get(`${API_URL}/profile`)).status()).toBe(401);
  expect((await request.get(`${API_URL}/admin/users`)).status()).toBe(401);
  await page.goto("/profile");
  // /login automatically starts OAuth, so assert the destination rather than
  // racing the transient /login location.
  await expect(page.getByRole("heading", { name: "E2E identity provider" })).toBeVisible();
});
