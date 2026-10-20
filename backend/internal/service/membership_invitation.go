package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/mailer"
	"github.com/ubcesports/memberships/internal/membershippolicy"
	"github.com/ubcesports/memberships/internal/repository"
	"github.com/ubcesports/memberships/internal/util"
)

var (
	ErrMembershipInvitationNotFound    = errors.New("Membership invitation not found")
	ErrMembershipInvitationEmail       = errors.New("A membership invitation already exists for this email")
	ErrMembershipInvitationUserExists  = errors.New("A user with this email already exists; add their membership from the Users panel instead")
	ErrInvalidMembershipInvitation     = errors.New("Invalid membership invitation")
	ErrInvalidMembershipInvitationTier = errors.New("Invalid membership tier")
)

const actionMembershipInvitationCreated = "membership_invitation.created"

type MembershipInvitationService struct {
	repository *repository.MembershipInvitationRepository
}

func NewMembershipInvitationService(
	repository *repository.MembershipInvitationRepository,
) *MembershipInvitationService {
	return &MembershipInvitationService{repository: repository}
}

func (s *MembershipInvitationService) Create(
	ctx context.Context,
	request dto.CreateMembershipInvitationRequest,
	createdByUserID string,
	requestID string,
) (*dto.MembershipInvitationDTO, error) {
	normalized, err := normalizeMembershipInvitation(request)
	if err != nil {
		return nil, s.auditMembershipInvitationCreateFailure(
			ctx, createdByUserID, actionMembershipInvitationCreated, requestID,
			fmt.Sprintf("Failed to create membership invitation for %s", strings.TrimSpace(request.Email)),
			err,
		)
	}
	if err := s.ensureUserDoesNotExist(ctx, normalized.Email); err != nil {
		return nil, s.auditMembershipInvitationCreateFailure(
			ctx, createdByUserID, actionMembershipInvitationCreated, requestID,
			fmt.Sprintf("Failed to create membership invitation for %s", normalized.Email),
			err,
		)
	}

	var invitation *dto.MembershipInvitationDTO
	err = s.repository.WithTx(ctx, func(repo *repository.MembershipInvitationRepository) error {
		invitation, err = repo.Create(ctx, normalized, createdByUserID)
		if err != nil {
			return err
		}

		return repo.CreateAuditLog(ctx, repository.MembershipInvitationAuditLogParams{
			ActorUserID: createdByUserID,
			Action:      actionMembershipInvitationCreated,
			Outcome:     db.AdminAuditOutcomeTypeSuccess,
			RequestID:   requestID,
			Description: fmt.Sprintf(
				"Created membership invitation for %s (%s, $%.2f CAD via %s)",
				invitation.Email,
				invitation.TierTitle,
				float64(invitation.AmountPaidCents)/100,
				invitation.PaymentMethod,
			),
		})
	})
	if err != nil {
		operationErr := membershipInvitationError(err)
		return nil, s.auditMembershipInvitationCreateFailure(
			ctx, createdByUserID, actionMembershipInvitationCreated, requestID,
			fmt.Sprintf("Failed to create membership invitation for %s", normalized.Email),
			operationErr,
		)
	}
	if err := sendMembershipInvitation(invitation, requestID); err != nil {
		return nil, fmt.Errorf("send membership invitation: %w", err)
	}
	return invitation, nil
}

func (s *MembershipInvitationService) GetAll(ctx context.Context) ([]dto.MembershipInvitationDTO, error) {
	return s.repository.GetAll(ctx)
}

func (s *MembershipInvitationService) GetByID(
	ctx context.Context,
	id string,
) (*dto.MembershipInvitationDTO, error) {
	if _, err := util.GetValidatedUUID(id); err != nil {
		return nil, ErrInvalidMembershipInvitation
	}
	invitation, err := s.repository.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMembershipInvitationNotFound
	}
	return invitation, err
}

func (s *MembershipInvitationService) Delete(ctx context.Context, id string) error {
	if _, err := util.GetValidatedUUID(id); err != nil {
		return ErrInvalidMembershipInvitation
	}
	if err := s.repository.Delete(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMembershipInvitationNotFound
		}
		return err
	}
	return nil
}

func (s *MembershipInvitationService) ensureUserDoesNotExist(
	ctx context.Context,
	email string,
) error {
	exists, err := s.repository.UserExistsByEmail(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return ErrMembershipInvitationUserExists
	}
	return nil
}

