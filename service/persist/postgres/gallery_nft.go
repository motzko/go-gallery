package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/mikeydub/go-gallery/service/memstore"
	"github.com/mikeydub/go-gallery/service/persist"
)

// GalleryRepository is the repository for interacting with galleries in a postgres database
type GalleryRepository struct {
	db                 *sql.DB
	createStmt         *sql.Stmt
	updateStmt         *sql.Stmt
	addCollectionsStmt *sql.Stmt
	getByUserIDStmt    *sql.Stmt
	getByIDStmt        *sql.Stmt

	galleriesCache memstore.Cache
}

// NewGalleryRepository creates a new GalleryRepository
func NewGalleryRepository(db *sql.DB, gCache memstore.Cache) *GalleryRepository {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	createStmt, err := db.PrepareContext(ctx, `INSERT INTO galleries (ID, VERSION, COLLECTIONS, OWNER_USER_ID) VALUES ($1, $2, $3, $4) RETURNING ID;`)
	checkNoErr(err)

	updateStmt, err := db.PrepareContext(ctx, `UPDATE galleries SET LAST_UPDATED = $1, COLLECTIONS = $2 WHERE ID = $3 AND OWNER_USER_ID = $4;`)
	checkNoErr(err)

	addCollectionsStmt, err := db.PrepareContext(ctx, `UPDATE galleries SET COLLECTIONS = COLLECTIONS || $1 WHERE ID = $2 AND OWNER_USER_ID = $3;`)
	checkNoErr(err)

	getByUserIDStmt, err := db.PrepareContext(ctx, `SELECT g.ID,g.VERSION,g.OWNER_USER_ID,g.CREATED_AT,g.LAST_UPDATED,
	c.ID,c.OWNER_USER_ID,c.NAME,c.VERSION,c.DELETED,c.COLLECTORS_NOTE,
	c.LAYOUT,c.CREATED_AT,c.LAST_UPDATED,
	n.ID,n.OWNER_ADDRESS,
	n.MULTIPLE_OWNERS,n.NAME,n.CONTRACT,n.TOKEN_COLLECTION_NAME,n.CREATOR_ADDRESS,n.CREATOR_NAME,
	n.IMAGE_URL,n.IMAGE_THUMBNAIL_URL,n.IMAGE_PREVIEW_URL,n.ANIMATION_ORIGINAL_URL,n.CREATED_AT 
	FROM galleries g, unnest(g.COLLECTIONS) WITH ORDINALITY AS u(coll, coll_ord)
	LEFT JOIN collections c ON c.ID = coll 
	LEFT JOIN LATERAL (SELECT n.*,nft,nft_ord FROM nfts n, unnest(c.NFTS) WITH ORDINALITY AS x(nft, nft_ord)) n ON n.ID = n.nft
	WHERE g.OWNER_USER_ID = $1 AND g.DELETED = false ORDER BY coll_ord,n.nft_ord;`)
	checkNoErr(err)

	getByIDStmt, err := db.PrepareContext(ctx, `SELECT g.ID,g.VERSION,g.OWNER_USER_ID,g.CREATED_AT,g.LAST_UPDATED,
	c.ID,c.OWNER_USER_ID,c.NAME,c.VERSION,c.DELETED,c.COLLECTORS_NOTE,
	c.LAYOUT,c.CREATED_AT,c.LAST_UPDATED,n.ID,n.OWNER_ADDRESS,
	n.MULTIPLE_OWNERS,n.NAME,n.CONTRACT,n.TOKEN_COLLECTION_NAME,n.CREATOR_ADDRESS,n.CREATOR_NAME, 
	n.IMAGE_URL,n.IMAGE_THUMBNAIL_URL,n.IMAGE_PREVIEW_URL,n.ANIMATION_ORIGINAL_URL,n.CREATED_AT 
	FROM galleries g, unnest(g.COLLECTIONS) WITH ORDINALITY AS u(coll, coll_ord)
	LEFT JOIN collections c ON c.ID = coll 
	LEFT JOIN LATERAL (SELECT n.*,nft,nft_ord FROM nfts n, unnest(c.NFTS) WITH ORDINALITY AS x(nft, nft_ord)) n ON n.ID = n.nft
	WHERE g.ID = $1 AND g.DELETED = false ORDER BY coll_ord,n.nft_ord;`)
	checkNoErr(err)

	return &GalleryRepository{db: db, createStmt: createStmt, updateStmt: updateStmt, addCollectionsStmt: addCollectionsStmt, getByUserIDStmt: getByUserIDStmt, getByIDStmt: getByIDStmt, galleriesCache: gCache}
}

