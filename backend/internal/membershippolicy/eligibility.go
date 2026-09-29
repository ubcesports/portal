package membershippolicy

import (
	"context"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/repository"
)

type ProfileReader interface {
	GetProfileByUserID(
		ctx context.Context,
		userID string,
	) (*dto.ProfileDTO, error)
}

type EligibilityService struct {
	membershipRepo *repository.MembershipRepository
	profileReader  ProfileReader
	policies       *Registry
}

func NewEligibilityService(
	membershipRepo *repository.MembershipRepository,
	profileReader ProfileReader,
	policies *Registry,
) *EligibilityService {
	return &EligibilityService{membershipRepo: membershipRepo, profileReader: profileReader, policies: policies}
}

/*
	Public functions
*/

func (s *EligibilityService) GetEligibleTiers(ctx context.Context, userId string) ([]dto.EligibleMembershipTierDTO, error) {
	now := time.Now()

	// Get user info
	user, err := s.profileReader.GetProfileByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}

	tiers, err := s.membershipRepo.GetActiveTiersWithPrices(ctx)
	if err != nil {
		return nil, err
	}

	memberships, err := s.membershipRepo.GetCurrentMembershipsWithTransactions(ctx, userId)
	if err != nil {
		return nil, err
	}

	result := make([]dto.EligibleMembershipTierDTO, 0, len(tiers))
	for i := range tiers {
		tier := &tiers[i]

		// 1. Tiers a user's group can't even see stay fully hidden (eg. a
		// regular member never sees the Executive tier at all), rather than
		// showing up with a reason.
		if !isUserGroupEligible(user.Groups, tier.RequiredGroup) {
			continue
		}

		base := dto.EligibleMembershipTierDTO{
			ID:             tier.ID,
			Title:          tier.Title,
			Description:    tier.Description,
			Slug:           tier.Slug,
			ProductId:      tier.ProductId,
			Benefits:       tier.Benefits,
			Limitations:    tier.Limitations,
			ProgramId:      tier.ProgramId,
			ProgramName:    tier.ProgramName,
			ExpirationType: tier.ExpirationType,
		}

		// 2. Check if user alr has a membership for the current program
		current, err := findCurrentMembershipForProgram(memberships, tier.ProgramId)
		if err != nil {
			return nil, err
		}

		// 3. Check policies for current program based on user's membership.
		// This takes priority over the purchase window: a tier that's
		// already owned, or blocked by exec/competitive status, should say
		// so rather than a possibly-irrelevant "purchase opens on" date.
		policy, err := s.policies.Get(tier.ProgramName)
		if err != nil {
			return nil, err
		}

		evaluation, err := policy.Evaluate(user, current, tier)
		if err != nil {
			return nil, err
		}
		if !evaluation.Allowed {
			base.UnavailableReason = evaluation.Reason
			result = append(result, base)
			continue
		}

		// 4. Otherwise-eligible tiers are unavailable while the purchase
		// window is closed.
		closed, err := IsPurchaseClosed(now, tier.ExpirationType)
		if err != nil {
			return nil, err
		}
		if closed {
			opensAt, err := NextPurchaseOpenDate(now, tier.ExpirationType)
			if err != nil {
				return nil, err
			}
			base.UnavailableReason = dto.ReasonPurchaseClosed
			base.PurchaseOpensAt = &opensAt
			result = append(result, base)
			continue
		}

		// 5. Get user's personalized price
		price, err := getTierPriceForUser(user.IsStudent, *tier)
		if err != nil {
			return nil, err
		}
		if price == nil {
			base.UnavailableReason = dto.ReasonUnavailable
			result = append(result, base)
			continue
		}

		finalPrice := price.Price

		// 6. Adjust price for upgrades
		if evaluation.PurchaseType == dto.PurchaseUpgrade {
			if current == nil {
				return nil, fmt.Errorf("upgrade not allowed without a membership in this program")
			}

			targetCents := int64(math.Round(price.Price * 100))
			amountDueCents := targetCents - current.Transaction.AmountPaidCents
			if amountDueCents < 0 {
				amountDueCents = 0
			}

			finalPrice = float64(amountDueCents) / 100
		}

		base.Eligible = true
		base.PurchaseType = evaluation.PurchaseType
		base.Price = &dto.MembershipTierPriceDTO{
			Price:             finalPrice,
			PriceId:           price.PriceId,
			IsStudentRequired: nil,
		}
		result = append(result, base)
	}

	return result, nil
}

// memberships can expire in one of the following ways (all times in America/Vancouver regardless of the purchaser's local time zone):
//   - year:     expires at the end of the current school year
//   - semester: expires at the end of the current semester
//   - day:      expires at the end of the current day'
//
// semester 1 starts at September 1 00:00:00 and ends on December 31 23:59:59
// semester 2 starts at January 1 00:00:00 and ends on April 30 23:59:59
//
// Expiry rules for year:
//   - Purchases made from January 1 through April 30 expire on
//     April 30 23:59:59 of the same calendar year.
//   - Purchases made on or after May 1 expire on
//     April 30 23:59:59 of the following calendar year.
//
// Expiry rules for semester:
//   - Purchases made from September 1 through December 31 expire on
//     December 31 23:59:59 of the same calendar year.
//   - Purchases made on or after January 1 expire on
//     April 30 23:59:59 of the same calendar year.
//
// Expiry rules for day:
//   - Purchase must expire on the same day at 23:59:59

