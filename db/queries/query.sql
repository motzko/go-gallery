-- name: GetUserById :one
SELECT * FROM users WHERE id = $1 AND deleted = false;

-- name: GetUserByIdBatch :batchone
SELECT * FROM users WHERE id = $1 AND deleted = false;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username_idempotent = lower(sqlc.arg(username)) AND deleted = false;

-- name: GetUserByUsernameBatch :batchone
SELECT * FROM users WHERE username_idempotent = lower($1) AND deleted = false;

-- name: GetUserByAddress :one
SELECT * FROM users WHERE sqlc.arg(address)::varchar = ANY(addresses) AND deleted = false;

-- name: GetUserByAddressBatch :batchone
SELECT * FROM users WHERE $1::varchar = ANY(addresses) AND deleted = false;

-- name: GetGalleryById :one
SELECT * FROM galleries WHERE id = $1 AND deleted = false;

-- name: GetGalleryByIdBatch :batchone
SELECT * FROM galleries WHERE id = $1 AND deleted = false;

-- name: GetGalleryByCollectionId :one
SELECT g.* FROM galleries g, collections c WHERE c.id = $1 AND c.deleted = false AND $1 = ANY(g.collections) AND g.deleted = false;

-- name: GetGalleryByCollectionIdBatch :batchone
SELECT g.* FROM galleries g, collections c WHERE c.id = $1 AND c.deleted = false AND $1 = ANY(g.collections) AND g.deleted = false;

-- name: GetGalleriesByUserId :many
SELECT * FROM galleries WHERE owner_user_id = $1 AND deleted = false;

-- name: GetGalleriesByUserIdBatch :batchmany
SELECT * FROM galleries WHERE owner_user_id = $1 AND deleted = false;

-- name: GetCollectionById :one
SELECT * FROM collections WHERE id = $1 AND deleted = false;

-- name: GetCollectionByIdBatch :batchone
SELECT * FROM collections WHERE id = $1 AND deleted = false;

-- name: GetCollectionsByGalleryId :many
SELECT c.* FROM galleries g, unnest(g.collections)
    WITH ORDINALITY AS x(coll_id, coll_ord)
    INNER JOIN collections c ON c.id = x.coll_id
    WHERE g.id = $1 AND g.deleted = false AND c.deleted = false ORDER BY x.coll_ord;

-- name: GetCollectionsByGalleryIdBatch :batchmany
SELECT c.* FROM galleries g, unnest(g.collections)
    WITH ORDINALITY AS x(coll_id, coll_ord)
    INNER JOIN collections c ON c.id = x.coll_id
    WHERE g.id = $1 AND g.deleted = false AND c.deleted = false ORDER BY x.coll_ord;

-- name: GetNftById :one
SELECT * FROM nfts WHERE id = $1 AND deleted = false;

-- name: GetNftByIdBatch :batchone
SELECT * FROM nfts WHERE id = $1 AND deleted = false;

-- name: GetNftsByCollectionId :many
SELECT n.* FROM collections c, unnest(c.nfts)
    WITH ORDINALITY AS x(nft_id, nft_ord)
    INNER JOIN nfts n ON n.id = x.nft_id
    WHERE c.id = $1 AND c.deleted = false AND n.deleted = false ORDER BY x.nft_ord;

-- name: GetNftsByCollectionIdBatch :batchmany
SELECT n.* FROM collections c, unnest(c.nfts)
    WITH ORDINALITY AS x(nft_id, nft_ord)
    INNER JOIN nfts n ON n.id = x.nft_id
    WHERE c.id = $1 AND c.deleted = false AND n.deleted = false ORDER BY x.nft_ord;

-- name: GetNftsByOwnerAddress :many
SELECT * FROM nfts WHERE owner_address = $1 AND deleted = false;

-- name: GetNftsByOwnerAddressBatch :batchmany
SELECT * FROM nfts WHERE owner_address = $1 AND deleted = false;

-- name: GetMembershipByMembershipId :one
SELECT * FROM membership WHERE id = $1 AND deleted = false;

-- name: GetMembershipByMembershipIdBatch :batchone
SELECT * FROM membership WHERE id = $1 AND deleted = false;

-- name: GetWalletByID :one
SELECT * FROM wallets INNER JOIN addresses ON wallets.address = addresses.id WHERE wallets.id = $1 AND wallets.deleted = false AND addresses.deleted = false;

-- name: GetWalletByIDBatch :batchone
SELECT wallets.* FROM wallets INNER JOIN addresses ON wallets.address = addresses.id WHERE wallets.id = $1 AND wallets.deleted = false AND addresses.deleted = false;

-- name: GetWalletByAddress :one
SELECT wallets.* FROM wallets INNER JOIN addresses ON wallets.address = addresses.id WHERE addresses.address_value = $1 AND addresses.chain = $2 AND wallets.deleted = false AND addresses.deleted = false;

-- name: GetAddressByID :one
SELECT * FROM addresses WHERE id = $1 AND deleted = false;

-- name: GetAddressByIDBatch :batchone
SELECT * FROM addresses WHERE id = $1 AND deleted = false;

-- name: GetAddressByDetails :one
SELECT * FROM addresses WHERE address_value = $1 AND chain = $2 AND deleted = false;

-- name: GetAddressByDetailsBatch :batchone
SELECT * FROM addresses WHERE address_value = $1 AND chain = $2 AND deleted = false;