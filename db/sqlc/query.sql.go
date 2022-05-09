// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: query.sql

package sqlc

import (
	"context"
	"database/sql"

	"github.com/mikeydub/go-gallery/service/persist"
)

const getAddressByDetails = `-- name: GetAddressByDetails :one
SELECT id, created_at, last_updated, deleted, version, address_value, chain FROM addresses WHERE address_value = $1 AND chain = $2 AND deleted = false
`

type GetAddressByDetailsParams struct {
	AddressValue persist.AddressValue
	Chain        persist.Chain
}

func (q *Queries) GetAddressByDetails(ctx context.Context, arg GetAddressByDetailsParams) (Address, error) {
	row := q.db.QueryRow(ctx, getAddressByDetails, arg.AddressValue, arg.Chain)
	var i Address
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.LastUpdated,
		&i.Deleted,
		&i.Version,
		&i.AddressValue,
		&i.Chain,
	)
	return i, err
}

const getAddressByID = `-- name: GetAddressByID :one
SELECT id, created_at, last_updated, deleted, version, address_value, chain FROM addresses WHERE id = $1 AND deleted = false
`

func (q *Queries) GetAddressByID(ctx context.Context, id persist.DBID) (Address, error) {
	row := q.db.QueryRow(ctx, getAddressByID, id)
	var i Address
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.LastUpdated,
		&i.Deleted,
		&i.Version,
		&i.AddressValue,
		&i.Chain,
	)
	return i, err
}

const getAddressByWalletID = `-- name: GetAddressByWalletID :one
SELECT id, created_at, last_updated, deleted, version, address_value, chain FROM addresses WHERE ID = (SELECT ADDRESS FROM wallets WHERE wallets.ID = $1) AND deleted = false
`

func (q *Queries) GetAddressByWalletID(ctx context.Context, id persist.DBID) (Address, error) {
	row := q.db.QueryRow(ctx, getAddressByWalletID, id)
	var i Address
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.LastUpdated,
		&i.Deleted,
		&i.Version,
		&i.AddressValue,
		&i.Chain,
	)
	return i, err
}

const getCollectionById = `-- name: GetCollectionById :one
SELECT id, deleted, owner_user_id, nfts, version, last_updated, created_at, hidden, collectors_note, name, layout FROM collections WHERE id = $1 AND deleted = false
`

func (q *Queries) GetCollectionById(ctx context.Context, id persist.DBID) (Collection, error) {
	row := q.db.QueryRow(ctx, getCollectionById, id)
	var i Collection
	err := row.Scan(
		&i.ID,
		&i.Deleted,
		&i.OwnerUserID,
		&i.Nfts,
		&i.Version,
		&i.LastUpdated,
		&i.CreatedAt,
		&i.Hidden,
		&i.CollectorsNote,
		&i.Name,
		&i.Layout,
	)
	return i, err
}

const getCollectionsByGalleryId = `-- name: GetCollectionsByGalleryId :many
SELECT c.id, c.deleted, c.owner_user_id, c.nfts, c.version, c.last_updated, c.created_at, c.hidden, c.collectors_note, c.name, c.layout FROM galleries g, unnest(g.collections)
    WITH ORDINALITY AS x(coll_id, coll_ord)
    INNER JOIN collections c ON c.id = x.coll_id
    WHERE g.id = $1 AND g.deleted = false AND c.deleted = false ORDER BY x.coll_ord
`

