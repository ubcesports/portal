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
	GroupPresident       GroupType = "president"
)

type ExecDisplayGroupType string

const (
	ExecDisplayGroupTypePresident       ExecDisplayGroupType = "president"
	ExecDisplayGroupTypeBoard           ExecDisplayGroupType = "board"
	ExecDisplayGroupTypeCentralDirector ExecDisplayGroupType = "central_director"
	ExecDisplayGroupTypeGameDirector    ExecDisplayGroupType = "game_director"
	ExecDisplayGroupTypeExecutive       ExecDisplayGroupType = "executive"
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

type ExecSocialPlatformType string

const (
	ExecSocialPlatformInstagram ExecSocialPlatformType = "instagram"
	ExecSocialPlatformX         ExecSocialPlatformType = "x"
	ExecSocialPlatformTwitch    ExecSocialPlatformType = "twitch"
	ExecSocialPlatformYoutube   ExecSocialPlatformType = "youtube"
	ExecSocialPlatformTiktok    ExecSocialPlatformType = "tiktok"
	ExecSocialPlatformLinkedIn  ExecSocialPlatformType = "linkedin"
)
