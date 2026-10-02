import type { BrowserContext } from "@playwright/test";
import type { TestPersona } from "./personas";

export async function signInAs(
  context: BrowserContext,
  persona: TestPersona,
) {
  await context.addCookies([
    {
      name: "limen_session",
      value: persona.sessionToken,
      url: "http://localhost:8080",
      sameSite: "Lax",
    },
  ]);
}