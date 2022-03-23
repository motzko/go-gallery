package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/mikeydub/go-gallery/service/persist"
	"github.com/sirupsen/logrus"
)

// CollectionRepository is the repository for interacting with collections in a postgres database
type CollectionRepository struct {
	db                           *sql.DB
	galleryRepo                  *GalleryRepository
	createStmt                   *sql.Stmt
	getByUserIDOwnerStmt         *sql.Stmt
	getByUserIDOwnerRawStmt      *sql.Stmt
	getByUserIDStmt              *sql.Stmt
	getByUserIDRawStmt           *sql.Stmt
	getByIDOwnerStmt             *sql.Stmt
	getByIDOwnerRawStmt          *sql.Stmt
	getByIDStmt                  *sql.Stmt
	getByIDRawStmt               *sql.Stmt
	updateInfoStmt               *sql.Stmt
	updateHiddenStmt             *sql.Stmt
	updateNFTsStmt               *sql.Stmt
	nftsToRemoveStmt             *sql.Stmt
	deleteNFTsStmt               *sql.Stmt
	removeNFTFromCollectionsStmt *sql.Stmt
	getNFTsForAddressStmt        *sql.Stmt
	deleteCollectionStmt         *sql.Stmt
	getUserAddressesStmt         *sql.Stmt
	getUnassignedNFTsStmt        *sql.Stmt
	checkOwnNFTsStmt             *sql.Stmt
	getOpenseaIDForNFTStmt       *sql.Stmt
	deleteNFTStmt                *sql.Stmt
	updateOwnerAddressStmt       *sql.Stmt
	getOwnerAddressStmt          *sql.Stmt
}

