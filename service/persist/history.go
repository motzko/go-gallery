package persist

import (
	"context"
	"time"
)

// OwnershipHistory represents a list of owners for an NFT.
type OwnershipHistory struct {
	Version      NullInt64       `bson:"version"       json:"version"` // schema version for this model
	ID           DBID            `bson:"_id"           json:"id"`
	CreationTime CreationTime    `bson:"created_at" json:"created_at"`
	Deleted      NullBool        `bson:"deleted" json:"-"`
	LastUpdated  LastUpdatedTime `bson:"last_updated" json:"last_updated"`

	NFTID  DBID    `bson:"nft_id" json:"nft_id"`
	Owners []Owner `bson:"owners" json:"owners"`
}

// Owner represents a single owner of an NFT.
type Owner struct {
	Address      Address    `bson:"address" json:"address"`
	UserID       DBID       `bson:"user_id" json:"user_id"`
	Username     NullString `bson:"username" json:"username"`
	TimeObtained time.Time  `bson:"time_obtained" json:"time_obtained"`
}

// OwnershipHistoryRepository is the interface for the OwnershipHistory persistence layer
type OwnershipHistoryRepository interface {
	Upsert(context.Context, DBID, OwnershipHistory) error
}