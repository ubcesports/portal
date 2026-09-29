package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/stripe/stripe-go/v86"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/mailer"
	"github.com/ubcesports/memberships/internal/membershippolicy"
	"github.com/ubcesports/memberships/internal/repository"
	"github.com/ubcesports/memberships/internal/stripeclient"
	"github.com/ubcesports/memberships/internal/util"
)

type MembershipService struct {
	membershipRepo     *repository.MembershipRepository
	stripeClient       *stripeclient.Client
	profileService     *ProfileService
	eligibilityService *membershippolicy.EligibilityService
}

func NewMembershipService(membershipRepo *repository.MembershipRepository, stripeClient *stripeclient.Client, profileService *ProfileService, eligibilityService *membershippolicy.EligibilityService) *MembershipService {
	return &MembershipService{membershipRepo: membershipRepo, stripeClient: stripeClient, profileService: profileService, eligibilityService: eligibilityService}
}

/*
	Consts
*/

var (
	ErrMembershipAlreadyExists    = errors.New("An active un-upgradeable membership already exists!")
	ErrTierNotEligible            = errors.New("Requested membership tier not eligible for current user. Please try another value.")
	ErrTierNotFound               = errors.New("Tier with given tier id not found.")
	ErrMembershipPurchaseClosed   = errors.New("Membership purchases are closed until the next membership period.")
	ErrPendingCheckoutAlreadyPaid = errors.New("A previous checkout payment is still being processed. Please wait a moment and refresh or contact an admin.")
	ErrOfflinePaymentMethod       = errors.New("Manual membership addition only accepts cash or etransfer.")
	ErrInvalidMembershipTier      = errors.New("A valid membership tier ID is required.")
)

const actionMembershipAdded = "user.membership.added"

/*
	Public functions
*/

func (s *MembershipService) GetPublicTiersAndPrices(ctx context.Context) ([]dto.MembershipTierDTO, error) {
	tiers, err := s.membershipRepo.GetPublicTiersAndPrices(ctx)
	if err != nil {
		return nil, err
	}

	returnTiers := make([]dto.MembershipTierDTO, 0, len(tiers))
	tierIndexById := make(map[string]int)

	for _, tier := range tiers {
		tierId := tier.ID.String()
		if !tier.PriceInCents.Valid {
			return nil, fmt.Errorf("membership tier price %s has no database price", tier.StripePriceID.String)
		}

		// Set up price dto
		var isStudentRequired *bool
		if tier.IsStudentRequired.Valid {
			isStudentRequired = &tier.IsStudentRequired.Bool
		} else {
			isStudentRequired = nil
		}

		priceDto := dto.MembershipTierPriceDTO{
			Price:             float64(tier.PriceInCents.Int64) / 100,
			PriceId:           tier.StripePriceID.String,
			IsStudentRequired: isStudentRequired,
		}

		if index, exists := tierIndexById[tierId]; exists {
			// If tier already exists in return body, append price to it
			returnTiers[index].Prices = append(returnTiers[index].Prices, priceDto)
		} else {
			// If tier doesn't exist in return body, add a new tier with price
			tierIndexById[tierId] = len(returnTiers)
			returnTiers = append(returnTiers, dto.MembershipTierDTO{
				ID:             tier.ID.String(),
				Title:          tier.Title,
				Description:    tier.Description.String,
				Benefits:       tier.Benefits,
				Limitations:    tier.Limitations,
				Slug:           tier.Slug.String,
				ProductId:      tier.StripeProductID.String,
				Prices:         []dto.MembershipTierPriceDTO{priceDto},
				ProgramId:      tier.ProgramID.String(),
				ProgramName:    tier.ProgramName,
				ExpirationType: dto.MembershipExpirationType(tier.ExpirationType),
			})
		}
	}

	return returnTiers, nil
}

func (s *MembershipService) GetCurrentMembershipsWithTransactions(ctx context.Context, userId string) ([]dto.MembershipDTO, error) {
	return s.membershipRepo.GetCurrentMembershipsWithTransactions(ctx, userId)
}