// NewCollectionRepository creates a new CollectionRepository
func NewCollectionRepository(db *sql.DB, galleryRepo *GalleryRepository) *CollectionRepository {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	createStmt, err := db.PrepareContext(ctx, `INSERT INTO collections (ID, VERSION, NAME, COLLECTORS_NOTE, OWNER_USER_ID, LAYOUT, NFTS, HIDDEN) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING ID;`)
	checkNoErr(err)

	getByUserIDOwnerStmt, err := db.PrepareContext(ctx, `SELECT c.ID,c.OWNER_USER_ID,c.NAME,c.VERSION,c.DELETED,c.COLLECTORS_NOTE,
	c.LAYOUT,c.HIDDEN,c.CREATED_AT,c.LAST_UPDATED,n.ID,n.OWNER_ADDRESS,
	n.MULTIPLE_OWNERS,n.NAME,n.CONTRACT,n.TOKEN_COLLECTION_NAME,n.CREATOR_ADDRESS,n.CREATOR_NAME, 
	n.IMAGE_URL,n.IMAGE_THUMBNAIL_URL,n.IMAGE_PREVIEW_URL,n.ANIMATION_ORIGINAL_URL,n.ANIMATION_URL,n.CREATED_AT 
	FROM collections c, unnest(c.NFTS) WITH ORDINALITY AS u(nft, ordinality)
	LEFT JOIN nfts n ON n.ID = nft
	WHERE c.OWNER_USER_ID = $1 AND c.DELETED = false ORDER BY ordinality;`)
	checkNoErr(err)

	getByUserIDOwnerRawStmt, err := db.PrepareContext(ctx, `SELECT c.ID,c.OWNER_USER_ID,c.NAME,c.VERSION,c.DELETED,c.COLLECTORS_NOTE,c.LAYOUT,c.HIDDEN,c.CREATED_AT,c.LAST_UPDATED FROM collections c WHERE c.OWNER_USER_ID = $1 AND c.DELETED = false;`)
	checkNoErr(err)

	getByUserIDStmt, err := db.PrepareContext(ctx, `SELECT c.ID,c.OWNER_USER_ID,c.NAME,c.VERSION,c.DELETED,c.COLLECTORS_NOTE,
	c.LAYOUT,c.HIDDEN,c.CREATED_AT,c.LAST_UPDATED,n.ID,n.OWNER_ADDRESS,
	n.MULTIPLE_OWNERS,n.NAME,n.CONTRACT,n.TOKEN_COLLECTION_NAME,n.CREATOR_ADDRESS,n.CREATOR_NAME, 
	n.IMAGE_URL,n.IMAGE_THUMBNAIL_URL,n.IMAGE_PREVIEW_URL,n.ANIMATION_ORIGINAL_URL,n.ANIMATION_URL,n.CREATED_AT 
	FROM collections c,unnest(c.NFTS) WITH ORDINALITY AS u(nft, ordinality) 
	LEFT JOIN nfts n ON n.ID = nft 
	WHERE c.OWNER_USER_ID = $1 AND c.HIDDEN = false AND c.DELETED = false ORDER BY ordinality;`)
	checkNoErr(err)
	getByUserIDRawStmt, err := db.PrepareContext(ctx, `SELECT c.ID,c.OWNER_USER_ID,c.NAME,c.VERSION,c.DELETED,c.COLLECTORS_NOTE,c.LAYOUT,c.HIDDEN,c.CREATED_AT,c.LAST_UPDATED FROM collections c WHERE c.OWNER_USER_ID = $1 AND c.DELETED = false AND c.HIDDEN = false;`)
	checkNoErr(err)

	getByIDOwnerStmt, err := db.PrepareContext(ctx, `SELECT c.ID,c.OWNER_USER_ID,c.NAME,c.VERSION,c.DELETED,c.COLLECTORS_NOTE,
	c.LAYOUT,c.HIDDEN,c.CREATED_AT,c.LAST_UPDATED,n.ID,n.OWNER_ADDRESS,
	n.MULTIPLE_OWNERS,n.NAME,n.CONTRACT,n.TOKEN_COLLECTION_NAME,n.CREATOR_ADDRESS,n.CREATOR_NAME, 
	n.IMAGE_URL,n.IMAGE_THUMBNAIL_URL,n.IMAGE_PREVIEW_URL,n.ANIMATION_ORIGINAL_URL,n.ANIMATION_URL,n.CREATED_AT 
	FROM collections c, unnest(c.NFTS) WITH ORDINALITY AS u(nft, ordinality)
	LEFT JOIN nfts n ON n.ID = nft
	WHERE c.ID = $1 AND c.DELETED = false ORDER BY ordinality;`)
	checkNoErr(err)

	getByIDOwnerRawStmt, err := db.PrepareContext(ctx, `SELECT c.ID,c.OWNER_USER_ID,c.NAME,c.VERSION,c.DELETED,c.COLLECTORS_NOTE,c.LAYOUT,c.HIDDEN,c.CREATED_AT,c.LAST_UPDATED FROM collections c WHERE c.ID = $1 AND c.DELETED = false;`)
	checkNoErr(err)

	getByIDStmt, err := db.PrepareContext(ctx, `SELECT c.ID,c.OWNER_USER_ID,c.NAME,c.VERSION,c.DELETED,c.COLLECTORS_NOTE,
	c.LAYOUT,c.HIDDEN,c.CREATED_AT,c.LAST_UPDATED,n.ID,n.OWNER_ADDRESS,
	n.MULTIPLE_OWNERS,n.NAME,n.CONTRACT,n.TOKEN_COLLECTION_NAME,n.CREATOR_ADDRESS,n.CREATOR_NAME,
	n.IMAGE_URL,n.IMAGE_THUMBNAIL_URL,n.IMAGE_PREVIEW_URL,n.ANIMATION_ORIGINAL_URL,n.ANIMATION_URL,n.CREATED_AT 
	FROM collections c, unnest(c.NFTS) WITH ORDINALITY AS u(nft, ordinality)
	LEFT JOIN nfts n ON n.ID = nft
	WHERE c.ID = $1 AND c.HIDDEN = false AND c.DELETED = false ORDER BY ordinality;`)
	checkNoErr(err)

	getByIDRawStmt, err := db.PrepareContext(ctx, `SELECT c.ID,c.OWNER_USER_ID,c.NAME,c.VERSION,c.DELETED,c.COLLECTORS_NOTE,c.LAYOUT,c.HIDDEN,c.CREATED_AT,c.LAST_UPDATED FROM collections c WHERE c.ID = $1 AND c.DELETED = false AND c.HIDDEN = false;`)
	checkNoErr(err)

	updateInfoStmt, err := db.PrepareContext(ctx, `UPDATE collections SET COLLECTORS_NOTE = $1, NAME = $2, LAST_UPDATED = $3 WHERE ID = $4 AND OWNER_USER_ID = $5`)
	checkNoErr(err)

	updateHiddenStmt, err := db.PrepareContext(ctx, `UPDATE collections SET HIDDEN = $1, LAST_UPDATED = $2 WHERE ID = $3 AND OWNER_USER_ID = $4`)
	checkNoErr(err)

	updateNFTsStmt, err := db.PrepareContext(ctx, `UPDATE collections SET NFTS = $1, LAYOUT = $2, LAST_UPDATED = $3 WHERE ID = $4 AND OWNER_USER_ID = $5;`)
	checkNoErr(err)

	nftsToRemoveStmt, err := db.PrepareContext(ctx, `SELECT ID,OPENSEA_ID FROM nfts WHERE OWNER_ADDRESS = ANY($1) AND ID <> ALL($2);`)
	checkNoErr(err)

	deleteNFTsStmt, err := db.PrepareContext(ctx, `UPDATE nfts SET DELETED = true WHERE ID = ANY($1)`)
	checkNoErr(err)

	removeNFTFromCollectionsStmt, err := db.PrepareContext(ctx, `UPDATE collections SET NFTS = array_remove(NFTS, $1) WHERE OWNER_USER_ID = $2;`)
	checkNoErr(err)

	getNFTsForAddressStmt, err := db.PrepareContext(ctx, `SELECT ID FROM nfts WHERE OWNER_ADDRESS = ANY($1)`)
	checkNoErr(err)

	deleteCollectionStmt, err := db.PrepareContext(ctx, `UPDATE collections SET DELETED = true WHERE ID = $1 AND OWNER_USER_ID = $2;`)
	checkNoErr(err)

	getUserAddressesStmt, err := db.PrepareContext(ctx, `SELECT ADDRESSES FROM users WHERE ID = $1;`)
	checkNoErr(err)

	getUnassignedNFTsStmt, err := db.PrepareContext(ctx, `SELECT n.ID,n.CREATED_AT,n.NAME,n.CREATOR_ADDRESS,n.CREATOR_NAME,n.OWNER_ADDRESS,n.MULTIPLE_OWNERS,n.CONTRACT,n.IMAGE_URL,n.IMAGE_THUMBNAIL_URL,n.IMAGE_PREVIEW_URL,n.TOKEN_COLLECTION_NAME 
	FROM nfts n
	JOIN collections c on n.ID <> ALL(c.NFTS)
	WHERE c.OWNER_USER_ID = $1 AND n.OWNER_ADDRESS = ANY($2);`)
	checkNoErr(err)

	checkOwnNFTsStmt, err := db.PrepareContext(ctx, `SELECT COUNT(*) FROM nfts WHERE OWNER_ADDRESS = ANY($1) AND ID = ANY($2);`)
	checkNoErr(err)

	getOpenseaIDForNFTStmt, err := db.PrepareContext(ctx, `SELECT OPENSEA_ID FROM nfts WHERE ID = $1;`)
	checkNoErr(err)

	deleteNFTStmt, err := db.PrepareContext(ctx, `DELETE FROM nfts WHERE ID = $1;`)
	checkNoErr(err)

	updateOwnerAddressStmt, err := db.PrepareContext(ctx, `UPDATE nfts SET OWNER_ADDRESS = $1 WHERE ID = $2;`)
	checkNoErr(err)

	getOwnerAddressStmt, err := db.PrepareContext(ctx, `SELECT OWNER_ADDRESS FROM nfts WHERE ID = $1;`)
	checkNoErr(err)

	return &CollectionRepository{db: db, galleryRepo: galleryRepo, createStmt: createStmt, getByUserIDOwnerStmt: getByUserIDOwnerStmt, getByUserIDStmt: getByUserIDStmt, getByIDOwnerStmt: getByIDOwnerStmt, getByIDStmt: getByIDStmt, updateInfoStmt: updateInfoStmt, updateHiddenStmt: updateHiddenStmt, updateNFTsStmt: updateNFTsStmt, nftsToRemoveStmt: nftsToRemoveStmt, deleteNFTsStmt: deleteNFTsStmt, removeNFTFromCollectionsStmt: removeNFTFromCollectionsStmt, getNFTsForAddressStmt: getNFTsForAddressStmt, deleteCollectionStmt: deleteCollectionStmt, getUserAddressesStmt: getUserAddressesStmt, getUnassignedNFTsStmt: getUnassignedNFTsStmt, checkOwnNFTsStmt: checkOwnNFTsStmt, getByIDOwnerRawStmt: getByIDOwnerRawStmt, getByIDRawStmt: getByIDRawStmt, getByUserIDOwnerRawStmt: getByUserIDOwnerRawStmt, getByUserIDRawStmt: getByUserIDRawStmt, getOpenseaIDForNFTStmt: getOpenseaIDForNFTStmt, deleteNFTStmt: deleteNFTStmt, updateOwnerAddressStmt: updateOwnerAddressStmt, getOwnerAddressStmt: getOwnerAddressStmt}
}

