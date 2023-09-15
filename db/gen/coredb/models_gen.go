// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0

package coredb

import (
	"database/sql"
	"time"

	"github.com/jackc/pgtype"
	"github.com/mikeydub/go-gallery/service/persist"
)

type Admire struct {
	ID          persist.DBID `json:"id"`
	Version     int32        `json:"version"`
	FeedEventID persist.DBID `json:"feed_event_id"`
	ActorID     persist.DBID `json:"actor_id"`
	Deleted     bool         `json:"deleted"`
	CreatedAt   time.Time    `json:"created_at"`
	LastUpdated time.Time    `json:"last_updated"`
	PostID      persist.DBID `json:"post_id"`
	TokenID     persist.DBID `json:"token_id"`
}

type AlchemySpamContract struct {
	ID        persist.DBID    `json:"id"`
	Chain     persist.Chain   `json:"chain"`
	Address   persist.Address `json:"address"`
	CreatedAt time.Time       `json:"created_at"`
	IsSpam    bool            `json:"is_spam"`
}

type Collection struct {
	ID             persist.DBID                                     `json:"id"`
	Deleted        bool                                             `json:"deleted"`
	OwnerUserID    persist.DBID                                     `json:"owner_user_id"`
	Nfts           persist.DBIDList                                 `json:"nfts"`
	Version        sql.NullInt32                                    `json:"version"`
	LastUpdated    time.Time                                        `json:"last_updated"`
	CreatedAt      time.Time                                        `json:"created_at"`
	Hidden         bool                                             `json:"hidden"`
	CollectorsNote sql.NullString                                   `json:"collectors_note"`
	Name           sql.NullString                                   `json:"name"`
	Layout         persist.TokenLayout                              `json:"layout"`
	TokenSettings  map[persist.DBID]persist.CollectionTokenSettings `json:"token_settings"`
	GalleryID      persist.DBID                                     `json:"gallery_id"`
}

type Comment struct {
	ID          persist.DBID `json:"id"`
	Version     int32        `json:"version"`
	FeedEventID persist.DBID `json:"feed_event_id"`
	ActorID     persist.DBID `json:"actor_id"`
	ReplyTo     persist.DBID `json:"reply_to"`
	Comment     string       `json:"comment"`
	Deleted     bool         `json:"deleted"`
	CreatedAt   time.Time    `json:"created_at"`
	LastUpdated time.Time    `json:"last_updated"`
	PostID      persist.DBID `json:"post_id"`
}

type Community struct {
	ID               persist.DBID          `json:"id"`
	Version          int32                 `json:"version"`
	Name             string                `json:"name"`
	Description      string                `json:"description"`
	CommunityType    persist.CommunityType `json:"community_type"`
	CommunitySubtype string                `json:"community_subtype"`
	CommunityKey     string                `json:"community_key"`
	CreatedAt        time.Time             `json:"created_at"`
	LastUpdated      time.Time             `json:"last_updated"`
	Deleted          bool                  `json:"deleted"`
}

type Contract struct {
	ID                    persist.DBID    `json:"id"`
	Deleted               bool            `json:"deleted"`
	Version               sql.NullInt32   `json:"version"`
	CreatedAt             time.Time       `json:"created_at"`
	LastUpdated           time.Time       `json:"last_updated"`
	Name                  sql.NullString  `json:"name"`
	Symbol                sql.NullString  `json:"symbol"`
	Address               persist.Address `json:"address"`
	CreatorAddress        persist.Address `json:"creator_address"`
	Chain                 persist.Chain   `json:"chain"`
	ProfileBannerUrl      sql.NullString  `json:"profile_banner_url"`
	ProfileImageUrl       sql.NullString  `json:"profile_image_url"`
	BadgeUrl              sql.NullString  `json:"badge_url"`
	Description           sql.NullString  `json:"description"`
	OwnerAddress          persist.Address `json:"owner_address"`
	IsProviderMarkedSpam  bool            `json:"is_provider_marked_spam"`
	ParentID              persist.DBID    `json:"parent_id"`
	OverrideCreatorUserID persist.DBID    `json:"override_creator_user_id"`
}

