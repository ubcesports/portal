import { signInAs } from "./support/auth";
import { expect, test } from "./support/fixtures";

/*
  Student pricing
*/

test("student sees their price", async ({ page, context, createPersona }) => {
  const student = await createPersona({
    isStudent: true,
    groups: ["member"],
  });

  await signInAs(context, student);
  await page.goto("/pricing");

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");

  await expect(basicCard).toContainText("Your eligible price");
  await expect(basicCard).toContainText("$15");
  await expect(basicCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");

  await expect(loungeCard).toContainText("Your eligible price");
  await expect(loungeCard).toContainText("$25");
  await expect(loungeCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(dayPassCard).toContainText("Your eligible price");
  await expect(dayPassCard).toContainText("$5");
  await expect(dayPassCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();
});

/*
  Non student pricing
*/

test("non student sees their price", async ({ page, context, createPersona }) => {
  const nonstudent = await createPersona({
    isStudent: false,
    groups: ["member"],
  });

  await signInAs(context, nonstudent);
  await page.goto("/pricing");

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");

  await expect(basicCard).toContainText("Your eligible price");
  await expect(basicCard).toContainText("$22.5");
  await expect(basicCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");

  await expect(loungeCard).toContainText("Your eligible price");
  await expect(loungeCard).toContainText("$32.5");
  await expect(loungeCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(dayPassCard).toContainText("Your eligible price");
  await expect(dayPassCard).toContainText("$10");
  await expect(dayPassCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();
});

/*
  Competitive team pricing
*/

test("student competitive player sees their price", async ({ page, context, createPersona }) => {
  const player = await createPersona({
    isStudent: true,
    groups: ["member", "competitive_team"],
  });

  await signInAs(context, player);
  await page.goto("/pricing");

  const compTierCard = page
    .getByRole("heading", { name: "Competitive Player Tier" })
    .locator("xpath=ancestor::article");

  await expect(compTierCard).toContainText("$0");
  await expect(compTierCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");

  await expect(basicCard).toContainText("$15");
  await expect(basicCard).toContainText("$22.5");

  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");

  await expect(loungeCard).toContainText("$25");
  await expect(loungeCard).toContainText("$32.5");

  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(dayPassCard).toContainText("$5");
  await expect(dayPassCard).toContainText("$10");

  for (const card of [basicCard, loungeCard, dayPassCard]) {
    await expect(card.getByText("Not currently available")).toBeVisible();
    await expect(card.getByRole("button")).toHaveCount(0);
  }
});

test("non student competitive player sees their price", async ({
  page,
  context,
  createPersona,
}) => {
  const player = await createPersona({
    isStudent: false,
    groups: ["member", "competitive_team"],
  });

  await signInAs(context, player);
  await page.goto("/pricing");

  const compTierCard = page
    .getByRole("heading", { name: "Competitive Player Tier" })
    .locator("xpath=ancestor::article");

  await expect(compTierCard).toContainText("$0");
  await expect(compTierCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");

  await expect(basicCard).toContainText("$15");
  await expect(basicCard).toContainText("$22.5");

  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");

  await expect(loungeCard).toContainText("$25");
  await expect(loungeCard).toContainText("$32.5");

  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(dayPassCard).toContainText("$5");
  await expect(dayPassCard).toContainText("$10");

  for (const card of [basicCard, loungeCard, dayPassCard]) {
    await expect(card.getByText("Not currently available")).toBeVisible();
    await expect(card.getByRole("button")).toHaveCount(0);
  }
});

/*
  Exec pricing
*/

test("student exec sees their price", async ({ page, context, createPersona }) => {
  const exec = await createPersona({
    isStudent: true,
    groups: ["member", "executive"],
  });

  await signInAs(context, exec);
  await page.goto("/pricing");

  const compTierCard = page
    .getByRole("heading", { name: "Executive Tier" })
    .locator("xpath=ancestor::article");

  await expect(compTierCard).toContainText("$15");
  await expect(compTierCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");

  await expect(basicCard).toContainText("$15");
  await expect(basicCard).toContainText("$22.5");

  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");

  await expect(loungeCard).toContainText("$25");
  await expect(loungeCard).toContainText("$32.5");

  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(dayPassCard).toContainText("$5");
  await expect(dayPassCard).toContainText("$10");

  for (const card of [basicCard, loungeCard, dayPassCard]) {
    await expect(card.getByText("Not currently available")).toBeVisible();
    await expect(card.getByRole("button")).toHaveCount(0);
  }
});

test("non student exec sees their price", async ({ page, context, createPersona }) => {
  const exec = await createPersona({
    isStudent: false,
    groups: ["member", "executive"],
  });

  await signInAs(context, exec);
  await page.goto("/pricing");

  const compTierCard = page
    .getByRole("heading", { name: "Executive Tier" })
    .locator("xpath=ancestor::article");

  await expect(compTierCard).toContainText("$22.5");
  await expect(compTierCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");

  await expect(basicCard).toContainText("$15");
  await expect(basicCard).toContainText("$22.5");

  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");

  await expect(loungeCard).toContainText("$25");
  await expect(loungeCard).toContainText("$32.5");

  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(dayPassCard).toContainText("$5");
  await expect(dayPassCard).toContainText("$10");

  for (const card of [basicCard, loungeCard, dayPassCard]) {
    await expect(card.getByText("Not currently available")).toBeVisible();
    await expect(card.getByRole("button")).toHaveCount(0);
  }
});

/*
  Director pricing
*/

test("student director sees their price", async ({ page, context, createPersona }) => {
  const exec = await createPersona({
    isStudent: true,
    groups: ["member", "executive", "director"],
  });

  await signInAs(context, exec);
  await page.goto("/pricing");

  const compTierCard = page
    .getByRole("heading", { name: "Executive Tier" })
    .locator("xpath=ancestor::article");

  await expect(compTierCard).toContainText("$15");
  await expect(compTierCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");

  await expect(basicCard).toContainText("$15");
  await expect(basicCard).toContainText("$22.5");

  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");

  await expect(loungeCard).toContainText("$25");
  await expect(loungeCard).toContainText("$32.5");

  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(dayPassCard).toContainText("$5");
  await expect(dayPassCard).toContainText("$10");

  for (const card of [basicCard, loungeCard, dayPassCard]) {
    await expect(card.getByText("Not currently available")).toBeVisible();
    await expect(card.getByRole("button")).toHaveCount(0);
  }
});

test("non student director sees their price", async ({ page, context, createPersona }) => {
  const exec = await createPersona({
    isStudent: false,
    groups: ["member", "executive", "director"],
  });

  await signInAs(context, exec);
  await page.goto("/pricing");

  const compTierCard = page
    .getByRole("heading", { name: "Executive Tier" })
    .locator("xpath=ancestor::article");

  await expect(compTierCard).toContainText("$22.5");
  await expect(compTierCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");

  await expect(basicCard).toContainText("$15");
  await expect(basicCard).toContainText("$22.5");

  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");

  await expect(loungeCard).toContainText("$25");
  await expect(loungeCard).toContainText("$32.5");

  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(dayPassCard).toContainText("$5");
  await expect(dayPassCard).toContainText("$10");

  for (const card of [basicCard, loungeCard, dayPassCard]) {
    await expect(card.getByText("Not currently available")).toBeVisible();
    await expect(card.getByRole("button")).toHaveCount(0);
  }
});

/*
  Board pricing
*/

test("student board member sees their price", async ({ page, context, createPersona }) => {
  const exec = await createPersona({
    isStudent: true,
    groups: ["member", "executive", "board"],
  });

  await signInAs(context, exec);
  await page.goto("/pricing");

  const compTierCard = page
    .getByRole("heading", { name: "Executive Tier" })
    .locator("xpath=ancestor::article");

  await expect(compTierCard).toContainText("$15");
  await expect(compTierCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");

  await expect(basicCard).toContainText("$15");
  await expect(basicCard).toContainText("$22.5");

  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");

  await expect(loungeCard).toContainText("$25");
  await expect(loungeCard).toContainText("$32.5");

  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(dayPassCard).toContainText("$5");
  await expect(dayPassCard).toContainText("$10");

  for (const card of [basicCard, loungeCard, dayPassCard]) {
    await expect(card.getByText("Not currently available")).toBeVisible();
    await expect(card.getByRole("button")).toHaveCount(0);
  }
});

test("non student board member sees their price", async ({ page, context, createPersona }) => {
  const exec = await createPersona({
    isStudent: false,
    groups: ["member", "executive", "board"],
  });

  await signInAs(context, exec);
  await page.goto("/pricing");

  const compTierCard = page
    .getByRole("heading", { name: "Executive Tier" })
    .locator("xpath=ancestor::article");

  await expect(compTierCard).toContainText("$22.5");
  await expect(compTierCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");

  await expect(basicCard).toContainText("$15");
  await expect(basicCard).toContainText("$22.5");

  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");

  await expect(loungeCard).toContainText("$25");
  await expect(loungeCard).toContainText("$32.5");

  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(dayPassCard).toContainText("$5");
  await expect(dayPassCard).toContainText("$10");

  for (const card of [basicCard, loungeCard, dayPassCard]) {
    await expect(card.getByText("Not currently available")).toBeVisible();
    await expect(card.getByRole("button")).toHaveCount(0);
  }
});

/*
  Some unique group combinations
*/

test("student comp player + exec sees their price", async ({ page, context, createPersona }) => {
  const exec = await createPersona({
    isStudent: true,
    groups: ["member", "competitive_team", "executive"],
  });

  await signInAs(context, exec);
  await page.goto("/pricing");

  const compTierCard = page
    .getByRole("heading", { name: "Executive Tier" })
    .locator("xpath=ancestor::article");

  await expect(compTierCard).toContainText("$15");
  await expect(compTierCard.getByRole("button", { name: "Choose this pass" })).toBeEnabled();

  const basicCard = page
    .getByRole("heading", { name: "Basic Tier" })
    .locator("xpath=ancestor::article");

  await expect(basicCard).toContainText("$15");
  await expect(basicCard).toContainText("$22.5");

  const loungeCard = page
    .getByRole("heading", { name: "Lounge Tier" })
    .locator("xpath=ancestor::article");

  await expect(loungeCard).toContainText("$25");
  await expect(loungeCard).toContainText("$32.5");

  const dayPassCard = page
    .getByRole("heading", { name: "Day Pass" })
    .locator("xpath=ancestor::article");

  await expect(dayPassCard).toContainText("$5");
  await expect(dayPassCard).toContainText("$10");

  for (const card of [basicCard, loungeCard, dayPassCard]) {
    await expect(card.getByText("Not currently available")).toBeVisible();
    await expect(card.getByRole("button")).toHaveCount(0);
  }
});