func (q *Queries) GetCollectionsByGalleryId(ctx context.Context, id persist.DBID) ([]Collection, error) {
	rows, err := q.db.Query(ctx, getCollectionsByGalleryId, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Collection
	for rows.Next() {
		var i Collection
		if err := rows.Scan(
			&i.ID,
			&i.Deleted,
			&i.OwnerUserID,
			&i.Nfts,
			&i.Version,
			&i.LastUpdated,
			&i.CreatedAt,
			&i.Hidden,
			&i.CollectorsNote,
			&i.Name,
			&i.Layout,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getContractByAddress = `-- name: GetContractByAddress :one
select id, deleted, version, created_at, last_updated, name, symbol, address, latest_block, creator_address, chain FROM contracts WHERE address = $1 AND deleted = false
`

func (q *Queries) GetContractByAddress(ctx context.Context, address sql.NullString) (Contract, error) {
	row := q.db.QueryRow(ctx, getContractByAddress, address)
	var i Contract
	err := row.Scan(
		&i.ID,
		&i.Deleted,
		&i.Version,
		&i.CreatedAt,
		&i.LastUpdated,
		&i.Name,
		&i.Symbol,
		&i.Address,
		&i.LatestBlock,
		&i.CreatorAddress,
		&i.Chain,
	)
	return i, err
}

const getContractByDetails = `-- name: GetContractByDetails :one
select id, deleted, version, created_at, last_updated, name, symbol, address, latest_block, creator_address, chain FROM contracts WHERE address = (SELECT ID FROM addresses WHERE addresses.address_value = $1 AND addresses.chain = $2 AND addresses.deleted = false) AND deleted = false
`

type GetContractByDetailsParams struct {
	AddressValue persist.AddressValue
	Chain        persist.Chain
}

func (q *Queries) GetContractByDetails(ctx context.Context, arg GetContractByDetailsParams) (Contract, error) {
	row := q.db.QueryRow(ctx, getContractByDetails, arg.AddressValue, arg.Chain)
	var i Contract
	err := row.Scan(
		&i.ID,
		&i.Deleted,
		&i.Version,
		&i.CreatedAt,
		&i.LastUpdated,
		&i.Name,
		&i.Symbol,
		&i.Address,
		&i.LatestBlock,
		&i.CreatorAddress,
		&i.Chain,
	)
	return i, err
}

const getContractByID = `-- name: GetContractByID :one
select id, deleted, version, created_at, last_updated, name, symbol, address, latest_block, creator_address, chain FROM contracts WHERE id = $1 AND deleted = false
`

func (q *Queries) GetContractByID(ctx context.Context, id persist.DBID) (Contract, error) {
	row := q.db.QueryRow(ctx, getContractByID, id)
	var i Contract
	err := row.Scan(
		&i.ID,
		&i.Deleted,
		&i.Version,
		&i.CreatedAt,
		&i.LastUpdated,
		&i.Name,
		&i.Symbol,
		&i.Address,
		&i.LatestBlock,
		&i.CreatorAddress,
		&i.Chain,
	)
	return i, err
}

const getGalleriesByUserId = `-- name: GetGalleriesByUserId :many
SELECT id, deleted, last_updated, created_at, version, owner_user_id, collections FROM galleries WHERE owner_user_id = $1 AND deleted = false
`

func (q *Queries) GetGalleriesByUserId(ctx context.Context, ownerUserID persist.DBID) ([]Gallery, error) {
	rows, err := q.db.Query(ctx, getGalleriesByUserId, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Gallery
	for rows.Next() {
		var i Gallery
		if err := rows.Scan(
			&i.ID,
			&i.Deleted,
			&i.LastUpdated,
			&i.CreatedAt,
			&i.Version,
			&i.OwnerUserID,
			&i.Collections,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getGalleryByCollectionId = `-- name: GetGalleryByCollectionId :one
SELECT g.id, g.deleted, g.last_updated, g.created_at, g.version, g.owner_user_id, g.collections FROM galleries g, collections c WHERE c.id = $1 AND c.deleted = false AND $1 = ANY(g.collections) AND g.deleted = false
`

func (q *Queries) GetGalleryByCollectionId(ctx context.Context, id persist.DBID) (Gallery, error) {
	row := q.db.QueryRow(ctx, getGalleryByCollectionId, id)
	var i Gallery
	err := row.Scan(
		&i.ID,
		&i.Deleted,
		&i.LastUpdated,
		&i.CreatedAt,
		&i.Version,
		&i.OwnerUserID,
		&i.Collections,
	)
	return i, err
}

const getGalleryById = `-- name: GetGalleryById :one
SELECT id, deleted, last_updated, created_at, version, owner_user_id, collections FROM galleries WHERE id = $1 AND deleted = false
`

func (q *Queries) GetGalleryById(ctx context.Context, id persist.DBID) (Gallery, error) {
	row := q.db.QueryRow(ctx, getGalleryById, id)
	var i Gallery
	err := row.Scan(
		&i.ID,
		&i.Deleted,
		&i.LastUpdated,
		&i.CreatedAt,
		&i.Version,
		&i.OwnerUserID,
		&i.Collections,
	)
	return i, err
}

const getMembershipByMembershipId = `-- name: GetMembershipByMembershipId :one
SELECT id, deleted, version, created_at, last_updated, token_id, name, asset_url, owners FROM membership WHERE id = $1 AND deleted = false
`

func (q *Queries) GetMembershipByMembershipId(ctx context.Context, id persist.DBID) (Membership, error) {
	row := q.db.QueryRow(ctx, getMembershipByMembershipId, id)
	var i Membership
	err := row.Scan(
		&i.ID,
		&i.Deleted,
		&i.Version,
		&i.CreatedAt,
		&i.LastUpdated,
		&i.TokenID,
		&i.Name,
		&i.AssetUrl,
		&i.Owners,
	)
	return i, err
}

const getNftById = `-- name: GetNftById :one
SELECT id, deleted, version, last_updated, created_at, name, description, collectors_note, external_url, creator_address, creator_name, owner_address, multiple_owners, contract, opensea_id, opensea_token_id, token_collection_name, image_url, image_thumbnail_url, image_preview_url, image_original_url, animation_url, animation_original_url, acquisition_date, token_metadata_url FROM nfts WHERE id = $1 AND deleted = false
`

func (q *Queries) GetNftById(ctx context.Context, id persist.DBID) (Nft, error) {
	row := q.db.QueryRow(ctx, getNftById, id)
	var i Nft
	err := row.Scan(
		&i.ID,
		&i.Deleted,
		&i.Version,
		&i.LastUpdated,
		&i.CreatedAt,
		&i.Name,
		&i.Description,
		&i.CollectorsNote,
		&i.ExternalUrl,
		&i.CreatorAddress,
		&i.CreatorName,
		&i.OwnerAddress,
		&i.MultipleOwners,
		&i.Contract,
		&i.OpenseaID,
		&i.OpenseaTokenID,
		&i.TokenCollectionName,
		&i.ImageUrl,
		&i.ImageThumbnailUrl,
		&i.ImagePreviewUrl,
		&i.ImageOriginalUrl,
		&i.AnimationUrl,
		&i.AnimationOriginalUrl,
		&i.AcquisitionDate,
		&i.TokenMetadataUrl,
	)
	return i, err
}

const getNftsByCollectionId = `-- name: GetNftsByCollectionId :many
SELECT n.id, n.deleted, n.version, n.last_updated, n.created_at, n.name, n.description, n.collectors_note, n.external_url, n.creator_address, n.creator_name, n.owner_address, n.multiple_owners, n.contract, n.opensea_id, n.opensea_token_id, n.token_collection_name, n.image_url, n.image_thumbnail_url, n.image_preview_url, n.image_original_url, n.animation_url, n.animation_original_url, n.acquisition_date, n.token_metadata_url FROM collections c, unnest(c.nfts)
    WITH ORDINALITY AS x(nft_id, nft_ord)
    INNER JOIN nfts n ON n.id = x.nft_id
    WHERE c.id = $1 AND c.deleted = false AND n.deleted = false ORDER BY x.nft_ord
`

func (q *Queries) GetNftsByCollectionId(ctx context.Context, id persist.DBID) ([]Nft, error) {
	rows, err := q.db.Query(ctx, getNftsByCollectionId, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Nft
	for rows.Next() {
		var i Nft
		if err := rows.Scan(
			&i.ID,
			&i.Deleted,
			&i.Version,
			&i.LastUpdated,
			&i.CreatedAt,
			&i.Name,
			&i.Description,
			&i.CollectorsNote,
			&i.ExternalUrl,
			&i.CreatorAddress,
			&i.CreatorName,
			&i.OwnerAddress,
			&i.MultipleOwners,
			&i.Contract,
			&i.OpenseaID,
			&i.OpenseaTokenID,
			&i.TokenCollectionName,
			&i.ImageUrl,
			&i.ImageThumbnailUrl,
			&i.ImagePreviewUrl,
			&i.ImageOriginalUrl,
			&i.AnimationUrl,
			&i.AnimationOriginalUrl,
			&i.AcquisitionDate,
			&i.TokenMetadataUrl,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getNftsByOwnerAddress = `-- name: GetNftsByOwnerAddress :many
SELECT id, deleted, version, last_updated, created_at, name, description, collectors_note, external_url, creator_address, creator_name, owner_address, multiple_owners, contract, opensea_id, opensea_token_id, token_collection_name, image_url, image_thumbnail_url, image_preview_url, image_original_url, animation_url, animation_original_url, acquisition_date, token_metadata_url FROM nfts WHERE owner_address = $1 AND deleted = false
`

func (q *Queries) GetNftsByOwnerAddress(ctx context.Context, ownerAddress persist.DBID) ([]Nft, error) {
	rows, err := q.db.Query(ctx, getNftsByOwnerAddress, ownerAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Nft
	for rows.Next() {
		var i Nft
		if err := rows.Scan(
			&i.ID,
			&i.Deleted,
			&i.Version,
			&i.LastUpdated,
			&i.CreatedAt,
			&i.Name,
			&i.Description,
			&i.CollectorsNote,
			&i.ExternalUrl,
			&i.CreatorAddress,
			&i.CreatorName,
			&i.OwnerAddress,
			&i.MultipleOwners,
			&i.Contract,
			&i.OpenseaID,
			&i.OpenseaTokenID,
			&i.TokenCollectionName,
			&i.ImageUrl,
			&i.ImageThumbnailUrl,
			&i.ImagePreviewUrl,
			&i.ImageOriginalUrl,
			&i.AnimationUrl,
			&i.AnimationOriginalUrl,
			&i.AcquisitionDate,
			&i.TokenMetadataUrl,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTokenByID = `-- name: GetTokenByID :one
SELECT id, deleted, version, created_at, last_updated, name, description, contract_address, collectors_note, media, chain, token_uri, token_type, token_id, quantity, ownership_history, token_metadata, external_url, block_number, owner_user_id, owner_addresses, collection_name FROM tokens WHERE id = $1 AND deleted = false
`

func (q *Queries) GetTokenByID(ctx context.Context, id persist.DBID) (Token, error) {
	row := q.db.QueryRow(ctx, getTokenByID, id)
	var i Token
	err := row.Scan(
		&i.ID,
		&i.Deleted,
		&i.Version,
		&i.CreatedAt,
		&i.LastUpdated,
		&i.Name,
		&i.Description,
		&i.ContractAddress,
		&i.CollectorsNote,
		&i.Media,
		&i.Chain,
		&i.TokenUri,
		&i.TokenType,
		&i.TokenID,
		&i.Quantity,
		&i.OwnershipHistory,
		&i.TokenMetadata,
		&i.ExternalUrl,
		&i.BlockNumber,
		&i.OwnerUserID,
		&i.OwnerAddresses,
		&i.CollectionName,
	)
	return i, err
}

const getTokensByUserID = `-- name: GetTokensByUserID :many
SELECT id, deleted, version, created_at, last_updated, name, description, contract_address, collectors_note, media, chain, token_uri, token_type, token_id, quantity, ownership_history, token_metadata, external_url, block_number, owner_user_id, owner_addresses, collection_name FROM tokens WHERE owner_user_id = $1 AND deleted = false
`

func (q *Queries) GetTokensByUserID(ctx context.Context, ownerUserID persist.DBID) ([]Token, error) {
	rows, err := q.db.Query(ctx, getTokensByUserID, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Token
	for rows.Next() {
		var i Token
		if err := rows.Scan(
			&i.ID,
			&i.Deleted,
			&i.Version,
			&i.CreatedAt,
			&i.LastUpdated,
			&i.Name,
			&i.Description,
			&i.ContractAddress,
			&i.CollectorsNote,
			&i.Media,
			&i.Chain,
			&i.TokenUri,
			&i.TokenType,
			&i.TokenID,
			&i.Quantity,
			&i.OwnershipHistory,
			&i.TokenMetadata,
			&i.ExternalUrl,
			&i.BlockNumber,
			&i.OwnerUserID,
			&i.OwnerAddresses,
			&i.CollectionName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserByAddress = `-- name: GetUserByAddress :one
SELECT id, deleted, version, last_updated, created_at, username, username_idempotent, addresses, bio FROM users WHERE $1::varchar = ANY(addresses) AND deleted = false
`

func (q *Queries) GetUserByAddress(ctx context.Context, address string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByAddress, address)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Deleted,
		&i.Version,
		&i.LastUpdated,
		&i.CreatedAt,
		&i.Username,
		&i.UsernameIdempotent,
		&i.Addresses,
		&i.Bio,
	)
	return i, err
}

const getUserById = `-- name: GetUserById :one
SELECT id, deleted, version, last_updated, created_at, username, username_idempotent, addresses, bio FROM users WHERE id = $1 AND deleted = false
`

func (q *Queries) GetUserById(ctx context.Context, id persist.DBID) (User, error) {
	row := q.db.QueryRow(ctx, getUserById, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Deleted,
		&i.Version,
		&i.LastUpdated,
		&i.CreatedAt,
		&i.Username,
		&i.UsernameIdempotent,
		&i.Addresses,
		&i.Bio,
	)
	return i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
SELECT id, deleted, version, last_updated, created_at, username, username_idempotent, addresses, bio FROM users WHERE username_idempotent = lower($1) AND deleted = false
`

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByUsername, username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Deleted,
		&i.Version,
		&i.LastUpdated,
		&i.CreatedAt,
		&i.Username,
		&i.UsernameIdempotent,
		&i.Addresses,
		&i.Bio,
	)
	return i, err
}

const getWalletByAddress = `-- name: GetWalletByAddress :one
SELECT id, created_at, last_updated, deleted, version, address, wallet_type FROM wallets WHERE address = $1 AND deleted = false
`

func (q *Queries) GetWalletByAddress(ctx context.Context, address persist.DBID) (Wallet, error) {
	row := q.db.QueryRow(ctx, getWalletByAddress, address)
	var i Wallet
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.LastUpdated,
		&i.Deleted,
		&i.Version,
		&i.Address,
		&i.WalletType,
	)
	return i, err
}

const getWalletByAddressDetails = `-- name: GetWalletByAddressDetails :one
SELECT wallets.id, wallets.created_at, wallets.last_updated, wallets.deleted, wallets.version, wallets.address, wallets.wallet_type FROM wallets INNER JOIN addresses ON wallets.address = addresses.id WHERE addresses.address_value = $1 AND addresses.chain = $2 AND wallets.deleted = false AND addresses.deleted = false
`

type GetWalletByAddressDetailsParams struct {
	AddressValue persist.AddressValue
	Chain        persist.Chain
}

func (q *Queries) GetWalletByAddressDetails(ctx context.Context, arg GetWalletByAddressDetailsParams) (Wallet, error) {
	row := q.db.QueryRow(ctx, getWalletByAddressDetails, arg.AddressValue, arg.Chain)
	var i Wallet
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.LastUpdated,
		&i.Deleted,
		&i.Version,
		&i.Address,
		&i.WalletType,
	)
	return i, err
}

const getWalletByID = `-- name: GetWalletByID :one
SELECT id, created_at, last_updated, deleted, version, address, wallet_type FROM wallets WHERE id = $1 AND deleted = false
`

func (q *Queries) GetWalletByID(ctx context.Context, id persist.DBID) (Wallet, error) {
	row := q.db.QueryRow(ctx, getWalletByID, id)
	var i Wallet
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.LastUpdated,
		&i.Deleted,
		&i.Version,
		&i.Address,
		&i.WalletType,
	)
	return i, err
}

const getWalletsByUserID = `-- name: GetWalletsByUserID :many
SELECT w.id, w.created_at, w.last_updated, w.deleted, w.version, w.address, w.wallet_type FROM users u, unnest(u.addresses) WITH ORDINALITY AS a(addr, addr_ord) INNER JOIN wallets w on w.address = a.addr WHERE u.id = $1 AND u.deleted = false AND w.deleted = false ORDER BY a.addr_ord
`

func (q *Queries) GetWalletsByUserID(ctx context.Context, id persist.DBID) ([]Wallet, error) {
	rows, err := q.db.Query(ctx, getWalletsByUserID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Wallet
	for rows.Next() {
		var i Wallet
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.LastUpdated,
			&i.Deleted,
			&i.Version,
			&i.Address,
			&i.WalletType,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
