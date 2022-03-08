package persist

import (
	"context"
	"database/sql/driver"
	"encoding/json"
)

// MembershipTier represents the membership tier of a user
type MembershipTier struct {
	Version      NullInt32         `json:"version"` // schema version for this model
	ID           DBID              `json:"id" binding:"required"`
	CreationTime CreationTime      `json:"created_at"`
	Deleted      NullBool          `json:"-"`
	LastUpdated  LastUpdatedTime   `json:"last_updated"`
	Name         NullString        `json:"name"`
	TokenID      TokenID           `json:"token_id"`
	AssetURL     NullString        `json:"asset_url"`
	Owners       []MembershipOwner `json:"owners"`
}

// MembershipOwner represents a user who owns a membership card
type MembershipOwner struct {
	UserID      DBID         `json:"user_id"`
	Address     Address      `json:"address"`
	Username    NullString   `json:"username"`
	PreviewNFTs []NullString `json:"preview_nfts"`
}

// MembershipRepository represents the interface for interacting with the persisted state of users
type MembershipRepository interface {
	UpsertByTokenID(context.Context, TokenID, MembershipTier) error
	GetByTokenID(context.Context, TokenID) (MembershipTier, error)
	GetAll(context.Context) ([]MembershipTier, error)
}

// Value implements the database/sql/driver Valuer interface for the membership owner type
func (o MembershipOwner) Value() (driver.Value, error) {
	return json.Marshal(o)
}

// Scan implements the database/sql Scanner interface for the membership owner type
func (o *MembershipOwner) Scan(src interface{}) error {
	if src == nil {
		*o = MembershipOwner{}
		return nil
	}
	return json.Unmarshal(src.([]uint8), o)
}

// ErrMembershipNotFoundByTokenID represents an error when a membership is not found by token id
type ErrMembershipNotFoundByTokenID struct {
	TokenID TokenID
}

func (e ErrMembershipNotFoundByTokenID) Error() string {
	return "membership not found by token id: " + e.TokenID.String()
}