// Create creates a new collection in the database
func (c *CollectionRepository) Create(pCtx context.Context, pColl persist.CollectionDB) (persist.DBID, error) {
	err := ensureNFTsOwnedByUser(pCtx, c, pColl.OwnerUserID, pColl.NFTs)
	if err != nil {
		return "", err
	}

	var id persist.DBID
	err = c.createStmt.QueryRowContext(pCtx, persist.GenerateID(), pColl.Version, pColl.Name, pColl.CollectorsNote, pColl.OwnerUserID, pColl.Layout, pq.Array(pColl.NFTs), pColl.Hidden).Scan(&id)
	if err != nil {
		return "", err
	}
	if err := c.galleryRepo.RefreshCache(pCtx, pColl.OwnerUserID); err != nil {
		return "", err
	}
	return id, nil
}

// GetByUserID returns all collections owned by a user
func (c *CollectionRepository) GetByUserID(pCtx context.Context, pUserID persist.DBID, pShowHidden bool) ([]persist.Collection, error) {
	var stmt *sql.Stmt
	var rawStmt *sql.Stmt
	if pShowHidden {
		stmt = c.getByUserIDOwnerStmt
		rawStmt = c.getByUserIDOwnerRawStmt
	} else {
		stmt = c.getByUserIDStmt
		rawStmt = c.getByUserIDRawStmt
	}
	res, err := stmt.QueryContext(pCtx, pUserID)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	collections := make(map[persist.DBID]persist.Collection)
	for res.Next() {
		var collection persist.Collection
		var nft persist.CollectionNFT
		err = res.Scan(&collection.ID, &collection.OwnerUserID, &collection.Name, &collection.Version, &collection.Deleted, &collection.CollectorsNote, &collection.Layout, &collection.Hidden, &collection.CreationTime, &collection.LastUpdated, &nft.ID, &nft.OwnerAddress, &nft.MultipleOwners, &nft.Name, &nft.Contract, &nft.TokenCollectionName, &nft.CreatorAddress, &nft.CreatorName, &nft.ImageURL, &nft.ImageThumbnailURL, &nft.ImagePreviewURL, &nft.AnimationOriginalURL, &nft.AnimationURL, &nft.CreationTime)
		if err != nil {
			return nil, err
		}

		if coll, ok := collections[collection.ID]; !ok {
			collection.NFTs = []persist.CollectionNFT{nft}
			collections[collection.ID] = collection
		} else {
			coll.NFTs = append(coll.NFTs, nft)
			collections[collection.ID] = coll
		}
	}

	if err := res.Err(); err != nil {
		return nil, err
	}

	result := make([]persist.Collection, 0, len(collections))

	if len(collections) == 0 {
		colls, err := rawStmt.QueryContext(pCtx, pUserID)
		if err != nil {
			return nil, err
		}
		defer colls.Close()
		for colls.Next() {
			var rawColl persist.Collection
			err = colls.Scan(&rawColl.ID, &rawColl.OwnerUserID, &rawColl.Name, &rawColl.Version, &rawColl.Deleted, &rawColl.CollectorsNote, &rawColl.Layout, &rawColl.Hidden, &rawColl.CreationTime, &rawColl.LastUpdated)
			if err != nil {
				return nil, err
			}
			rawColl.NFTs = []persist.CollectionNFT{}
			result = append(result, rawColl)
		}
		if err := colls.Err(); err != nil {
			return nil, err
		}
		return result, nil
	}

	for _, collection := range collections {
		result = append(result, collection)
	}

	return result, nil
}

