"use client";

import { Loader2, MailPlus } from "lucide-react";
import { useId, useState, type FormEvent } from "react";
import { toast } from "sonner";
import { ActionButton } from "@/components/action-button";
import type {
  AdminMembershipTierOption,
  CreateMembershipInvitationRequest,
  OfflinePaymentMethod,
} from "@/lib/types/admin.types";

const FIELD_CLASS_NAME =
  "h-11 w-full border border-brand-border bg-brand-surface px-3 text-sm text-brand-text outline-none transition placeholder:text-brand-text-subtle focus:border-brand-primary focus:ring-2 focus:ring-brand-primary/30 disabled:cursor-not-allowed disabled:opacity-60";

type MembershipInvitationFormProps = {
  tierOptions: AdminMembershipTierOption[];
  isLoadingTiers: boolean;
  isSubmitting: boolean;
  onSubmit: (request: CreateMembershipInvitationRequest) => Promise<void>;
};

export function MembershipInvitationForm({
  tierOptions,
  isLoadingTiers,
  isSubmitting,
  onSubmit,
}: MembershipInvitationFormProps) {
  const formID = useId();
  const [email, setEmail] = useState("");
  const [tierID, setTierID] = useState("");
  const [amountPaid, setAmountPaid] = useState("");
  const [paymentMethod, setPaymentMethod] = useState<OfflinePaymentMethod | "">("");

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const dollars = Number(amountPaid);
    if (!Number.isFinite(dollars) || dollars < 0) {
      toast.error("Enter a valid amount paid");
      return;
    }
    if (!paymentMethod) {
      toast.error("Select a payment method");
      return;
    }

    try {
      await onSubmit({
        email: email.trim(),
        tier_id: tierID,
        amount_paid_cents: Math.round(dollars * 100),
        payment_method: paymentMethod,
      });
    } catch {
      // The shared API client displays the server error.
      return;
    }

    setEmail("");
    setTierID("");
    setAmountPaid("");
    setPaymentMethod("");
  };

  return (
    <div className="border-b border-brand-border px-5 py-5 sm:px-6">
      <div className="border-l-2 border-brand-primary bg-white/2.5 p-4 sm:p-5">
        <div className="mb-5 flex items-start gap-3">
          <span className="flex size-9 shrink-0 items-center justify-center border border-brand-primary/40 bg-brand-primary/15 text-brand-text">
            <MailPlus aria-hidden="true" className="size-4" />
          </span>
          <div>
            <h2 className="text-sm font-semibold text-brand-text">Record an in-person purchase</h2>
            <p className="mt-1 text-sm text-brand-text-subtle">
              Saving the purchase immediately emails the member their signup link.
            </p>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="grid gap-4 md:grid-cols-2 xl:grid-cols-7">
          <label className="flex flex-col gap-1.5 text-sm text-brand-text-subtle xl:col-span-2">
            <span>Email</span>
            <input
              id={`${formID}-email`}
              required
              type="email"
              autoComplete="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              placeholder="member@example.com"
              disabled={isSubmitting}
              className={FIELD_CLASS_NAME}
            />
          </label>

          <label className="flex flex-col gap-1.5 text-sm text-brand-text-subtle xl:col-span-2">
            <span>Membership</span>
            <select
              id={`${formID}-tier`}
              required
              value={tierID}
              onChange={(event) => setTierID(event.target.value)}
              disabled={isSubmitting || isLoadingTiers || tierOptions.length === 0}
              className={FIELD_CLASS_NAME}
            >
              <option value="">
                {isLoadingTiers
                  ? "Loading memberships…"
                  : tierOptions.length === 0
                    ? "No memberships available"
                    : "Select a membership"}
              </option>
              {tierOptions.map((tier) => (
                <option key={tier.id} value={tier.id}>
                  {tier.program_name} — {tier.title}
                </option>
              ))}
            </select>
          </label>

          <label className="flex flex-col gap-1.5 text-sm text-brand-text-subtle">
            <span>Amount paid</span>
            <div className="relative">
              <span className="pointer-events-none absolute inset-y-0 left-3 flex items-center text-sm text-brand-text-subtle">
                $
              </span>
              <input
                id={`${formID}-amount`}
                required
                type="number"
                inputMode="decimal"
                min="0"
                step="0.01"
                value={amountPaid}
                onChange={(event) => setAmountPaid(event.target.value)}
                placeholder="0.00"
                disabled={isSubmitting}
                className={`${FIELD_CLASS_NAME} pl-7`}
              />
            </div>
          </label>

          <label className="flex flex-col gap-1.5 text-sm text-brand-text-subtle">
            <span>Payment</span>
            <select
              id={`${formID}-payment`}
              required
              value={paymentMethod}
              onChange={(event) =>
                setPaymentMethod(event.target.value as OfflinePaymentMethod | "")
              }
              disabled={isSubmitting}
              className={FIELD_CLASS_NAME}
            >
              <option value="">Select a payment method</option>
              <option value="cash">Cash</option>
              <option value="etransfer">E-transfer</option>
            </select>
          </label>

          <div className="flex items-end md:col-span-2 xl:col-span-1">
            <ActionButton
              type="submit"
              loading={isSubmitting}
              disabled={isLoadingTiers || tierOptions.length === 0}
              icon={<MailPlus aria-hidden="true" className="size-4" />}
              loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
              className="h-11 w-full border-brand-primary bg-brand-primary px-4 text-white hover:border-brand-primary-hover hover:bg-brand-primary-hover"
            >
              {isSubmitting ? "Saving" : "Add & send invite"}
            </ActionButton>
          </div>
        </form>
      </div>
    </div>
  );
}
