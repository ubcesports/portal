package dto

type RoleType string

const (
	RoleMember RoleType = "member"
	RoleAdmin  RoleType = "admin"
)

type GroupType string

const (
	GroupMember          GroupType = "member"
	GroupCompetitiveTeam GroupType = "competitive_team"
	GroupExecutive       GroupType = "executive"
	GroupDirector        GroupType = "director"
	GroupBoard           GroupType = "board"
)

type TransactionStatusType string

const (
	TransactionPending   TransactionStatusType = "pending"
	TransactionCompleted TransactionStatusType = "completed"
	TransactionFailed    TransactionStatusType = "failed"
	TransactionRefunded  TransactionStatusType = "refunded"
	TransactionExpired   TransactionStatusType = "expired"
)

type PurchaseType string

const (
	PurchaseNew     PurchaseType = "new"
	PurchaseUpgrade PurchaseType = "upgrade"
)

type AdminAuditLogOutcomeType string

const (
	AuditLogSuccess AdminAuditLogOutcomeType = "success"
	AuditLogFailed  AdminAuditLogOutcomeType = "failed"
	AuditLogDenied  AdminAuditLogOutcomeType = "denied"
)

type MembershipExpirationType string

const (
	MembershipExpirationSemester MembershipExpirationType = "semester"
	MembershipExpirationYear     MembershipExpirationType = "year"
	MembershipExpirationDay      MembershipExpirationType = "day"
)

type PaymentMethodType string

const (
	PaymentMethodStripe    PaymentMethodType = "stripe"
	PaymentMethodCash      PaymentMethodType = "cash"
	PaymentMethodEtransfer PaymentMethodType = "etransfer"
)

type TierUnavailableReason string

const (
	ReasonAlreadyOwned          TierUnavailableReason = "already_owned"
	ReasonNotEligibleCurrent    TierUnavailableReason = "not_eligible_current_membership"
	ReasonExecutiveRestricted   TierUnavailableReason = "executive_restricted"
	ReasonCompetitiveRestricted TierUnavailableReason = "competitive_restricted"
	ReasonPurchaseClosed        TierUnavailableReason = "purchase_closed"
	ReasonUnavailable           TierUnavailableReason = "unavailable"
)