// GetByID returns a collection by its ID
func (c *CollectionRepository) GetByID(pCtx context.Context, pID persist.DBID, pShowHidden bool) (persist.Collection, error) {
	var stmt *sql.Stmt
	var rawStmt *sql.Stmt
	if pShowHidden {
		stmt = c.getByIDOwnerStmt
		rawStmt = c.getByIDOwnerRawStmt
	} else {
		stmt = c.getByIDStmt
		rawStmt = c.getByIDRawStmt
	}

	res, err := stmt.QueryContext(pCtx, pID)
	if err != nil {
		return persist.Collection{}, err
	}
	defer res.Close()

	var collection persist.Collection
	nfts := make([]persist.CollectionNFT, 0, 10)
	i := 0
	for ; res.Next(); i++ {
		colID := collection.ID
		var nft persist.CollectionNFT
		err = res.Scan(&collection.ID, &collection.OwnerUserID, &collection.Name, &collection.Version, &collection.Deleted, &collection.CollectorsNote, &collection.Layout, &collection.Hidden, &collection.CreationTime, &collection.LastUpdated, &nft.ID, &nft.OwnerAddress, &nft.MultipleOwners, &nft.Name, &nft.Contract, &nft.TokenCollectionName, &nft.CreatorAddress, &nft.CreatorName, &nft.ImageURL, &nft.ImageThumbnailURL, &nft.ImagePreviewURL, &nft.AnimationOriginalURL, &nft.AnimationURL, &nft.CreationTime)
		if err != nil {
			return persist.Collection{}, err
		}
		if colID != "" && colID != collection.ID {
			return persist.Collection{}, errors.New("multiple collections found")
		}

		nfts = append(nfts, nft)
	}
	if err := res.Err(); err != nil {
		return persist.Collection{}, err
	}
	if collection.ID == "" {
		collection.NFTs = []persist.CollectionNFT{}
		err := rawStmt.QueryRowContext(pCtx, pID).Scan(&collection.ID, &collection.OwnerUserID, &collection.Name, &collection.Version, &collection.Deleted, &collection.CollectorsNote, &collection.Layout, &collection.Hidden, &collection.CreationTime, &collection.LastUpdated)
		if err != nil {
			if err == sql.ErrNoRows {
				return persist.Collection{}, persist.ErrCollectionNotFoundByID{ID: pID}
			}
			return persist.Collection{}, err
		}
		if collection.ID != pID {
			return persist.Collection{}, persist.ErrCollectionNotFoundByID{ID: pID}
		}
		return collection, nil
	}

	collection.NFTs = nfts

	return collection, nil
}