func (s *MembershipService) GetAllMembershipsWithTransactions(ctx context.Context, userId string) (*[]dto.MembershipDTO, error) {
	memberships, err := s.membershipRepo.GetAllMembershipsWithTransactions(ctx, userId)
	if err != nil {
		// If user has no current membership, return nil
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	returnMemberships := make([]dto.MembershipDTO, 0, len(memberships))
	for _, membership := range memberships {

		membershipDto := dto.MembershipDTO{
			ID:          membership.ID.String(),
			TierId:      membership.TierID.String(),
			TierTitle:   membership.TierTitle,
			Slug:        membership.Slug.String,
			StartedAt:   membership.StartedAt.Time,
			ExpiresAt:   membership.ExpiresAt.Time,
			CancelledAt: util.TimestampPointer(membership.CancelledAt),
			ProgramName: membership.ProgramName,
			ProgramId:   membership.ProgramID.String(),
			Transaction: dto.TransactionDTO{
				ID:                    membership.TransactionID.String(),
				AmountPaid:            fmt.Sprintf("%.2f", float64(membership.AmountPaidCents.Int64)/100),
				Status:                dto.TransactionStatusType(membership.Status),
				GroupAtPurchase:       dto.GroupType(membership.GroupAtPurchase.GroupType),
				StudentAtPurchase:     membership.StudentAtPurchase.Bool,
				StripePaymentIntentId: membership.StripePaymentIntentID.String,
				PurchaseType:          dto.PurchaseType(membership.PurchaseType.PurchaseType),
				PaymentMethod:         dto.PaymentMethodType(membership.PaymentMethod),
			},
		}
		returnMemberships = append(returnMemberships, membershipDto)
	}

	return &returnMemberships, nil
}

func (s *MembershipService) GetEligibleTiersWithPrices(ctx context.Context, userId string) ([]dto.EligibleMembershipTierDTO, error) {
	return s.eligibilityService.GetEligibleTiers(ctx, userId)
}

func (s *MembershipService) CreateCheckoutSession(ctx context.Context, userId string, req dto.CheckoutSessionRequest) (*dto.CheckoutSessionResponse, error) {
	// 1. Let the centralized eligibility service determine whether the
	// requested tier can be purchased, its purchase type, and final price.
	eligibleTiers, err := s.eligibilityService.GetEligibleTiers(
		ctx,
		userId,
	)
	if err != nil {
		return nil, err
	}

	// 2. Check if user is eligible for this tier. GetEligibleTiers already
	// accounts for the purchase window (an otherwise-eligible tier while
	// purchases are closed comes back with Eligible: false), so there's no
	// separate closed-window check needed here.
	var selectedTier *dto.EligibleMembershipTierDTO
	for i := range eligibleTiers {
		if eligibleTiers[i].ID == req.TierId && eligibleTiers[i].Eligible {
			selectedTier = &eligibleTiers[i]
			break
		}
	}
	if selectedTier == nil {
		return nil, ErrTierNotEligible
	}

	// Get user profile to create checkout session with their email
	profile, err := s.profileService.GetProfileByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}

	// 3. If there is a pending transaction, then expire it and its stripe checkout session
	var transactionId string
	err = s.membershipRepo.WithTx(ctx, func(mr *repository.MembershipRepository) error {
		pending, err := mr.GetPendingTransactionForUpdate(ctx, userId)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		if err == nil {
			if pending.StripeCheckoutSessionID.Valid && pending.StripeCheckoutSessionID.String != "" {
				_, expireErr := s.stripeClient.ExpireCheckoutSession(ctx, pending.StripeCheckoutSessionID.String)
				if expireErr != nil {
					session, getErr := s.stripeClient.GetCheckoutSession(ctx, pending.StripeCheckoutSessionID.String)
					if getErr != nil {
						return expireErr
					}

					if session.Status == stripe.CheckoutSessionStatusComplete {
						return ErrPendingCheckoutAlreadyPaid
					}

					if session.Status != stripe.CheckoutSessionStatusExpired {
						return expireErr
					}
				}
			}

			if err := mr.ExpirePendingTransactionById(ctx, pending.ID.String()); err != nil {
				return err
			}
		}

		transactionId, err = mr.CreatePendingTransaction(ctx, repository.CreatePendingTransactionParams{
			UserId:            userId,
			TierId:            selectedTier.ID,
			GroupAtPurchase:   getGroupAtPurchase(profile.Groups),
			StudentAtPurchase: profile.IsStudent,
			PurchaseType:      selectedTier.PurchaseType,
			PaymentMethod:     dto.PaymentMethodStripe,
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	// 4. Create new pending transaction and checkout session

	// Create stripe checkout session
	session, err := s.stripeClient.CreateCheckoutSession(ctx, stripeclient.CheckoutSessionRequest{
		TransactionID: transactionId,
		UserID:        userId,
		CustomerEmail: profile.Email,
		PriceID:       selectedTier.Price.PriceId,
		ProductID:     selectedTier.ProductId,
		AmountInCents: int64(math.Round(selectedTier.Price.Price * 100)),
		Currency:      "cad", // Always in canadian dollars
		IsUpgrade:     selectedTier.PurchaseType == dto.PurchaseUpgrade,
	})
	if err != nil {
		markFailedErr := s.membershipRepo.UpdateTransactionStatusById(ctx, transactionId, dto.TransactionFailed)
		if markFailedErr != nil {
			return nil, fmt.Errorf("create checkout session failed: %w; also failed to mark transaction failed: %v", err, markFailedErr)
		}
		return nil, err
	}

	// Put stripe checkout session id into pending transaction
	rowsAffected, err := s.membershipRepo.PutStripeCheckoutSessionId(ctx, transactionId, session.ID)
	if err != nil {
		_, expireErr := s.stripeClient.ExpireCheckoutSession(ctx, session.ID)
		markFailedErr := s.membershipRepo.UpdateTransactionStatusById(ctx, transactionId, dto.TransactionFailed)
		if expireErr != nil || markFailedErr != nil {
			return nil, fmt.Errorf("save checkout session id failed: %w; expire checkout session failed: %v; mark transaction failed failed: %v", err, expireErr, markFailedErr)
		}
		return nil, err
	}

	if rowsAffected != 1 {
		_, expireErr := s.stripeClient.ExpireCheckoutSession(ctx, session.ID)
		if expireErr != nil {
			return nil, fmt.Errorf(
				"transaction is no longer pending; also failed to expire checkout session: %w",
				expireErr,
			)
		}

		return nil, fmt.Errorf("transaction is no longer pending")
	}

	return &dto.CheckoutSessionResponse{Url: session.URL}, nil
}

func (s *MembershipService) AddMembershipToUser(ctx context.Context, actorId string, targetUserId string, requestId string, addMembershipRequest dto.AdminAddMembershipToUserRequest) (returnErr error) {
	var fulfilled *fulfilledPurchase

	defer func() {
		if returnErr == nil {
			return
		}

		auditErr := s.membershipRepo.CreateMembershipAuditLog(ctx, repository.CreateMembershipAuditLogParams{
			ActorUserId:  actorId,
			TargetUserId: targetUserId,
			Action:       actionMembershipAdded,
			Outcome:      db.AdminAuditOutcomeTypeFailed,
			RequestId:    requestId,
			Description:  fmt.Sprintf("Failed to add membership tier %s using %s payment", addMembershipRequest.TierId, addMembershipRequest.PaymentMethod),
		})
		if auditErr != nil {
			returnErr = errors.Join(returnErr, auditErr)
		}
	}()

	if addMembershipRequest.PaymentMethod != dto.PaymentMethodCash && addMembershipRequest.PaymentMethod != dto.PaymentMethodEtransfer {
		return ErrOfflinePaymentMethod
	}

	if _, err := util.GetValidatedUUID(addMembershipRequest.TierId); err != nil {
		return ErrInvalidMembershipTier
	}

	// 1. Let the centralized eligibility service determine whether the
	// requested tier can be purchased, its purchase type, and final price.
	eligibleTiers, err := s.eligibilityService.GetEligibleTiers(
		ctx,
		targetUserId,
	)
	if err != nil {
		return err
	}

	// 2. Check if user is eligible for this tier. GetEligibleTiers already
	// accounts for the purchase window (an otherwise-eligible tier while
	// purchases are closed comes back with Eligible: false), so there's no
	// separate closed-window check needed here.
	var selectedTier *dto.EligibleMembershipTierDTO
	for i := range eligibleTiers {
		if eligibleTiers[i].ID == addMembershipRequest.TierId && eligibleTiers[i].Eligible {
			selectedTier = &eligibleTiers[i]
			break
		}
	}
	if selectedTier == nil {
		return ErrTierNotEligible
	}

	// Get user profile to create checkout session with their email
	profile, err := s.profileService.GetProfileByUserID(ctx, targetUserId)
	if err != nil {
		return err
	}

	purchasedAt := time.Now()
	amountPaidCents := int64(math.Round(selectedTier.Price.Price * 100))

	// 3. If there is a pending transaction, then expire it
	err = s.membershipRepo.WithTx(ctx, func(mr *repository.MembershipRepository) error {
		pending, err := mr.GetPendingTransactionForUpdate(ctx, targetUserId)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		if err == nil {
			if pending.StripeCheckoutSessionID.Valid && pending.StripeCheckoutSessionID.String != "" {
				_, expireErr := s.stripeClient.ExpireCheckoutSession(ctx, pending.StripeCheckoutSessionID.String)
				if expireErr != nil {
					session, getErr := s.stripeClient.GetCheckoutSession(ctx, pending.StripeCheckoutSessionID.String)
					if getErr != nil {
						return expireErr
					}

					if session.Status == stripe.CheckoutSessionStatusComplete {
						return ErrPendingCheckoutAlreadyPaid
					}

					if session.Status != stripe.CheckoutSessionStatusExpired {
						return expireErr
					}
				}
			}

			if err := mr.ExpirePendingTransactionById(ctx, pending.ID.String()); err != nil {
				return err
			}
		}

		// 4. Create new membership and transaction

		// Create new pending transaction
		transactionId, err := mr.CreatePendingTransaction(ctx, repository.CreatePendingTransactionParams{
			UserId:            targetUserId,
			TierId:            selectedTier.ID,
			GroupAtPurchase:   getGroupAtPurchase(profile.Groups),
			StudentAtPurchase: profile.IsStudent,
			PurchaseType:      selectedTier.PurchaseType,
			PaymentMethod:     addMembershipRequest.PaymentMethod,
		})
		if err != nil {
			return err
		}

		transaction, err := mr.GetTransactionByTransactionIdForUpdate(ctx, transactionId)
		if err != nil {
			return err
		}

		if err := mr.CancelActiveMembershipsByUserIdAndProgramId(
			ctx,
			transaction.UserID.String(),
			transaction.ProgramID.String(),
			purchasedAt,
		); err != nil {
			return err
		}

		// 5. Create the fulfilled membership.
		expiresAt, err := membershippolicy.MembershipExpiresAt(purchasedAt, dto.MembershipExpirationType(transaction.ExpirationType))
		if err != nil {
			return err
		}
		membershipId, err := mr.CreateMembership(ctx, repository.CreateMembershipParams{
			UserId:    transaction.UserID.String(),
			TierId:    transaction.TierID.String(),
			StartedAt: purchasedAt,
			ExpiresAt: expiresAt,
		})
		if err != nil {
			return err
		}

		// 6. Record payment details and mark the transaction completed.
		if err := mr.CompleteTransaction(ctx, repository.CompleteTransactionParams{
			TransactionId:         transaction.ID.String(),
			MembershipId:          membershipId,
			StripePaymentIntentId: "",
			AmountPaidCents:       amountPaidCents,
		}); err != nil {
			return err
		}

		if err := mr.CreateMembershipAuditLog(ctx, repository.CreateMembershipAuditLogParams{
			ActorUserId:  actorId,
			TargetUserId: targetUserId,
			Action:       actionMembershipAdded,
			Outcome:      db.AdminAuditOutcomeTypeSuccess,
			RequestId:    requestId,
			Description: fmt.Sprintf(
				"Added %s membership using %s payment ($%.2f CAD)",
				selectedTier.Title,
				addMembershipRequest.PaymentMethod,
				float64(amountPaidCents)/100,
			),
		}); err != nil {
			return err
		}

		fulfilled = &fulfilledPurchase{
			userId:       transaction.UserID.String(),
			tierId:       transaction.TierID.String(),
			purchaseType: transaction.PurchaseType.PurchaseType,
			amountCents:  amountPaidCents,
		}
		return nil
	})
	if err != nil {
		return err
	}

	if fulfilled != nil {
		s.sendPurchaseEmail(ctx, *fulfilled)
	}
	return nil
}

/*
	Stripe webhook callback functions
*/

// fulfilledPurchase is the info needed to email a user after a checkout
// session has been fulfilled inside HandleCheckoutPaid's transaction.
type fulfilledPurchase struct {
	userId       string
	tierId       string
	purchaseType db.PurchaseType
	amountCents  int64
}

func (s *MembershipService) HandleCheckoutPaid(ctx context.Context, session *stripe.CheckoutSession, occurredAt time.Time) error {
	var fulfilled *fulfilledPurchase

	err := s.membershipRepo.WithTx(ctx, func(mr *repository.MembershipRepository) error {
		// 1. Lock the transaction for this checkout session so duplicate webhooks cannot fulfill it twice.
		transaction, err := mr.GetTransactionByCheckoutSessionIdForUpdate(ctx, session.ID)
		if err != nil {
			return err
		}

		// 2. Ensure membership purchase is not closed
		isClosed, err := membershippolicy.IsPurchaseClosed(occurredAt, dto.MembershipExpirationType(transaction.ExpirationType))
		if err != nil {
			return err
		}
		if isClosed {
			return mr.ExpirePendingTransactionById(ctx, transaction.ID.String())
		}

		// 3. Make the handler idempotent: Stripe can retry or duplicate webhook delivery.
		if transaction.Status == db.TransactionStatusTypeCompleted {
			return nil
		}
		if transaction.Status != db.TransactionStatusTypePending {
			return nil
		}

		// 4. Confirm Stripe says this checkout is paid before fulfilling the membership.
		if session.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
			return fmt.Errorf("checkout session %s is not paid", session.ID)
		}

		// 5. Cross-check Stripe metadata against the locked DB transaction.
		if session.Metadata["transaction_id"] != transaction.ID.String() {
			return fmt.Errorf("checkout session %s transaction metadata mismatch", session.ID)
		}
		if session.Metadata["user_id"] != transaction.UserID.String() {
			return fmt.Errorf("checkout session %s user metadata mismatch", session.ID)
		}

		// 6. Cancel any old active membership before creating the new fulfilled membership.
		if err := mr.CancelActiveMembershipsByUserIdAndProgramId(ctx, transaction.UserID.String(), transaction.ProgramID.String(), occurredAt); err != nil {
			return err
		}

		// 7. Create the fulfilled membership.
		expiresAt, err := membershippolicy.MembershipExpiresAt(occurredAt, dto.MembershipExpirationType(transaction.ExpirationType))
		if err != nil {
			return err
		}
		membershipId, err := mr.CreateMembership(ctx, repository.CreateMembershipParams{
			UserId:    transaction.UserID.String(),
			TierId:    transaction.TierID.String(),
			StartedAt: occurredAt,
			ExpiresAt: expiresAt,
		})
		if err != nil {
			return err
		}

		// 8. Record payment details and mark the transaction completed.
		var paymentIntentId string
		if session.PaymentIntent != nil {
			paymentIntentId = session.PaymentIntent.ID
		}
		if err := mr.CompleteTransaction(ctx, repository.CompleteTransactionParams{
			TransactionId:         transaction.ID.String(),
			MembershipId:          membershipId,
			StripePaymentIntentId: paymentIntentId,
			AmountPaidCents:       session.AmountTotal,
		}); err != nil {
			return err
		}

		fulfilled = &fulfilledPurchase{
			userId:       transaction.UserID.String(),
			tierId:       transaction.TierID.String(),
			purchaseType: transaction.PurchaseType.PurchaseType,
			amountCents:  session.AmountTotal,
		}
		return nil
	})
	if err != nil {
		return err
	}

	if fulfilled != nil {
		s.sendPurchaseEmail(ctx, *fulfilled)
	}
	return nil
}

// sendPurchaseEmail emails the user after a checkout session has been
// fulfilled. The membership itself is already saved by this point, so a
// failure here is logged and swallowed rather than surfaced to the caller.
func (s *MembershipService) sendPurchaseEmail(ctx context.Context, fulfilled fulfilledPurchase) {
	tier, err := s.getTierByTierId(ctx, fulfilled.tierId)
	if err != nil {
		slog.Error("send purchase email: get tier failed", "error", err, "user_id", fulfilled.userId)
		return
	}

	profile, err := s.profileService.GetProfileByUserID(ctx, fulfilled.userId)
	if err != nil {
		slog.Error("send purchase email: get profile failed", "error", err, "user_id", fulfilled.userId)
		return
	}

	data := purchaseEmail(tier.Title, fulfilled.purchaseType, fulfilled.amountCents)

	html, err := mailer.RenderEmail(data)
	if err != nil {
		slog.Error("render purchase email failed", "error", err, "user_id", fulfilled.userId)
		return
	}

	mailer.SendEmailAsync(
		[]string{profile.Email},
		data.Heading,
		html,
		middleware.GetReqID(ctx),
		fulfilled.userId,
	)
}

// purchaseEmail builds the content for the purchase/upgrade confirmation
// email. purchaseType distinguishes a brand-new membership from an upgrade
// of an existing one.
func purchaseEmail(tierTitle string, purchaseType db.PurchaseType, amountPaidCents int64) mailer.EmailData {
	heading := "Your membership is confirmed 🎉"
	subheading := "Thanks for your purchase! Here's a summary of your new membership."
	transactionTypeLabel := "New"
	if purchaseType == db.PurchaseTypeUpgrade {
		heading = "Your membership was upgraded"
		subheading = "Here's a summary of your upgrade."
		transactionTypeLabel = "Upgrade"
	}

	return mailer.EmailData{
		Title:      heading,
		Heading:    heading,
		Subheading: subheading,
		Rows: mailer.NewRows(
			"Tier", tierTitle,
			"Transaction type", transactionTypeLabel,
			"Amount paid", fmt.Sprintf("$%.2f CAD", float64(amountPaidCents)/100),
		),
		CTAText: "View your membership",
		CTAURL:  mailer.FrontendURL() + "/profile",
	}
}

// RunExpiryNotifications emails members whose active membership expires in
// exactly one week, and members whose active membership expires today.
//
// This checks two exact dates rather than a rolling per-user window, because
// membership expiry dates come from a small set of fixed calendar cutoffs
// (see membershippolicy.MembershipExpiresAt: end of semester, end of school
// year, or end of the purchase day) rather than N days from purchase —
// checking a range instead would re-send the same email on every day of
// that window. A "day" tier's expiry is always earlier than the +7 check
// reaches it, so it will only ever produce the "expired" email, never the
// "expiring soon" one — that's expected, not a bug.
//
// Intended to be called once per day by a scheduler; failures are logged
// and swallowed per-membership so one bad row doesn't block the rest.
func (s *MembershipService) RunExpiryNotifications(ctx context.Context) {
	location, err := vancouverLocation()
	if err != nil {
		slog.Error("expiry notifications: load location failed", "error", err)
		return
	}

	now := time.Now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)

	s.sendExpiryEmails(ctx, today.AddDate(0, 0, 7), expiringSoonEmail)
	s.sendExpiryEmails(ctx, today.AddDate(0, 0, -1), expiredEmail)
}

func (s *MembershipService) HandleCheckoutExpired(ctx context.Context, sessionId string) error {
	return s.membershipRepo.UpdatePendingTransactionStatusByCheckoutId(ctx, sessionId, dto.TransactionExpired)
}

func (s *MembershipService) HandleCheckoutFailed(ctx context.Context, sessionId string) error {
	return s.membershipRepo.UpdatePendingTransactionStatusByCheckoutId(ctx, sessionId, dto.TransactionFailed)
}

// expiryEmailContent builds an expiry-related email's content for one
// membership row.
type expiryEmailContent func(row db.GetActiveMembershipsExpiringOnDateRow) mailer.EmailData

func expiringSoonEmail(row db.GetActiveMembershipsExpiringOnDateRow) mailer.EmailData {
	heading := "Your membership expires in 1 week"
	return mailer.EmailData{
		Title:      heading,
		Heading:    heading,
		Subheading: "Your membership is expiring soon. Renew now to keep your access without any interruption.",
		Rows: mailer.NewRows(
			"Tier", row.TierTitle,
			"Expires on", formatVancouverDate(row.ExpiresAt.Time),
		),
		CTAText: "Renew your membership",
		CTAURL:  mailer.FrontendURL() + "/pricing",
	}
}

// formatVancouverDate renders t as a calendar date in America/Vancouver
// time. t is stored as a UTC timestamp, so formatting it directly (without
// converting first) can show the wrong day — eg. an expiry stored as
// "April 30 23:59:59 Vancouver" is "May 1, ~07:00 UTC".
func formatVancouverDate(t time.Time) string {
	location, err := vancouverLocation()
	if err != nil {
		location = time.UTC
	}
	return t.In(location).Format("January 2, 2006")
}

func expiredEmail(row db.GetActiveMembershipsExpiringOnDateRow) mailer.EmailData {
	heading := "Your membership has expired"
	return mailer.EmailData{
		Title:      heading,
		Heading:    heading,
		Subheading: "Your membership has expired. Renew anytime to regain access.",
		Rows:       mailer.NewRows("Tier", row.TierTitle),
		CTAText:    "Renew your membership",
		CTAURL:     mailer.FrontendURL() + "/pricing",
	}
}

// sendExpiryEmails emails every active membership expiring on date using the
// given content builder.
func (s *MembershipService) sendExpiryEmails(ctx context.Context, date time.Time, content expiryEmailContent) {
	rows, err := s.membershipRepo.GetActiveMembershipsExpiringOnDate(ctx, date)
	if err != nil {
		slog.Error("expiry notifications: query failed", "error", err, "date", date.Format(time.DateOnly))
		return
	}

	for _, row := range rows {
		data := content(row)

		html, err := mailer.RenderEmail(data)
		if err != nil {
			slog.Error("expiry notifications: render email failed", "error", err, "membership_id", row.ID.String())
			continue
		}

		mailer.SendEmailAsync([]string{row.Email}, data.Heading, html, "", row.UserID.String())
	}
}

/*
	Private functions
*/

// vancouverLocation returns the time zone all membership-year and expiry
// calculations are performed in, regardless of the caller's local time zone.
func vancouverLocation() (*time.Location, error) {
	return time.LoadLocation("America/Vancouver")
}

// returns the highest priority group a user belongs to at the time of membership purchase.
//
// Group priority (highest to lowest):
//   - Board
//   - Director
//   - Executive
//   - Competitive Team
//   - Member (default)
//
// If a user belongs to multiple groups, the highest priority group is used.
func getGroupAtPurchase(groups []dto.GroupType) dto.GroupType {
	if slices.Contains(groups, dto.GroupBoard) {
		return dto.GroupBoard
	}
	if slices.Contains(groups, dto.GroupDirector) {
		return dto.GroupDirector
	}
	if slices.Contains(groups, dto.GroupExecutive) {
		return dto.GroupExecutive
	}
	if slices.Contains(groups, dto.GroupCompetitiveTeam) {
		return dto.GroupCompetitiveTeam
	}
	return dto.GroupMember
}

func (s *MembershipService) getTierByTierId(ctx context.Context, tierId string) (*dto.MembershipTierDTO, error) {
	tier, err := s.membershipRepo.GetTierByTierId(ctx, tierId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTierNotFound
		}

		return nil, err
	}

	return &dto.MembershipTierDTO{
		ID:          tier.ID.String(),
		Title:       tier.Title,
		Description: tier.Description.String,
		Benefits:    tier.Benefits,
		Limitations: tier.Limitations,
		Slug:        tier.Slug.String,
		ProductId:   tier.StripeProductID.String,
		Prices:      []dto.MembershipTierPriceDTO{},
		ProgramId:   tier.ProgramID.String(),
		ProgramName: tier.ProgramName,
	}, nil
}