type ContractCommunityMembership struct {
	ID          persist.DBID `json:"id"`
	Version     int32        `json:"version"`
	ContractID  persist.DBID `json:"contract_id"`
	CommunityID persist.DBID `json:"community_id"`
	CreatedAt   time.Time    `json:"created_at"`
	LastUpdated time.Time    `json:"last_updated"`
	Deleted     bool         `json:"deleted"`
}

type ContractCreator struct {
	ContractID     persist.DBID    `json:"contract_id"`
	CreatorUserID  persist.DBID    `json:"creator_user_id"`
	Chain          persist.Chain   `json:"chain"`
	CreatorAddress persist.Address `json:"creator_address"`
}

type ContractRelevance struct {
	ID    persist.DBID `json:"id"`
	Score int32        `json:"score"`
}

type DevMetadataUser struct {
	UserID          persist.DBID  `json:"user_id"`
	HasEmailAddress persist.Email `json:"has_email_address"`
	Deleted         bool          `json:"deleted"`
}

type EarlyAccess struct {
	Address persist.Address `json:"address"`
}

type Event struct {
	ID             persist.DBID         `json:"id"`
	Version        int32                `json:"version"`
	ActorID        sql.NullString       `json:"actor_id"`
	ResourceTypeID persist.ResourceType `json:"resource_type_id"`
	SubjectID      persist.DBID         `json:"subject_id"`
	UserID         persist.DBID         `json:"user_id"`
	TokenID        persist.DBID         `json:"token_id"`
	CollectionID   persist.DBID         `json:"collection_id"`
	Action         persist.Action       `json:"action"`
	Data           persist.EventData    `json:"data"`
	Deleted        bool                 `json:"deleted"`
	LastUpdated    time.Time            `json:"last_updated"`
	CreatedAt      time.Time            `json:"created_at"`
	GalleryID      persist.DBID         `json:"gallery_id"`
	CommentID      persist.DBID         `json:"comment_id"`
	AdmireID       persist.DBID         `json:"admire_id"`
	FeedEventID    persist.DBID         `json:"feed_event_id"`
	ExternalID     sql.NullString       `json:"external_id"`
	Caption        sql.NullString       `json:"caption"`
	GroupID        sql.NullString       `json:"group_id"`
	PostID         persist.DBID         `json:"post_id"`
}

type ExternalSocialConnection struct {
	ID                persist.DBID `json:"id"`
	Version           int32        `json:"version"`
	SocialAccountType string       `json:"social_account_type"`
	FollowerID        persist.DBID `json:"follower_id"`
	FolloweeID        persist.DBID `json:"followee_id"`
	CreatedAt         time.Time    `json:"created_at"`
	LastUpdated       time.Time    `json:"last_updated"`
	Deleted           bool         `json:"deleted"`
}

type FeedBlocklist struct {
	ID          persist.DBID   `json:"id"`
	UserID      persist.DBID   `json:"user_id"`
	Action      persist.Action `json:"action"`
	LastUpdated time.Time      `json:"last_updated"`
	CreatedAt   time.Time      `json:"created_at"`
	Deleted     bool           `json:"deleted"`
}

type FeedEntity struct {
	ID             persist.DBID `json:"id"`
	FeedEntityType int32        `json:"feed_entity_type"`
	CreatedAt      time.Time    `json:"created_at"`
	ActorID        persist.DBID `json:"actor_id"`
}

type FeedEntityScore struct {
	ID             persist.DBID     `json:"id"`
	CreatedAt      time.Time        `json:"created_at"`
	ActorID        persist.DBID     `json:"actor_id"`
	Action         persist.Action   `json:"action"`
	ContractIds    persist.DBIDList `json:"contract_ids"`
	Interactions   int32            `json:"interactions"`
	FeedEntityType int32            `json:"feed_entity_type"`
	LastUpdated    time.Time        `json:"last_updated"`
}