// Update updates a collection in the database
func (c *CollectionRepository) Update(pCtx context.Context, pID persist.DBID, pUserID persist.DBID, pUpdate interface{}) error {
	var res sql.Result
	var err error
	switch pUpdate.(type) {
	case persist.CollectionUpdateInfoInput:
		update := pUpdate.(persist.CollectionUpdateInfoInput)
		res, err = c.updateInfoStmt.ExecContext(pCtx, update.CollectorsNote, update.Name, time.Now(), pID, pUserID)
	case persist.CollectionUpdateHiddenInput:
		update := pUpdate.(persist.CollectionUpdateHiddenInput)
		res, err = c.updateHiddenStmt.ExecContext(pCtx, update.Hidden, time.Now(), pID, pUserID)
	default:
		return fmt.Errorf("unsupported update type: %T", pUpdate)
	}
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return persist.ErrCollectionNotFoundByID{ID: pID}
	}
	return c.galleryRepo.RefreshCache(pCtx, pUserID)
}

// UpdateNFTs updates the nfts of a collection in the database
func (c *CollectionRepository) UpdateNFTs(pCtx context.Context, pID persist.DBID, pUserID persist.DBID, pUpdate persist.CollectionUpdateNftsInput) error {

	err := ensureNFTsOwnedByUser(pCtx, c, pUserID, pUpdate.NFTs)
	if err != nil {
		return err
	}

	res, err := c.updateNFTsStmt.ExecContext(pCtx, pq.Array(pUpdate.NFTs), pUpdate.Layout, time.Now(), pID, pUserID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return persist.ErrCollectionNotFoundByID{ID: pID}
	}
	return c.galleryRepo.RefreshCache(pCtx, pUserID)
}

