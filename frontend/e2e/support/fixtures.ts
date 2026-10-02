import { expect, test as base } from "@playwright/test";
import {
  createPersona as createTestPersona,
  deletePersonas,
  type PersonaOptions,
  type TestPersona,
} from "./personas";

type PersonaFactory = (options: PersonaOptions) => Promise<TestPersona>;

type TestFixtures = {
  createPersona: PersonaFactory;
};

export const test = base.extend<TestFixtures>({
  createPersona: async ({}, provide) => {
    const createdPersonas: TestPersona[] = [];

    await provide(async (options) => {
      const persona = await createTestPersona(options);
      createdPersonas.push(persona);
      return persona;
    });

    await deletePersonas(createdPersonas.map((persona) => persona.userId));
  },
});

export { expect };