type FeedEntityScoreView struct {
	ID             persist.DBID     `json:"id"`
	CreatedAt      time.Time        `json:"created_at"`
	ActorID        persist.DBID     `json:"actor_id"`
	Action         persist.Action   `json:"action"`
	ContractIds    persist.DBIDList `json:"contract_ids"`
	Interactions   int32            `json:"interactions"`
	FeedEntityType int32            `json:"feed_entity_type"`
	LastUpdated    time.Time        `json:"last_updated"`
}

type FeedEvent struct {
	ID          persist.DBID          `json:"id"`
	Version     int32                 `json:"version"`
	OwnerID     persist.DBID          `json:"owner_id"`
	Action      persist.Action        `json:"action"`
	Data        persist.FeedEventData `json:"data"`
	EventTime   time.Time             `json:"event_time"`
	EventIds    persist.DBIDList      `json:"event_ids"`
	Deleted     bool                  `json:"deleted"`
	LastUpdated time.Time             `json:"last_updated"`
	CreatedAt   time.Time             `json:"created_at"`
	Caption     sql.NullString        `json:"caption"`
	GroupID     sql.NullString        `json:"group_id"`
}

type Follow struct {
	ID          persist.DBID `json:"id"`
	Follower    persist.DBID `json:"follower"`
	Followee    persist.DBID `json:"followee"`
	Deleted     bool         `json:"deleted"`
	CreatedAt   time.Time    `json:"created_at"`
	LastUpdated time.Time    `json:"last_updated"`
}

type Gallery struct {
	ID          persist.DBID     `json:"id"`
	Deleted     bool             `json:"deleted"`
	LastUpdated time.Time        `json:"last_updated"`
	CreatedAt   time.Time        `json:"created_at"`
	Version     sql.NullInt32    `json:"version"`
	OwnerUserID persist.DBID     `json:"owner_user_id"`
	Collections persist.DBIDList `json:"collections"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Hidden      bool             `json:"hidden"`
	Position    string           `json:"position"`
}

type GalleryRelevance struct {
	ID    persist.DBID `json:"id"`
	Score int32        `json:"score"`
}

type LegacyView struct {
	UserID      persist.DBID  `json:"user_id"`
	ViewCount   sql.NullInt32 `json:"view_count"`
	LastUpdated time.Time     `json:"last_updated"`
	CreatedAt   time.Time     `json:"created_at"`
	Deleted     sql.NullBool  `json:"deleted"`
}

type MarketplaceContract struct {
	ContractID persist.DBID `json:"contract_id"`
}

type MediaValidationRule struct {
	ID        persist.DBID `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	MediaType string       `json:"media_type"`
	Property  string       `json:"property"`
	Required  bool         `json:"required"`
}

type Membership struct {
	ID          persist.DBID            `json:"id"`
	Deleted     bool                    `json:"deleted"`
	Version     sql.NullInt32           `json:"version"`
	CreatedAt   time.Time               `json:"created_at"`
	LastUpdated time.Time               `json:"last_updated"`
	TokenID     persist.DBID            `json:"token_id"`
	Name        sql.NullString          `json:"name"`
	AssetUrl    sql.NullString          `json:"asset_url"`
	Owners      persist.TokenHolderList `json:"owners"`
}

type Merch struct {
	ID           persist.DBID    `json:"id"`
	Deleted      bool            `json:"deleted"`
	Version      sql.NullInt32   `json:"version"`
	CreatedAt    time.Time       `json:"created_at"`
	LastUpdated  time.Time       `json:"last_updated"`
	TokenID      persist.TokenID `json:"token_id"`
	ObjectType   int32           `json:"object_type"`
	DiscountCode sql.NullString  `json:"discount_code"`
	Redeemed     bool            `json:"redeemed"`
}

