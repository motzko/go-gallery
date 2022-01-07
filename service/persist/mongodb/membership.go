package mongodb

import (
	"context"

	"github.com/mikeydub/go-gallery/service/persist"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const membershipColName = "membership"

// MembershipRepository is a repository for storing membership information in the database
type MembershipRepository struct {
	membershipsStorage *storage
}

// NewMembershipRepository returns a new instance of a membership repository
func NewMembershipRepository(mgoClient *mongo.Client) *MembershipRepository {
	return &MembershipRepository{
		membershipsStorage: newStorage(mgoClient, 0, galleryDBName, membershipColName),
	}
}

// UpsertByTokenID upserts an membership tier by a given token ID
func (c *MembershipRepository) UpsertByTokenID(pCtx context.Context, pTokenID persist.TokenID, pUpsert persist.MembershipTier) error {

	_, err := c.membershipsStorage.upsert(pCtx, bson.M{
		"token_id": pTokenID,
	}, pUpsert)
	if err != nil {
		return err
	}

	return nil
}

// GetByTokenID returns a membership tier by token ID
func (c *MembershipRepository) GetByTokenID(pCtx context.Context, pTokenID persist.TokenID) (persist.MembershipTier, error) {

	result := []persist.MembershipTier{}
	err := c.membershipsStorage.find(pCtx, bson.M{"token_id": pTokenID}, &result)

	if err != nil {
		return persist.MembershipTier{}, err
	}

	if len(result) < 1 {
		return persist.MembershipTier{}, persist.ErrMembershipNotFoundByTokenID{TokenID: pTokenID}
	}

	if len(result) > 1 {
		logrus.Errorf("found more than one membership tier for token ID: %s", pTokenID)
	}

	return result[0], nil
}

// GetAll returns all membership tiers
func (c *MembershipRepository) GetAll(pCtx context.Context) ([]persist.MembershipTier, error) {

	result := []persist.MembershipTier{}
	err := c.membershipsStorage.find(pCtx, bson.M{}, &result)

	if err != nil {
		return nil, err
	}

	return result, nil
}