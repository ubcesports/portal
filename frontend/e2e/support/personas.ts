import { randomUUID } from "node:crypto";
import { Client } from "pg";

export type Group = "member" | "competitive_team" | "executive" | "director" | "board";

export type PersonaOptions = {
  isStudent: boolean;
  groups?: Group[];
};

export type TestPersona = {
  userId: string;
  email: string;
  sessionToken: string;
  isStudent: boolean;
  groups: Group[];
};

export async function createPersona({ isStudent, groups }: PersonaOptions): Promise<TestPersona> {
  const connectionString = process.env.E2E_DATABASE_URL;

  if (!connectionString) {
    throw new Error("E2E_DATABASE_URL is required");
  }

  const client = new Client({ connectionString });
  await client.connect();

  const unique = randomUUID();
  const email = `e2e-${unique}@example.test`;
  const studentId = isStudent ? `TEST-${unique.slice(0, 8)}` : null;
  const sessionToken = randomUUID();
  const personaGroups = [...new Set<Group>(groups ?? ["member"])];

  try {
    await client.query("BEGIN");

    const userResult = await client.query<{ id: string }>(
      `
        INSERT INTO users (
          email,
          student_id,
          role,
          full_name,
          email_verified_at,
          is_student,
          onboarding_completed_at,
          created_at,
          updated_at
        )
        VALUES (
          $1,
          $2,
          'member',
          $3,
          NOW(),
          $4,
          NOW(),
          NOW(),
          NOW()
        )
        RETURNING id
      `,
      [email, studentId, `E2E ${isStudent ? "Student" : "Community"} Member`, isStudent],
    );

    const userId = userResult.rows[0].id;

    if (personaGroups.length > 0) {
      await client.query(
        `
          INSERT INTO user_groups (user_id, "group")
          SELECT $1, unnest($2::group_type[])
        `,
        [userId, personaGroups],
      );
    }

    await client.query(
      `
        INSERT INTO sessions (
          token,
          user_id,
          created_at,
          expires_at,
          last_access,
          metadata
        )
        VALUES (
          $1,
          $2,
          NOW(),
          NOW() + INTERVAL '1 day',
          NOW(),
          $3
        )
      `,
      [
        sessionToken,
        userId,
        JSON.stringify({
          ip_address: "127.0.0.1",
          user_agent: "Playwright",
        }),
      ],
    );

    await client.query("COMMIT");

    return { userId, email, sessionToken, isStudent, groups: personaGroups };
  } catch (error) {
    await client.query("ROLLBACK");
    throw error;
  } finally {
    await client.end();
  }
}

export async function setPersonaGroups(persona: TestPersona, groups: Group[]) {
  const connectionString = process.env.E2E_DATABASE_URL;

  if (!connectionString) {
    throw new Error("E2E_DATABASE_URL is required");
  }

  const client = new Client({ connectionString });
  await client.connect();

  const personaGroups = [...new Set<Group>(groups)];

  try {
    await client.query("BEGIN");
    await client.query(`DELETE FROM user_groups WHERE user_id = $1`, [persona.userId]);

    if (personaGroups.length > 0) {
      await client.query(
        `
          INSERT INTO user_groups (user_id, "group")
          SELECT $1, unnest($2::group_type[])
        `,
        [persona.userId, personaGroups],
      );
    }

    await client.query("COMMIT");
    persona.groups = personaGroups;
  } catch (error) {
    await client.query("ROLLBACK");
    throw error;
  } finally {
    await client.end();
  }
}

export async function setPersonaStudentStatus(persona: TestPersona, isStudent: boolean) {
  const connectionString = process.env.E2E_DATABASE_URL;

  if (!connectionString) {
    throw new Error("E2E_DATABASE_URL is required");
  }

  const client = new Client({ connectionString });
  await client.connect();

  const studentId = isStudent ? `TEST-${randomUUID().slice(0, 8)}` : null;

  try {
    const result = await client.query(
      `
        UPDATE users
        SET
          is_student = $2,
          student_id = $3,
          updated_at = NOW()
        WHERE id = $1
      `,
      [persona.userId, isStudent, studentId],
    );

    if (result.rowCount !== 1) {
      throw new Error(`Persona user "${persona.userId}" was not found`);
    }

    persona.isStudent = isStudent;
  } finally {
    await client.end();
  }
}

export async function deletePersonas(userIds: string[]) {
  if (userIds.length === 0) {
    return;
  }

  const connectionString = process.env.E2E_DATABASE_URL;

  if (!connectionString) {
    throw new Error("E2E_DATABASE_URL is required");
  }

  const client = new Client({ connectionString });
  await client.connect();

  try {
    await client.query("BEGIN");

    await client.query(
      `
        DELETE FROM admin_audit_logs
        WHERE actor_user_id = ANY($1::uuid[])
           OR target_user_id = ANY($1::uuid[])
      `,
      [userIds],
    );

    await client.query(`DELETE FROM transactions WHERE user_id = ANY($1::uuid[])`, [userIds]);
    await client.query(`DELETE FROM memberships WHERE user_id = ANY($1::uuid[])`, [userIds]);
    await client.query(`DELETE FROM sessions WHERE user_id = ANY($1::uuid[])`, [userIds]);
    await client.query(`DELETE FROM accounts WHERE user_id = ANY($1::uuid[])`, [userIds]);
    await client.query(`DELETE FROM user_groups WHERE user_id = ANY($1::uuid[])`, [userIds]);
    await client.query(`DELETE FROM users WHERE id = ANY($1::uuid[])`, [userIds]);

    await client.query("COMMIT");
  } catch (error) {
    await client.query("ROLLBACK");
    throw error;
  } finally {
    await client.end();
  }
}