type MigrationValidation struct {
	ID                       persist.DBID   `json:"id"`
	MediaID                  persist.DBID   `json:"media_id"`
	ProcessingJobID          persist.DBID   `json:"processing_job_id"`
	Chain                    persist.Chain  `json:"chain"`
	Contract                 sql.NullString `json:"contract"`
	TokenID                  persist.DBID   `json:"token_id"`
	MediaType                interface{}    `json:"media_type"`
	RemappedTo               interface{}    `json:"remapped_to"`
	OldMedia                 pgtype.JSONB   `json:"old_media"`
	NewMedia                 pgtype.JSONB   `json:"new_media"`
	MediaTypeValidation      string         `json:"media_type_validation"`
	DimensionsValidation     string         `json:"dimensions_validation"`
	MediaUrlValidation       string         `json:"media_url_validation"`
	ThumbnailUrlValidation   string         `json:"thumbnail_url_validation"`
	LivePreviewUrlValidation string         `json:"live_preview_url_validation"`
	LastRefreshed            interface{}    `json:"last_refreshed"`
}

type Nonce struct {
	ID          persist.DBID    `json:"id"`
	Deleted     bool            `json:"deleted"`
	Version     sql.NullInt32   `json:"version"`
	LastUpdated time.Time       `json:"last_updated"`
	CreatedAt   time.Time       `json:"created_at"`
	UserID      persist.DBID    `json:"user_id"`
	Address     persist.Address `json:"address"`
	Value       sql.NullString  `json:"value"`
	Chain       persist.Chain   `json:"chain"`
}

type Notification struct {
	ID          persist.DBID             `json:"id"`
	Deleted     bool                     `json:"deleted"`
	OwnerID     persist.DBID             `json:"owner_id"`
	Version     sql.NullInt32            `json:"version"`
	LastUpdated time.Time                `json:"last_updated"`
	CreatedAt   time.Time                `json:"created_at"`
	Action      persist.Action           `json:"action"`
	Data        persist.NotificationData `json:"data"`
	EventIds    persist.DBIDList         `json:"event_ids"`
	FeedEventID persist.DBID             `json:"feed_event_id"`
	CommentID   persist.DBID             `json:"comment_id"`
	GalleryID   persist.DBID             `json:"gallery_id"`
	Seen        bool                     `json:"seen"`
	Amount      int32                    `json:"amount"`
	PostID      persist.DBID             `json:"post_id"`
	TokenID     persist.DBID             `json:"token_id"`
}

type OwnedContract struct {
	UserID         persist.DBID `json:"user_id"`
	UserCreatedAt  time.Time    `json:"user_created_at"`
	ContractID     persist.DBID `json:"contract_id"`
	OwnedCount     int64        `json:"owned_count"`
	DisplayedCount int64        `json:"displayed_count"`
	Displayed      bool         `json:"displayed"`
	LastUpdated    time.Time    `json:"last_updated"`
}

type PiiAccountCreationInfo struct {
	UserID    persist.DBID `json:"user_id"`
	IpAddress string       `json:"ip_address"`
	CreatedAt time.Time    `json:"created_at"`
}

type PiiForUser struct {
	UserID          persist.DBID    `json:"user_id"`
	PiiEmailAddress persist.Email   `json:"pii_email_address"`
	Deleted         bool            `json:"deleted"`
	PiiSocials      persist.Socials `json:"pii_socials"`
}

type PiiSocialsAuth struct {
	ID           persist.DBID           `json:"id"`
	Deleted      bool                   `json:"deleted"`
	Version      sql.NullInt32          `json:"version"`
	CreatedAt    time.Time              `json:"created_at"`
	LastUpdated  time.Time              `json:"last_updated"`
	UserID       persist.DBID           `json:"user_id"`
	Provider     persist.SocialProvider `json:"provider"`
	AccessToken  sql.NullString         `json:"access_token"`
	RefreshToken sql.NullString         `json:"refresh_token"`
}

