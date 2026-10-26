import type { GroupType } from "./user.types";
import type { RedirectUrlResponse } from "./api.types";

export type TransactionStatusType = "pending" | "completed" | "failed" | "refunded" | "expired";

export type MembershipStatus = "active" | "expired" | "cancelled";

export type PurchaseType = "new" | "upgrade";

export type PaymentMethod = "stripe" | "cash" | "etransfer";

export type MembershipExpirationType = "day" | "semester" | "year";

export type MembershipTierPrice = {
  price: number;
  price_id: string;
  is_student_required: boolean | null;
};

export type MembershipTier = {
  id: string;
  title: string;
  description: string;
  benefits: string[];
  limitations: string[];
  slug: string;
  product_id: string;
  prices: MembershipTierPrice[];
  program_id: string;
  program_name: string;
  expiration_type: MembershipExpirationType;
};

export type TierUnavailableReason =
  | "already_owned"
  | "not_eligible_current_membership"
  | "executive_restricted"
  | "competitive_restricted"
  | "purchase_closed"
  | "unavailable";

export type EligibleMembershipTier = Omit<MembershipTier, "prices"> &
  (
    | {
        eligible: true;
        purchase_type: PurchaseType;
        prices: MembershipTierPrice;
      }
    | {
        eligible: false;
        unavailable_reason: TierUnavailableReason;
        purchase_opens_at?: string;
      }
  );

export type Transaction = {
  id: string;
  amount_paid: string;
  status: TransactionStatusType;
  group_at_purchase: GroupType;
  student_at_purchase: boolean;
  stripe_payment_intent_id: string;
  purchase_type: PurchaseType;
  payment_method: PaymentMethod;
};

export type Membership = {
  id: string;
  tier_id: string;
  tier_title: string;
  started_at: string;
  expires_at: string;
  cancelled_at: string | null;
  transaction: Transaction;
  slug: string;
  program_id: string;
  program_name: string;
};

export type CheckoutResponse = RedirectUrlResponse;

export type CheckoutRequest = { tier_id: string };