// RedeemForUser fulfills the invitation matching the onboarded
// user's email. The membership, completed transaction, and done flag are
// committed atomically, making retries safe.
func (s *MembershipInvitationService) RedeemForUser(
	ctx context.Context,
	profile *dto.ProfileDTO,
) (bool, error) {
	redeemed := false
	err := s.repository.WithRedemptionTx(
		ctx,
		func(
			invitationRepo *repository.MembershipInvitationRepository,
			membershipRepo *repository.MembershipRepository,
		) error {
			invitation, err := invitationRepo.GetByEmailForUpdate(ctx, profile.Email)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			if err != nil {
				return err
			}

			expiresAt, err := membershippolicy.MembershipExpiresAt(
				invitation.PurchasedAt,
				invitation.ExpirationType,
			)
			if err != nil {
				return err
			}

			transactionID, err := membershipRepo.CreatePendingTransaction(
				ctx,
				repository.CreatePendingTransactionParams{
					UserId:            profile.ID,
					TierId:            invitation.TierID,
					GroupAtPurchase:   getGroupAtPurchase(profile.Groups),
					StudentAtPurchase: profile.IsStudent,
					PurchaseType:      dto.PurchaseNew,
					PaymentMethod:     invitation.PaymentMethod,
				},
			)
			if err != nil {
				return err
			}

			if err := membershipRepo.CancelActiveMembershipsByUserIdAndProgramId(
				ctx,
				profile.ID,
				invitation.ProgramID,
				invitation.PurchasedAt,
			); err != nil {
				return err
			}

			membershipID, err := membershipRepo.CreateMembership(
				ctx,
				repository.CreateMembershipParams{
					UserId:    profile.ID,
					TierId:    invitation.TierID,
					StartedAt: invitation.PurchasedAt,
					ExpiresAt: expiresAt,
				},
			)
			if err != nil {
				return err
			}

			if err := membershipRepo.CompleteTransaction(
				ctx,
				repository.CompleteTransactionParams{
					TransactionId:         transactionID,
					MembershipId:          membershipID,
					StripePaymentIntentId: "",
					AmountPaidCents:       invitation.AmountPaidCents,
				},
			); err != nil {
				return err
			}

			if err := invitationRepo.MarkDone(ctx, invitation.ID); err != nil {
				return err
			}

			redeemed = true
			return nil
		},
	)
	return redeemed, err
}

func normalizeMembershipInvitation(
	request dto.CreateMembershipInvitationRequest,
) (dto.CreateMembershipInvitationRequest, error) {
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.TierID = strings.TrimSpace(request.TierID)

	address, err := mail.ParseAddress(request.Email)
	if err != nil || !strings.EqualFold(address.Address, request.Email) {
		return request, ErrInvalidMembershipInvitation
	}
	if request.AmountPaidCents < 0 {
		return request, ErrInvalidMembershipInvitation
	}
	if request.PaymentMethod != dto.PaymentMethodCash && request.PaymentMethod != dto.PaymentMethodEtransfer {
		return request, ErrOfflinePaymentMethod
	}
	if _, err := util.GetValidatedUUID(request.TierID); err != nil {
		return request, ErrInvalidMembershipInvitationTier
	}

	return request, nil
}

func membershipInvitationError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrMembershipInvitationNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrMembershipInvitationEmail
		case "23503":
			return ErrInvalidMembershipInvitationTier
		case "23514", "23502":
			return ErrInvalidMembershipInvitation
		}
	}
	return err
}

func (s *MembershipInvitationService) auditMembershipInvitationCreateFailure(
	ctx context.Context,
	actorUserID string,
	action string,
	requestID string,
	description string,
	operationErr error,
) error {
	auditErr := s.repository.CreateAuditLog(ctx, repository.MembershipInvitationAuditLogParams{
		ActorUserID: actorUserID,
		Action:      action,
		Outcome:     db.AdminAuditOutcomeTypeFailed,
		RequestID:   requestID,
		Description: description,
	})
	if auditErr != nil {
		return errors.Join(operationErr, auditErr)
	}
	return operationErr
}

func sendMembershipInvitation(
	invitation *dto.MembershipInvitationDTO,
	requestID string,
) error {
	html, err := mailer.RenderEmail(mailer.EmailData{
		Title:      "Redeem your UBCEA membership",
		Heading:    "Welcome to UBCEA!",
		Subheading: "Thanks for purchasing a membership in person. If you would like to redeem it, sign up for a UBCEA account using the email address below.",
		Rows: mailer.NewRows(
			"Email", invitation.Email,
			"Membership", invitation.TierTitle,
		),
		CTAText: "Sign up to redeem",
		CTAURL:  mailer.FrontendURL() + "/login",
	})
	if err != nil {
		return err
	}

	mailer.SendEmailAsync(
		[]string{invitation.Email},
		"Claim your UBCEA membership",
		html,
		requestID,
		invitation.ID,
	)
	return nil
}