type PiiUserView struct {
	ID                   persist.DBID                     `json:"id"`
	Deleted              bool                             `json:"deleted"`
	Version              sql.NullInt32                    `json:"version"`
	LastUpdated          time.Time                        `json:"last_updated"`
	CreatedAt            time.Time                        `json:"created_at"`
	Username             sql.NullString                   `json:"username"`
	UsernameIdempotent   sql.NullString                   `json:"username_idempotent"`
	Wallets              persist.WalletList               `json:"wallets"`
	Bio                  sql.NullString                   `json:"bio"`
	Traits               pgtype.JSONB                     `json:"traits"`
	Universal            bool                             `json:"universal"`
	NotificationSettings persist.UserNotificationSettings `json:"notification_settings"`
	EmailVerified        persist.EmailVerificationStatus  `json:"email_verified"`
	EmailUnsubscriptions persist.EmailUnsubscriptions     `json:"email_unsubscriptions"`
	FeaturedGallery      *persist.DBID                    `json:"featured_gallery"`
	PrimaryWalletID      persist.DBID                     `json:"primary_wallet_id"`
	UserExperiences      pgtype.JSONB                     `json:"user_experiences"`
	PiiEmailAddress      persist.Email                    `json:"pii_email_address"`
	PiiSocials           persist.Socials                  `json:"pii_socials"`
}

type Post struct {
	ID          persist.DBID     `json:"id"`
	Version     int32            `json:"version"`
	TokenIds    persist.DBIDList `json:"token_ids"`
	ContractIds persist.DBIDList `json:"contract_ids"`
	ActorID     persist.DBID     `json:"actor_id"`
	Caption     sql.NullString   `json:"caption"`
	CreatedAt   time.Time        `json:"created_at"`
	LastUpdated time.Time        `json:"last_updated"`
	Deleted     bool             `json:"deleted"`
}

type ProfileImage struct {
	ID           persist.DBID               `json:"id"`
	UserID       persist.DBID               `json:"user_id"`
	TokenID      persist.DBID               `json:"token_id"`
	SourceType   persist.ProfileImageSource `json:"source_type"`
	Deleted      bool                       `json:"deleted"`
	CreatedAt    time.Time                  `json:"created_at"`
	LastUpdated  time.Time                  `json:"last_updated"`
	WalletID     persist.DBID               `json:"wallet_id"`
	EnsAvatarUri sql.NullString             `json:"ens_avatar_uri"`
	EnsDomain    sql.NullString             `json:"ens_domain"`
}

type PushNotificationTicket struct {
	ID               persist.DBID `json:"id"`
	PushTokenID      persist.DBID `json:"push_token_id"`
	TicketID         string       `json:"ticket_id"`
	CreatedAt        time.Time    `json:"created_at"`
	CheckAfter       time.Time    `json:"check_after"`
	NumCheckAttempts int32        `json:"num_check_attempts"`
	Deleted          bool         `json:"deleted"`
	Status           string       `json:"status"`
}

type PushNotificationToken struct {
	ID        persist.DBID `json:"id"`
	UserID    persist.DBID `json:"user_id"`
	PushToken string       `json:"push_token"`
	CreatedAt time.Time    `json:"created_at"`
	Deleted   bool         `json:"deleted"`
}

type RecommendationResult struct {
	ID                persist.DBID  `json:"id"`
	Version           sql.NullInt32 `json:"version"`
	UserID            persist.DBID  `json:"user_id"`
	RecommendedUserID persist.DBID  `json:"recommended_user_id"`
	RecommendedCount  sql.NullInt32 `json:"recommended_count"`
	CreatedAt         time.Time     `json:"created_at"`
	LastUpdated       time.Time     `json:"last_updated"`
	Deleted           bool          `json:"deleted"`
}

type ReprocessJob struct {
	ID           int          `json:"id"`
	TokenStartID persist.DBID `json:"token_start_id"`
	TokenEndID   persist.DBID `json:"token_end_id"`
}

type ScrubbedPiiAccountCreationInfo struct {
	UserID    persist.DBID    `json:"user_id"`
	IpAddress persist.Address `json:"ip_address"`
	CreatedAt time.Time       `json:"created_at"`
}

type ScrubbedPiiForUser struct {
	UserID          persist.DBID    `json:"user_id"`
	PiiEmailAddress persist.Email   `json:"pii_email_address"`
	Deleted         bool            `json:"deleted"`
	PiiSocials      persist.Socials `json:"pii_socials"`
}