func MembershipExpiresAt(
	purchasedAt time.Time,
	expirationType dto.MembershipExpirationType,
) (time.Time, error) {
	location, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		return time.Time{}, err
	}

	localPurchasedAt := purchasedAt.In(location)
	year := localPurchasedAt.Year()

	switch expirationType {
	case dto.MembershipExpirationYear:
		if localPurchasedAt.Month() >= time.May {
			year++
		}

		return time.Date(
			year,
			time.April,
			30,
			23, 59, 59,
			0,
			location,
		), nil

	case dto.MembershipExpirationSemester:
		switch localPurchasedAt.Month() {
		case time.January, time.February, time.March, time.April:
			return time.Date(
				year,
				time.April,
				30,
				23, 59, 59,
				0,
				location,
			), nil

		case time.September, time.October, time.November, time.December:
			return time.Date(
				year,
				time.December,
				31,
				23, 59, 59,
				0,
				location,
			), nil

		default:
			return time.Time{}, fmt.Errorf(
				"cannot calculate semester expiration outside an active semester: %s",
				localPurchasedAt.Format(time.DateOnly),
			)
		}

	case dto.MembershipExpirationDay:
		return time.Date(
			year,
			localPurchasedAt.Month(),
			localPurchasedAt.Day(),
			23, 59, 59,
			0,
			location,
		), nil

	default:
		return time.Time{}, fmt.Errorf(
			"unsupported membership expiration type %q",
			expirationType,
		)
	}
}

// IsPurchaseClosed reports whether purchasing a membership is currently closed.
// All closure boundaries are evaluated in the America/Vancouver timezone.
//
// The start date is inclusive and the end date is exclusive.
//
// Yearly memberships:
//   - Closed from April 28 00:00:00 through April 30 23:59:59.
//
// Semesterly memberships:
//   - Closed from December 30 00:00 through December 31 23:59:59.
//   - Closed from April 28 00:00 through August 31 23:59:59.
//
// Day memberships:
//   - Closed from May 1 00:00:00 through August 31 23:59:59.
func IsPurchaseClosed(
	now time.Time,
	expirationType dto.MembershipExpirationType,
) (bool, error) {
	location, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		return false, err
	}

	localNow := now.In(location)
	year := localNow.Year()

	isWithin := func(start, end time.Time) bool {
		return !localNow.Before(start) && localNow.Before(end)
	}

	switch expirationType {
	case dto.MembershipExpirationYear:
		return isWithin(
			time.Date(year, time.April, 28, 0, 0, 0, 0, location),
			time.Date(year, time.May, 1, 0, 0, 0, 0, location),
		), nil

	case dto.MembershipExpirationSemester:
		winterClosure :=
			localNow.Month() == time.December &&
				localNow.Day() >= 30

		summerClosure := isWithin(
			time.Date(year, time.April, 28, 0, 0, 0, 0, location),
			time.Date(year, time.September, 1, 0, 0, 0, 0, location),
		)

		return winterClosure || summerClosure, nil

	case dto.MembershipExpirationDay:
		return isWithin(
			time.Date(year, time.May, 1, 0, 0, 0, 0, location),
			time.Date(year, time.September, 1, 0, 0, 0, 0, location),
		), nil

	default:
		return false, fmt.Errorf(
			"unsupported membership expiration type %q",
			expirationType,
		)
	}
}

// NextPurchaseOpenDate reports when a currently-closed purchase window
// reopens, for display on the pricing page. Only meaningful to call when
// IsPurchaseClosed is true for the same (now, expirationType) pair.
func NextPurchaseOpenDate(
	now time.Time,
	expirationType dto.MembershipExpirationType,
) (time.Time, error) {
	location, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		return time.Time{}, err
	}

	localNow := now.In(location)
	year := localNow.Year()

	switch expirationType {
	case dto.MembershipExpirationYear:
		// Only ever closed April 28-30, reopening May 1 the same year.
		return time.Date(year, time.May, 1, 0, 0, 0, 0, location), nil

	case dto.MembershipExpirationSemester:
		// Closed either in the last days of December (reopens New Year's
		// Day) or from April 28 through August (reopens September 1).
		if localNow.Month() == time.December {
			return time.Date(year+1, time.January, 1, 0, 0, 0, 0, location), nil
		}
		return time.Date(year, time.September, 1, 0, 0, 0, 0, location), nil

	case dto.MembershipExpirationDay:
		return time.Date(year, time.September, 1, 0, 0, 0, 0, location), nil

	default:
		return time.Time{}, fmt.Errorf(
			"unsupported membership expiration type %q",
			expirationType,
		)
	}
}

/*
	Helper functions
*/

func isUserGroupEligible(
	userGroups []dto.GroupType,
	requiredGroup dto.GroupType,
) bool {
	if slices.Contains(userGroups, requiredGroup) {
		return true
	}

	if requiredGroup == dto.GroupExecutive {
		return slices.Contains(userGroups, dto.GroupDirector) || slices.Contains(userGroups, dto.GroupBoard)
	}

	return false
}

func getTierPriceForUser(
	studentStatus bool,
	reqTier dto.MembershipTierDTO,
) (*dto.MembershipTierPriceDTO, error) {
	var match *dto.MembershipTierPriceDTO

	for i := range reqTier.Prices {
		currPrice := &reqTier.Prices[i]

		eligible := currPrice.IsStudentRequired == nil || *currPrice.IsStudentRequired == studentStatus
		if !eligible {
			continue
		}

		match = currPrice
		break
	}

	return match, nil
}

func findCurrentMembershipForProgram(
	memberships []dto.MembershipDTO,
	programId string,
) (*dto.MembershipDTO, error) {
	var match *dto.MembershipDTO

	for i := range memberships {
		membership := &memberships[i]

		if membership.ProgramId != programId {
			continue
		}

		if match != nil {
			return nil, fmt.Errorf(
				"multiple active memberships found for program %q",
				membership.ProgramName,
			)
		}

		match = membership
	}

	return match, nil
}
