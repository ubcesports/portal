import { execFile } from "node:child_process";
import path from "node:path";
import { promisify } from "node:util";
import { test as base, expect, type Page } from "@playwright/test";

export const API_URL = "http://localhost:18080";
const exec = promisify(execFile);

export const test = base.extend<{ resetDatabase: void }>({
  resetDatabase: [
    async ({}, use) => {
      if (!process.env.E2E_PROJECT) {
        throw new Error("Run make e2e (or make e2e-ui) from the repository root.");
      }
      await exec("bash", [path.resolve(process.cwd(), "../scripts/e2e.sh"), "--reset"]);
      await use();
    },
    { auto: true },
  ],
});

export { expect };

export async function login(page: Page, persona: "member" | "admin" | "new" | "history") {
  await page.goto("/login");
  await expect(page.getByRole("heading", { name: "E2E identity provider" })).toBeVisible();
  await page.getByLabel("Test identity").selectOption(persona);
  await page.getByRole("button", { name: "Sign in", exact: true }).click();
  await expect(page).toHaveURL(
    persona === "new" ? "http://localhost:13000/onboard" : "http://localhost:13000/",
  );
}
