import { test, expect, login, API_URL } from "../fixtures";

test("real OAuth callback creates a persistent member session", async ({ page }) => {
  await login(page, "member");
  await page.goto("/profile");
  const profileEmail = page.getByRole("main").getByText("member@e2e.invalid", { exact: true });
  await expect(profileEmail).toBeVisible();
  await expect(page.getByText("No active memberships.", { exact: true })).toBeVisible();
  await page.reload();
  await expect(profileEmail).toBeVisible();
  const response = await page.request.get(`${API_URL}/profile`);
  expect(response.status()).toBe(200);
  expect(await response.json()).toMatchObject({ user: { email: "member@e2e.invalid" } });
});

test("member cannot access admin UI or API", async ({ page }) => {
  await login(page, "member");
  expect((await page.request.get(`${API_URL}/admin/users`)).status()).toBe(403);
  await page.goto("/admin/users");
  await expect(page).toHaveURL(/\/403$/);
});

test("incomplete user completes validated student onboarding", async ({ page }) => {
  await login(page, "new");
  const response = await page.request.get(`${API_URL}/membership/me/all`);
  expect(response.status()).toBe(403);
  expect(await response.json()).toMatchObject({ code: "ONBOARDING_REQUIRED" });
  await expect(page.getByRole("button", { name: "Continue", exact: true })).toBeDisabled();
  await page.getByLabel("Full name", { exact: true }).fill("E2E Onboarded Student");
  await page.getByRole("button", { name: /Yes/ }).click();
  await page.getByLabel("Student number", { exact: true }).fill("123");
  await expect(page.getByRole("button", { name: "Continue", exact: true })).toBeDisabled();
  await page.getByLabel("Student number", { exact: true }).fill("12345678");
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(page).toHaveURL("http://localhost:13000/");
  await page.goto("/profile");
  await expect(page.getByText("12345678", { exact: true })).toBeVisible();
  await page.reload();
  await expect(page.getByText("12345678", { exact: true })).toBeVisible();
  expect((await page.request.get(`${API_URL}/membership/me/all`)).status()).toBe(200);
});

test("profile distinguishes active memberships from history", async ({ page }) => {
  await login(page, "history");
  await page.goto("/profile");
  await expect(
    page.getByRole("heading", { name: "Active memberships", exact: true }),
  ).toBeVisible();
  await expect(page.getByText("Basic Tier", { exact: true })).toBeVisible();
  await expect(
    page.getByRole("heading", { name: "Membership history", exact: true }),
  ).toBeVisible();
  await expect(page.getByText("Day Pass", { exact: true })).toBeVisible();
  await expect(page.getByText("Expired", { exact: true })).toBeVisible();
});