type Session struct {
	ID                   persist.DBID `json:"id"`
	UserID               persist.DBID `json:"user_id"`
	CreatedAt            time.Time    `json:"created_at"`
	CreatedWithUserAgent string       `json:"created_with_user_agent"`
	CreatedWithPlatform  string       `json:"created_with_platform"`
	CreatedWithOs        string       `json:"created_with_os"`
	LastRefreshed        time.Time    `json:"last_refreshed"`
	LastUserAgent        string       `json:"last_user_agent"`
	LastPlatform         string       `json:"last_platform"`
	LastOs               string       `json:"last_os"`
	CurrentRefreshID     string       `json:"current_refresh_id"`
	ActiveUntil          time.Time    `json:"active_until"`
	Invalidated          bool         `json:"invalidated"`
	LastUpdated          time.Time    `json:"last_updated"`
	Deleted              bool         `json:"deleted"`
}

type SpamUserScore struct {
	UserID        persist.DBID `json:"user_id"`
	Score         int32        `json:"score"`
	DecidedIsSpam sql.NullBool `json:"decided_is_spam"`
	DecidedAt     sql.NullTime `json:"decided_at"`
	Deleted       bool         `json:"deleted"`
	CreatedAt     time.Time    `json:"created_at"`
}

type Token struct {
	ID                   persist.DBID               `json:"id"`
	Deleted              bool                       `json:"deleted"`
	Version              sql.NullInt32              `json:"version"`
	CreatedAt            time.Time                  `json:"created_at"`
	LastUpdated          time.Time                  `json:"last_updated"`
	Name                 sql.NullString             `json:"name"`
	Description          sql.NullString             `json:"description"`
	CollectorsNote       sql.NullString             `json:"collectors_note"`
	TokenUri             sql.NullString             `json:"token_uri"`
	TokenType            sql.NullString             `json:"token_type"`
	TokenID              persist.TokenID            `json:"token_id"`
	Quantity             persist.HexString          `json:"quantity"`
	OwnershipHistory     persist.AddressAtBlockList `json:"ownership_history"`
	ExternalUrl          sql.NullString             `json:"external_url"`
	BlockNumber          sql.NullInt64              `json:"block_number"`
	OwnerUserID          persist.DBID               `json:"owner_user_id"`
	OwnedByWallets       persist.DBIDList           `json:"owned_by_wallets"`
	Chain                persist.Chain              `json:"chain"`
	Contract             persist.DBID               `json:"contract"`
	IsUserMarkedSpam     sql.NullBool               `json:"is_user_marked_spam"`
	IsProviderMarkedSpam sql.NullBool               `json:"is_provider_marked_spam"`
	LastSynced           time.Time                  `json:"last_synced"`
	FallbackMedia        persist.FallbackMedia      `json:"fallback_media"`
	TokenMediaID         persist.DBID               `json:"token_media_id"`
	IsCreatorToken       bool                       `json:"is_creator_token"`
	IsHolderToken        bool                       `json:"is_holder_token"`
	Displayable          bool                       `json:"displayable"`
}

type TokenCommunityMembership struct {
	ID          persist.DBID `json:"id"`
	Version     int32        `json:"version"`
	TokenID     persist.DBID `json:"token_id"`
	CommunityID persist.DBID `json:"community_id"`
	CreatedAt   time.Time    `json:"created_at"`
	LastUpdated time.Time    `json:"last_updated"`
	Deleted     bool         `json:"deleted"`
}

type TokenMedia struct {
	ID              persist.DBID          `json:"id"`
	CreatedAt       time.Time             `json:"created_at"`
	LastUpdated     time.Time             `json:"last_updated"`
	Version         int32                 `json:"version"`
	ContractID      persist.DBID          `json:"contract_id"`
	TokenID         persist.TokenID       `json:"token_id"`
	Chain           persist.Chain         `json:"chain"`
	Active          bool                  `json:"active"`
	Metadata        persist.TokenMetadata `json:"metadata"`
	Media           persist.Media         `json:"media"`
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	ProcessingJobID persist.DBID          `json:"processing_job_id"`
	Deleted         bool                  `json:"deleted"`
}