// ClaimNFTs claims nfts from a collection in the database
func (c *CollectionRepository) ClaimNFTs(pCtx context.Context, pUserID persist.DBID, pOwnerAddresses []persist.Address, pUpdate persist.CollectionUpdateNftsInput) error {
	nftsToRemove, err := c.nftsToRemoveStmt.QueryContext(pCtx, pq.Array(pOwnerAddresses), pq.Array(pUpdate.NFTs))
	if err != nil {
		return err
	}
	defer nftsToRemove.Close()

	removing := map[persist.DBID]persist.NullInt64{}
	for nftsToRemove.Next() {
		var id persist.DBID
		var openseaID persist.NullInt64

		err = nftsToRemove.Scan(&id, &openseaID)
		if err != nil {
			return err
		}
		removing[id] = openseaID
	}

	if err := nftsToRemove.Err(); err != nil {
		return err
	}

	newOpenseaIDs := map[persist.DBID]persist.NullInt64{}
	for _, id := range pUpdate.NFTs {
		var openseaID persist.NullInt64
		err := c.getOpenseaIDForNFTStmt.QueryRowContext(pCtx, id).Scan(&openseaID)
		if err != nil {
			return err
		}
		newOpenseaIDs[id] = openseaID
	}

	tx, err := c.db.BeginTx(pCtx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deleteManyStmt := tx.StmtContext(pCtx, c.deleteNFTsStmt)
	deleteStmt := tx.StmtContext(pCtx, c.deleteNFTStmt)
	removeFromCollStmt := tx.StmtContext(pCtx, c.removeNFTFromCollectionsStmt)
	updateOwnerStmt := tx.StmtContext(pCtx, c.updateOwnerAddressStmt)

	nftsToRemoveIDs := make([]persist.DBID, 0, len(removing))

	for removeID, removeOpenseaID := range removing {
		remove := true
		for id, openseaID := range newOpenseaIDs {
			if removeOpenseaID == openseaID {
				remove = false

				var newAddress persist.NullString
				err := c.getOwnerAddressStmt.QueryRowContext(pCtx, id).Scan(&newAddress)
				if err != nil {
					return err
				}

				_, err = deleteStmt.ExecContext(pCtx, id)
				if err != nil {
					return err
				}

				_, err = updateOwnerStmt.ExecContext(pCtx, newAddress, removeID)
				if err != nil {
					return err
				}
			}
		}
		if remove {
			nftsToRemoveIDs = append(nftsToRemoveIDs, removeID)
		}
	}

	_, err = deleteManyStmt.ExecContext(pCtx, pq.Array(nftsToRemoveIDs))
	if err != nil {
		return err
	}

	for _, nft := range nftsToRemoveIDs {
		logrus.Infof("removing nft %s from collections for %s because no longer owns NFT", nft, pUserID)
		_, err := removeFromCollStmt.ExecContext(pCtx, nft, pUserID)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return c.galleryRepo.RefreshCache(pCtx, pUserID)
}

// RemoveNFTsOfAddresses removes nfts of addresses from a collection in the database
func (c *CollectionRepository) RemoveNFTsOfAddresses(pCtx context.Context, pID persist.DBID, pAddresses []persist.Address) error {
	nfts, err := c.getNFTsForAddressStmt.QueryContext(pCtx, pq.Array(pAddresses))
	if err != nil {
		return err
	}
	defer nfts.Close()

	nftsIDs := []persist.DBID{}
	for nfts.Next() {
		var id persist.DBID
		err = nfts.Scan(&id)
		if err != nil {
			return err
		}
		nftsIDs = append(nftsIDs, id)
	}

	if err := nfts.Err(); err != nil {
		return err
	}

	_, err = c.deleteNFTsStmt.ExecContext(pCtx, pq.Array(nftsIDs))
	if err != nil {
		return err
	}

	for _, nft := range nftsIDs {
		_, err = c.removeNFTFromCollectionsStmt.ExecContext(pCtx, nft, pID)
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveNFTsOfOldAddresses removes nfts of addresses that a user no longer has
func (c *CollectionRepository) RemoveNFTsOfOldAddresses(pCtx context.Context, pUserID persist.DBID) error {
	colls, err := c.GetByUserID(pCtx, pUserID, true)
	if err != nil {
		return err
	}

	var addresses []persist.Address
	if err := c.getUserAddressesStmt.QueryRowContext(pCtx, pUserID).Scan(pq.Array(&addresses)); err != nil {
		return err
	}

	for _, coll := range colls {
		for _, nft := range coll.NFTs {
			if !containsAddress(addresses, nft.OwnerAddress) {
				logrus.Infof("removing nft %s from collections for %s because NFT is of old address", nft.ID, pUserID)
				_, err := c.removeNFTFromCollectionsStmt.ExecContext(pCtx, nft.ID, pUserID)
				if err != nil {
					return err
				}
			}
		}
	}

	return c.galleryRepo.RefreshCache(pCtx, pUserID)
}

// Delete deletes a collection from the database
func (c *CollectionRepository) Delete(pCtx context.Context, pID persist.DBID, pUserID persist.DBID) error {
	res, err := c.deleteCollectionStmt.ExecContext(pCtx, pID, pUserID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return persist.ErrCollectionNotFoundByID{ID: pID}
	}
	return c.galleryRepo.RefreshCache(pCtx, pUserID)
}

// GetUnassigned returns all unassigned nfts
func (c *CollectionRepository) GetUnassigned(pCtx context.Context, pUserID persist.DBID) (persist.Collection, error) {

	var addresses []persist.Address
	err := c.getUserAddressesStmt.QueryRowContext(pCtx, pUserID).Scan(pq.Array(&addresses))

	rows, err := c.getUnassignedNFTsStmt.QueryContext(pCtx, pUserID, pq.Array(addresses))
	if err != nil {
		return persist.Collection{}, err
	}
	defer rows.Close()

	nfts := []persist.CollectionNFT{}
	for rows.Next() {
		var nft persist.CollectionNFT
		err = rows.Scan(&nft.ID, &nft.CreationTime, &nft.Name, &nft.CreatorAddress, &nft.CreatorName, &nft.OwnerAddress, &nft.MultipleOwners, &nft.Contract, &nft.ImageURL, &nft.ImageThumbnailURL, &nft.ImagePreviewURL, &nft.TokenCollectionName)
		if err != nil {
			return persist.Collection{}, err
		}
		nfts = append(nfts, nft)
	}

	if err := rows.Err(); err != nil {
		return persist.Collection{}, err
	}

	return persist.Collection{
		NFTs: nfts,
	}, nil

}

// RefreshUnassigned refreshes the unassigned nfts
func (c *CollectionRepository) RefreshUnassigned(context.Context, persist.DBID) error {
	return nil
}

func ensureNFTsOwnedByUser(pCtx context.Context, c *CollectionRepository, pUserID persist.DBID, nfts []persist.DBID) error {
	var addresses []persist.Address
	err := c.getUserAddressesStmt.QueryRowContext(pCtx, pUserID).Scan(pq.Array(&addresses))
	if err != nil {
		return err
	}

	var ct int64
	err = c.checkOwnNFTsStmt.QueryRowContext(pCtx, pq.Array(addresses), pq.Array(nfts)).Scan(&ct)
	if err != nil {
		return err
	}
	if ct != int64(len(nfts)) {
		return errNotOwnedByUser
	}
	return nil
}

func containsAddress(addresses []persist.Address, address persist.Address) bool {
	for _, a := range addresses {
		if a == address {
			return true
		}
	}
	return false
}
