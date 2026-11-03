import { test, expect, login, API_URL } from "../fixtures";

async function chooseDayPass(page: import("@playwright/test").Page) {
  await login(page, "member");
  await page.goto("/pricing");
  const card = page
    .getByRole("article")
    .filter({ has: page.getByRole("heading", { name: "Day Pass", exact: true }) });
  await card.getByRole("button", { name: "Choose this pass", exact: true }).click();
  await expect(page.getByRole("heading", { name: "E2E checkout" })).toBeVisible();
  return page.url();
}

test("checkout and duplicate signed webhooks grant exactly one membership", async ({ page }) => {
  const checkoutURL = await chooseDayPass(page);
  await page.getByRole("button", { name: "Pay", exact: true }).click();
  await expect(page).toHaveURL(/\/checkout\?success=true$/);
  await page.goto("/profile");
  await expect(page.getByText("Day Pass", { exact: true })).toBeVisible();
  const before = await page.request.get(`${API_URL}/membership/me/all`);
  expect(before.ok()).toBeTruthy();
  const memberships = await before.json();
  // Replay provider completion through the real signature-verifying handler.
  expect((await page.request.post(`${checkoutURL}/complete`, { maxRedirects: 0 })).status()).toBe(
    303,
  );
  const after = await page.request.get(`${API_URL}/membership/me/all`);
  expect(await after.json()).toEqual(memberships);
  await page.reload();
  await expect(page.getByText("Day Pass", { exact: true })).toHaveCount(1);
  await page.goto("/pricing");
  const card = page
    .getByRole("article")
    .filter({ has: page.getByRole("heading", { name: "Day Pass", exact: true }) });
  await expect(card.getByText("Membership already owned", { exact: true })).toBeVisible();
});

test("cancelled checkout grants no membership", async ({ page }) => {
  await chooseDayPass(page);
  await page.getByRole("button", { name: "Cancel payment", exact: true }).click();
  await expect(page).toHaveURL(/\/checkout\?canceled=true$/);
  await page.goto("/profile");
  await expect(page.getByText("No active memberships.", { exact: true })).toBeVisible();
  await page.reload();
  await expect(page.getByText("No active memberships.", { exact: true })).toBeVisible();
});

test("invalid payment webhook signatures are rejected", async ({ request }) => {
  const response = await request.post(`${API_URL}/webhooks/stripe`, {
    headers: { "Stripe-Signature": "t=1,v1=invalid" },
    data: { type: "checkout.session.completed", data: { object: {} } },
  });
  expect(response.status()).toBe(400);
  expect(await response.json()).toMatchObject({ code: "INVALID_WEBHOOK_SIGNATURE" });
});
