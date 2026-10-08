import { signInAs } from "./support/auth";
import { expect, test } from "./support/fixtures";
import { giveActiveMembership } from "./support/memberships";
import { setPersonaStudentStatus } from "./support/personas";

test("student who purchased Basic for $15 sees a $10 Lounge upgrade", async ({
  page,
  context,
  createPersona,
}) => {
  const student = await createPersona({
    isStudent: true,
    groups: ["member"],
  });

  await giveActiveMembership(student, {
    tier: "basic",
    amountPaidCents: 1500,
  });

  await signInAs(context, student);
  await page.goto("/pricing");

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");
  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");
  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(basicCard.getByText("Not currently available")).toBeVisible();
  await expect(basicCard.getByRole("button")).toHaveCount(0);

  await expect(loungeCard).toContainText("Your upgrade price");
  await expect(loungeCard).toContainText("$10");
  await expect(loungeCard.getByRole("button", { name: "Upgrade to Lounge" })).toBeEnabled();

  await expect(dayPassCard.getByText("Not currently available")).toBeVisible();
  await expect(dayPassCard.getByRole("button")).toHaveCount(0);
});

test("student with a $15 Basic pass sees a $17.50 upgrade after becoming non-student", async ({
  page,
  context,
  createPersona,
}) => {
  const student = await createPersona({
    isStudent: true,
    groups: ["member"],
  });

  await giveActiveMembership(student, {
    tier: "basic",
    amountPaidCents: 1500,
  });

  await signInAs(context, student);
  await page.goto("/pricing");

  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");

  await expect(loungeCard).toContainText("Your upgrade price");
  await expect(loungeCard).toContainText("$10");

  await setPersonaStudentStatus(student, false);
  await page.reload();

  await expect(loungeCard).toContainText("Your upgrade price");
  await expect(loungeCard).toContainText("$17.5");
  await expect(loungeCard.getByRole("button", { name: "Upgrade to Lounge" })).toBeEnabled();
});

test("non-student who purchased Basic for $22.50 sees a $10 Lounge upgrade", async ({
  page,
  context,
  createPersona,
}) => {
  const nonStudent = await createPersona({
    isStudent: false,
    groups: ["member"],
  });

  await giveActiveMembership(nonStudent, {
    tier: "basic",
    amountPaidCents: 2250,
  });

  await signInAs(context, nonStudent);
  await page.goto("/pricing");

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");
  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");
  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(basicCard.getByText("Not currently available")).toBeVisible();
  await expect(basicCard.getByRole("button")).toHaveCount(0);

  await expect(loungeCard).toContainText("Your upgrade price");
  await expect(loungeCard).toContainText("$10");
  await expect(loungeCard.getByRole("button", { name: "Upgrade to Lounge" })).toBeEnabled();

  await expect(dayPassCard.getByText("Not currently available")).toBeVisible();
  await expect(dayPassCard.getByRole("button")).toHaveCount(0);
});

test("non-student with a $22.50 Basic pass sees a $2.50 upgrade after becoming a student", async ({
  page,
  context,
  createPersona,
}) => {
  const nonStudent = await createPersona({
    isStudent: false,
    groups: ["member"],
  });

  await giveActiveMembership(nonStudent, {
    tier: "basic",
    amountPaidCents: 2250,
  });

  await signInAs(context, nonStudent);
  await page.goto("/pricing");

  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");

  await expect(loungeCard).toContainText("Your upgrade price");
  await expect(loungeCard).toContainText("$10");

  await setPersonaStudentStatus(nonStudent, true);
  await page.reload();

  await expect(loungeCard).toContainText("Your upgrade price");
  await expect(loungeCard).toContainText(/\$2\.50?/);
  await expect(loungeCard.getByRole("button", { name: "Upgrade to Lounge" })).toBeEnabled();
});

test("student with a $5 Day pass can purchase Basic or Lounge at full student price", async ({
  page,
  context,
  createPersona,
}) => {
  const student = await createPersona({
    isStudent: true,
    groups: ["member"],
  });

  await giveActiveMembership(student, {
    tier: "day",
    amountPaidCents: 500,
  });

  await signInAs(context, student);
  await page.goto("/pricing");

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");
  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");
  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(basicCard).toContainText("Your eligible price");
  await expect(basicCard).toContainText("$15");
  await expect(basicCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  await expect(loungeCard).toContainText("Your eligible price");
  await expect(loungeCard).toContainText("$25");
  await expect(loungeCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  await expect(dayPassCard.getByText("Not currently available")).toBeVisible();
  await expect(dayPassCard.getByRole("button")).toHaveCount(0);
});

test("non-student with a $10 Day pass can purchase Basic or Lounge at full community price", async ({
  page,
  context,
  createPersona,
}) => {
  const nonStudent = await createPersona({
    isStudent: false,
    groups: ["member"],
  });

  await giveActiveMembership(nonStudent, {
    tier: "day",
    amountPaidCents: 1000,
  });

  await signInAs(context, nonStudent);
  await page.goto("/pricing");

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");
  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");
  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(basicCard).toContainText("Your eligible price");
  await expect(basicCard).toContainText("$22.5");
  await expect(basicCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  await expect(loungeCard).toContainText("Your eligible price");
  await expect(loungeCard).toContainText("$32.5");
  await expect(loungeCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  await expect(dayPassCard.getByText("Not currently available")).toBeVisible();
  await expect(dayPassCard.getByRole("button")).toHaveCount(0);
});

test("Lounge holder cannot purchase another pass", async ({ page, context, createPersona }) => {
  const student = await createPersona({
    isStudent: true,
    groups: ["member"],
  });

  await giveActiveMembership(student, {
    tier: "lounge",
    amountPaidCents: 2500,
  });

  await signInAs(context, student);
  await page.goto("/pricing");

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");
  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");
  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  for (const card of [basicCard, loungeCard, dayPassCard]) {
    await expect(card.getByText("Not currently available")).toBeVisible();
    await expect(card.getByRole("button")).toHaveCount(0);
  }
});

test("Competitive holder cannot purchase another pass", async ({
  page,
  context,
  createPersona,
}) => {
  const player = await createPersona({
    isStudent: true,
    groups: ["member", "competitive_team"],
  });

  await giveActiveMembership(player, {
    tier: "competitive_team",
    amountPaidCents: 0,
    groupAtPurchase: "competitive_team",
  });

  await signInAs(context, player);
  await page.goto("/pricing");

  await expect(page.getByRole("heading", { name: "Competitive Player Tier" })).toHaveCount(0);

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");
  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");
  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  for (const card of [basicCard, loungeCard, dayPassCard]) {
    await expect(card.getByText("Not currently available")).toBeVisible();
    await expect(card.getByRole("button")).toHaveCount(0);
  }
});

test("Executive holder cannot purchase another pass", async ({ page, context, createPersona }) => {
  const executive = await createPersona({
    isStudent: true,
    groups: ["member", "executive"],
  });

  await giveActiveMembership(executive, {
    tier: "executive",
    amountPaidCents: 1500,
    groupAtPurchase: "executive",
  });

  await signInAs(context, executive);
  await page.goto("/pricing");

  await expect(page.getByRole("heading", { name: "Executive Tier" })).toHaveCount(0);

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");
  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");
  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  for (const card of [basicCard, loungeCard, dayPassCard]) {
    await expect(card.getByText("Not currently available")).toBeVisible();
    await expect(card.getByRole("button")).toHaveCount(0);
  }
});
