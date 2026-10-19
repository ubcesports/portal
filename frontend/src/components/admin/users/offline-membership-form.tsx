"use client";

import { Loader2, Plus, RefreshCw } from "lucide-react";
import { useMemo, useState, type FormEvent } from "react";
import { toast } from "sonner";
import { ActionButton } from "@/components/action-button";
import { SurfacePanel } from "@/components/surface-panel";
import { useAddOfflineMembership, useAdminEligibleMemberships } from "@/lib/admin/admin.hook";
import type { OfflinePaymentMethod } from "@/lib/types/admin.types";
import { formatMembershipPrice } from "@/components/membership/pricing";
import { titleCase } from "@/lib/utils/groups";

const FIELD_CLASS_NAME =
  "h-11 w-full border border-brand-border bg-brand-surface px-3 text-sm text-brand-text outline-none transition focus:border-brand-primary focus:ring-2 focus:ring-brand-primary/30 disabled:cursor-not-allowed disabled:opacity-60";

const PAYMENT_METHODS: { value: OfflinePaymentMethod; label: string }[] = [
  { value: "cash", label: "Cash" },
  { value: "etransfer", label: "E-transfer" },
];

type OfflineMembershipFormProps = {
  userId: string;
};

export function OfflineMembershipForm({ userId }: OfflineMembershipFormProps) {
  const [tierId, setTierId] = useState("");
  const [amountPaid, setAmountPaid] = useState("");
  const [paymentMethod, setPaymentMethod] = useState<OfflinePaymentMethod | "">("");
  const {
    data: eligibleMemberships = [],
    isPending: isLoadingEligibility,
    isError: eligibilityFailed,
    isFetching: isRefreshingEligibility,
    refetch: refetchEligibility,
  } = useAdminEligibleMemberships(userId);
  const { mutateAsync: addMembership, isPending: isAdding } = useAddOfflineMembership(userId);

  const selectedTier = useMemo(
    () => eligibleMemberships.find((tier) => tier.id === tierId),
    [eligibleMemberships, tierId],
  );

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (!selectedTier || !paymentMethod) {
      toast.error("Select a membership and payment method");
      return;
    }

    const dollars = Number(amountPaid);
    if (amountPaid.trim() === "" || !Number.isFinite(dollars) || dollars < 0) {
      toast.error("Enter a valid amount paid");
      return;
    }

    try {
      await addMembership({
        tier_id: selectedTier.id,
        amount_paid_cents: Math.round(dollars * 100),
        payment_method: paymentMethod,
      });
      toast.success(`${selectedTier.title} membership added`, {
        description: `${formatMembershipPrice(dollars)} paid via ${PAYMENT_METHODS.find(({ value }) => value === paymentMethod)?.label}.`,
      });
      setTierId("");
      setAmountPaid("");
      setPaymentMethod("");
    } catch {
      // The shared API client displays the server error.
    }
  };

  return (
    <SurfacePanel className="overflow-hidden bg-transparent">
      <div className="flex flex-col gap-3 border-b border-brand-border bg-brand-primary/10 px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-base font-semibold text-brand-text">Add membership</h2>
          </div>
          <p className="mt-1 text-sm text-brand-text-subtle">
            Grant an eligible membership after confirming payment was received.
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="p-5">
        {eligibilityFailed ? (
          <div
            role="alert"
            className="flex flex-col gap-3 border border-red-400/30 bg-red-400/10 p-4 sm:flex-row sm:items-center sm:justify-between"
          >
            <p className="text-sm text-brand-text-muted">
              Eligible memberships could not be loaded.
            </p>
            <ActionButton
              onClick={() => void refetchEligibility()}
              loading={isRefreshingEligibility}
              icon={<RefreshCw aria-hidden="true" className="size-4" />}
              loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
            >
              Try again
            </ActionButton>
          </div>
        ) : (
          <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(10rem,0.3fr)_minmax(13rem,0.45fr)_auto] lg:items-end">
            <label className="flex min-w-0 flex-col gap-1.5 text-sm text-brand-text-subtle">
              <span>Eligible membership</span>
              <select
                value={tierId}
                onChange={(event) => setTierId(event.target.value)}
                disabled={isLoadingEligibility || isAdding || eligibleMemberships.length === 0}
                className={FIELD_CLASS_NAME}
              >
                <option value="">
                  {isLoadingEligibility
                    ? "Loading eligible memberships…"
                    : eligibleMemberships.length === 0
                      ? "No eligible memberships"
                      : "Select a membership"}
                </option>
                {eligibleMemberships.map((tier) => (
                  <option key={tier.id} value={tier.id}>
                    {tier.program_name} — {tier.title} — {formatMembershipPrice(tier.prices.price)}
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
                  required
                  type="number"
                  inputMode="decimal"
                  min="0"
                  step="0.01"
                  value={amountPaid}
                  onChange={(event) => setAmountPaid(event.target.value)}
                  placeholder="0.00"
                  disabled={isAdding}
                  className={`${FIELD_CLASS_NAME} pl-7`}
                />
              </div>
            </label>

            <label className="flex flex-col gap-1.5 text-sm text-brand-text-subtle">
              <span>Payment method</span>
              <select
                value={paymentMethod}
                onChange={(event) =>
                  setPaymentMethod(event.target.value as OfflinePaymentMethod | "")
                }
                disabled={isAdding}
                className={FIELD_CLASS_NAME}
              >
                <option value="">Select payment type</option>
                {PAYMENT_METHODS.map((method) => (
                  <option key={method.value} value={method.value}>
                    {method.label}
                  </option>
                ))}
              </select>
            </label>

            <ActionButton
              type="submit"
              disabled={!selectedTier || !amountPaid || !paymentMethod}
              loading={isAdding}
              icon={<Plus aria-hidden="true" className="size-4" />}
              loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
              className="h-11 border-brand-primary bg-brand-primary px-5 hover:border-brand-primary-hover hover:bg-brand-primary-hover"
            >
              Add membership
            </ActionButton>
          </div>
        )}

        {!eligibilityFailed && selectedTier ? (
          <div className="mt-4 grid gap-3 border-l-2 border-brand-primary bg-white/3 px-4 py-3 text-sm sm:grid-cols-3">
            <div>
              <span className="block text-xs uppercase tracking-wide text-brand-text-subtle">
                Program
              </span>
              <span className="text-brand-text">{selectedTier.program_name}</span>
            </div>
            <div>
              <span className="block text-xs uppercase tracking-wide text-brand-text-subtle">
                Eligible price
              </span>
              <span className="text-brand-text">
                {formatMembershipPrice(selectedTier.prices.price)} CAD
              </span>
            </div>
            <div>
              <span className="block text-xs uppercase tracking-wide text-brand-text-subtle">
                Purchase
              </span>
              <span className="text-brand-text">{titleCase(selectedTier.purchase_type)}</span>
            </div>
          </div>
        ) : null}

        {!eligibilityFailed && !isLoadingEligibility && eligibleMemberships.length === 0 && (
          <p className="mt-3 text-sm text-brand-text-muted" role="status">
            This user currently has no membership options they are eligible to receive.
          </p>
        )}
      </form>
    </SurfacePanel>
  );
}