// Create creates a new gallery
func (g *GalleryRepository) Create(pCtx context.Context, pGallery persist.GalleryDB) (persist.DBID, error) {

	var id persist.DBID
	err := g.createStmt.QueryRowContext(pCtx, persist.GenerateID(), pGallery.Version, pq.Array(pGallery.Collections), pGallery.OwnerUserID).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

// Update updates the gallery with the given ID and ensures that gallery is owned by the given userID
func (g *GalleryRepository) Update(pCtx context.Context, pID persist.DBID, pUserID persist.DBID, pUpdate persist.GalleryUpdateInput) error {
	res, err := g.updateStmt.ExecContext(pCtx, pUpdate.LastUpdated, pq.Array(pUpdate.Collections), pID, pUserID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return persist.ErrGalleryNotFoundByID{ID: pID}
	}
	return nil
}

// AddCollections adds the given collections to the gallery with the given ID
func (g *GalleryRepository) AddCollections(pCtx context.Context, pID persist.DBID, pUserID persist.DBID, pCollections []persist.DBID) error {
	res, err := g.addCollectionsStmt.ExecContext(pCtx, pq.Array(pCollections), pID, pUserID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return persist.ErrGalleryNotFoundByID{ID: pID}
	}
	return nil
}

// GetByUserID returns the galleries owned by the given userID
func (g *GalleryRepository) GetByUserID(pCtx context.Context, pUserID persist.DBID) ([]persist.Gallery, error) {
	rows, err := g.getByUserIDStmt.QueryContext(pCtx, pUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	galleries := make(map[persist.DBID]persist.Gallery)
	collections := make(map[persist.DBID][]persist.Collection)
	var gallery persist.Gallery
	var collection persist.Collection
	var nft persist.CollectionNFT
	for rows.Next() {
		lastCollID := collection.ID

		err := rows.Scan(&gallery.ID, &gallery.Version, &gallery.OwnerUserID, &gallery.CreationTime, &gallery.LastUpdated,
			&collection.ID, &collection.OwnerUserID, &collection.Name, &collection.Version, &collection.Deleted, &collection.CollectorsNote,
			&collection.Layout, &collection.CreationTime, &collection.LastUpdated, &nft.ID, &nft.OwnerAddress,
			&nft.MultipleOwners, &nft.Name, &nft.Contract, &nft.TokenCollectionName, &nft.CreatorAddress, &nft.CreatorName,
			&nft.ImageURL, &nft.ImageThumbnailURL, &nft.ImagePreviewURL, &nft.AnimationOriginalURL, &nft.CreationTime)
		if err != nil {
			return nil, err
		}
		if _, ok := galleries[gallery.ID]; !ok {
			galleries[gallery.ID] = gallery
		}
		colls, ok := collections[gallery.ID]
		if !ok {
			colls = make([]persist.Collection, 0, 10)
			collection.NFTs = []persist.CollectionNFT{nft}
			colls = append(colls, collection)
			collections[gallery.ID] = colls
		} else {
			if lastCollID != collection.ID {
				collection.NFTs = []persist.CollectionNFT{nft}
				colls = append(colls, collection)
			} else {
				colls[len(colls)-1].NFTs = append(colls[len(colls)-1].NFTs, nft)
			}
		}
		collections[gallery.ID] = colls
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]persist.Gallery, 0, len(galleries))
	for _, gallery := range galleries {
		collections := collections[gallery.ID]
		gallery.Collections = make([]persist.Collection, 0, len(collections))
		for _, coll := range collections {
			gallery.Collections = append(gallery.Collections, coll)
		}
		result = append(result, gallery)
	}
	return result, nil

}

// GetByID returns the gallery with the given ID
func (g *GalleryRepository) GetByID(pCtx context.Context, pID persist.DBID) (persist.Gallery, error) {
	rows, err := g.getByIDStmt.QueryContext(pCtx, pID)
	if err != nil {
		return persist.Gallery{}, err
	}
	defer rows.Close()

	galleries := make(map[persist.DBID]persist.Gallery)
	collections := make(map[persist.DBID][]persist.Collection)
	var gallery persist.Gallery
	var collection persist.Collection
	var nft persist.CollectionNFT
	for rows.Next() {
		lastCollID := collection.ID

		err := rows.Scan(&gallery.ID, &gallery.Version, &gallery.OwnerUserID, &gallery.CreationTime, &gallery.LastUpdated,
			&collection.ID, &collection.OwnerUserID, &collection.Name, &collection.Version, &collection.Deleted, &collection.CollectorsNote,
			&collection.Layout, &collection.CreationTime, &collection.LastUpdated, &nft.ID, &nft.OwnerAddress,
			&nft.MultipleOwners, &nft.Name, &nft.Contract, &nft.TokenCollectionName, &nft.CreatorAddress, &nft.CreatorName,
			&nft.ImageURL, &nft.ImageThumbnailURL, &nft.ImagePreviewURL, &nft.AnimationOriginalURL, &nft.CreationTime)
		if err != nil {
			return persist.Gallery{}, err
		}
		if _, ok := galleries[gallery.ID]; !ok {
			galleries[gallery.ID] = gallery
		}
		colls, ok := collections[gallery.ID]
		if !ok {
			colls = make([]persist.Collection, 0, 10)
			collection.NFTs = []persist.CollectionNFT{nft}
			colls = append(colls, collection)
			collections[gallery.ID] = colls
		} else {
			if lastCollID != collection.ID {
				collection.NFTs = []persist.CollectionNFT{nft}
				colls = append(colls, collection)
			} else {
				colls[len(colls)-1].NFTs = append(colls[len(colls)-1].NFTs, nft)
			}
		}
		collections[gallery.ID] = colls
	}
	if err := rows.Err(); err != nil {
		return persist.Gallery{}, err
	}

	if len(galleries) > 1 {
		return persist.Gallery{}, errors.New("too many galleries")
	}

	for _, gallery := range galleries {
		collections := collections[gallery.ID]
		gallery.Collections = make([]persist.Collection, 0, len(collections))
		for _, coll := range collections {
			gallery.Collections = append(gallery.Collections, coll)
		}
		return gallery, nil
	}
	return persist.Gallery{}, persist.ErrGalleryNotFoundByID{ID: pID}
}

// RefreshCache deletes the given key in the cache
func (g *GalleryRepository) RefreshCache(pCtx context.Context, pUserID persist.DBID) error {
	return g.galleriesCache.Delete(pCtx, pUserID.String())
}