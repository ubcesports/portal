import { randomUUID } from "node:crypto";
import { Client } from "pg";
import type { Group, TestPersona } from "./personas";

export type TierSlug = "day" | "basic" | "lounge" | "competitive_team" | "executive";

type ActiveMembershipOptions = {
  tier: TierSlug;
  amountPaidCents: number;
  groupAtPurchase?: Group;
};

export async function giveActiveMembership(
  persona: TestPersona,
  { tier, amountPaidCents, groupAtPurchase = "member" }: ActiveMembershipOptions,
) {
  const connectionString = process.env.E2E_DATABASE_URL;

  if (!connectionString) {
    throw new Error("E2E_DATABASE_URL is required");
  }

  const client = new Client({ connectionString });
  await client.connect();

  try {
    await client.query("BEGIN");

    const tierResult = await client.query<{ id: string }>(
      `
        SELECT id
        FROM membership_tiers
        WHERE slug = $1
          AND is_active = TRUE
      `,
      [tier],
    );

    if (tierResult.rowCount !== 1) {
      throw new Error(`Active membership tier "${tier}" was not found`);
    }

    const tierId = tierResult.rows[0].id;
    const membershipResult = await client.query<{ id: string }>(
      `
        INSERT INTO memberships (
          user_id,
          tier_id,
          started_at,
          expires_at,
          cancelled_at
        )
        VALUES (
          $1,
          $2,
          NOW() - INTERVAL '1 day',
          NOW() + INTERVAL '90 days',
          NULL
        )
        RETURNING id
      `,
      [persona.userId, tierId],
    );

    const membershipId = membershipResult.rows[0].id;

    await client.query(
      `
        INSERT INTO transactions (
          user_id,
          membership_id,
          tier_id,
          stripe_payment_intent_id,
          status,
          group_at_purchase,
          student_at_purchase,
          amount_paid_cents,
          purchase_type
        )
        VALUES (
          $1,
          $2,
          $3,
          $4,
          'completed',
          $5::group_type,
          $6,
          $7,
          'new'
        )
      `,
      [
        persona.userId,
        membershipId,
        tierId,
        `e2e_pi_${randomUUID()}`,
        groupAtPurchase,
        persona.isStudent,
        amountPaidCents,
      ],
    );

    await client.query("COMMIT");

    return { membershipId, tierId };
  } catch (error) {
    await client.query("ROLLBACK");
    throw error;
  } finally {
    await client.end();
  }
}