type TokenMediasActive struct {
	ID               persist.DBID `json:"id"`
	LastUpdated      time.Time    `json:"last_updated"`
	MediaType        interface{}  `json:"media_type"`
	JobID            persist.DBID `json:"job_id"`
	TokenProperties  pgtype.JSONB `json:"token_properties"`
	PipelineMetadata pgtype.JSONB `json:"pipeline_metadata"`
}

type TokenMediasMissingProperty struct {
	ID          persist.DBID `json:"id"`
	MediaType   interface{}  `json:"media_type"`
	LastUpdated time.Time    `json:"last_updated"`
	IsValid     bool         `json:"is_valid"`
	Reason      []byte       `json:"reason"`
}

type TokenMediasNoValidationRule struct {
	ID          persist.DBID `json:"id"`
	MediaType   interface{}  `json:"media_type"`
	LastUpdated time.Time    `json:"last_updated"`
	IsValid     bool         `json:"is_valid"`
	Reason      string       `json:"reason"`
}

type TokenProcessingJob struct {
	ID               persist.DBID             `json:"id"`
	CreatedAt        time.Time                `json:"created_at"`
	TokenProperties  persist.TokenProperties  `json:"token_properties"`
	PipelineMetadata persist.PipelineMetadata `json:"pipeline_metadata"`
	ProcessingCause  persist.ProcessingCause  `json:"processing_cause"`
	ProcessorVersion string                   `json:"processor_version"`
	Deleted          bool                     `json:"deleted"`
}

type TopRecommendedUser struct {
	RecommendedUserID persist.DBID `json:"recommended_user_id"`
	Frequency         int64        `json:"frequency"`
	LastUpdated       interface{}  `json:"last_updated"`
}

type User struct {
	ID                   persist.DBID                     `json:"id"`
	Deleted              bool                             `json:"deleted"`
	Version              sql.NullInt32                    `json:"version"`
	LastUpdated          time.Time                        `json:"last_updated"`
	CreatedAt            time.Time                        `json:"created_at"`
	Username             sql.NullString                   `json:"username"`
	UsernameIdempotent   sql.NullString                   `json:"username_idempotent"`
	Wallets              persist.WalletList               `json:"wallets"`
	Bio                  sql.NullString                   `json:"bio"`
	Traits               pgtype.JSONB                     `json:"traits"`
	Universal            bool                             `json:"universal"`
	NotificationSettings persist.UserNotificationSettings `json:"notification_settings"`
	EmailVerified        persist.EmailVerificationStatus  `json:"email_verified"`
	EmailUnsubscriptions persist.EmailUnsubscriptions     `json:"email_unsubscriptions"`
	FeaturedGallery      *persist.DBID                    `json:"featured_gallery"`
	PrimaryWalletID      persist.DBID                     `json:"primary_wallet_id"`
	UserExperiences      pgtype.JSONB                     `json:"user_experiences"`
	ProfileImageID       persist.DBID                     `json:"profile_image_id"`
}

type UserRelevance struct {
	ID    persist.DBID `json:"id"`
	Score int32        `json:"score"`
}

type UserRole struct {
	ID          persist.DBID `json:"id"`
	UserID      persist.DBID `json:"user_id"`
	Role        persist.Role `json:"role"`
	Version     int32        `json:"version"`
	Deleted     bool         `json:"deleted"`
	CreatedAt   time.Time    `json:"created_at"`
	LastUpdated time.Time    `json:"last_updated"`
}

type Wallet struct {
	ID          persist.DBID       `json:"id"`
	CreatedAt   time.Time          `json:"created_at"`
	LastUpdated time.Time          `json:"last_updated"`
	Deleted     bool               `json:"deleted"`
	Version     sql.NullInt32      `json:"version"`
	Address     persist.Address    `json:"address"`
	WalletType  persist.WalletType `json:"wallet_type"`
	Chain       persist.Chain      `json:"chain"`
}
