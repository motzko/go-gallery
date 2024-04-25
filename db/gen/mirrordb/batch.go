// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: batch.go

package mirrordb

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/mikeydub/go-gallery/service/persist"
)

var (
	ErrBatchAlreadyClosed = errors.New("batch already closed")
)

const processBaseOwnerEntry = `-- name: ProcessBaseOwnerEntry :batchexec
with deletion as (
    delete from base.owners where $19::bool and simplehash_kafka_key = $1
)
insert into base.owners (
    simplehash_kafka_key,
    simplehash_nft_id,
    last_updated,
    kafka_offset,
    kafka_partition,
    kafka_timestamp,
    contract_address,
    token_id,
    owner_address,
    quantity,
    collection_id,
    first_acquired_date,
    last_acquired_date,
    first_acquired_transaction,
    last_acquired_transaction,
    minted_to_this_wallet,
    airdropped_to_this_wallet,
    sold_to_this_wallet
    )
    select
        $1,
        $2,
        now(),
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16,
        $17
    where $18::bool
    on conflict (simplehash_kafka_key) do update
        set simplehash_nft_id = excluded.simplehash_nft_id,
            contract_address = excluded.contract_address,
            token_id = excluded.token_id,
            owner_address = excluded.owner_address,
            quantity = excluded.quantity,
            collection_id = excluded.collection_id,
            first_acquired_date = excluded.first_acquired_date,
            last_acquired_date = excluded.last_acquired_date,
            first_acquired_transaction = excluded.first_acquired_transaction,
            last_acquired_transaction = excluded.last_acquired_transaction,
            minted_to_this_wallet = excluded.minted_to_this_wallet,
            airdropped_to_this_wallet = excluded.airdropped_to_this_wallet,
            sold_to_this_wallet = excluded.sold_to_this_wallet
`

type ProcessBaseOwnerEntryBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type ProcessBaseOwnerEntryParams struct {
	SimplehashKafkaKey       string           `db:"simplehash_kafka_key" json:"simplehash_kafka_key"`
	SimplehashNftID          *string          `db:"simplehash_nft_id" json:"simplehash_nft_id"`
	KafkaOffset              *int64           `db:"kafka_offset" json:"kafka_offset"`
	KafkaPartition           *int32           `db:"kafka_partition" json:"kafka_partition"`
	KafkaTimestamp           *time.Time       `db:"kafka_timestamp" json:"kafka_timestamp"`
	ContractAddress          *persist.Address `db:"contract_address" json:"contract_address"`
	TokenID                  pgtype.Numeric   `db:"token_id" json:"token_id"`
	OwnerAddress             *persist.Address `db:"owner_address" json:"owner_address"`
	Quantity                 pgtype.Numeric   `db:"quantity" json:"quantity"`
	CollectionID             *string          `db:"collection_id" json:"collection_id"`
	FirstAcquiredDate        *time.Time       `db:"first_acquired_date" json:"first_acquired_date"`
	LastAcquiredDate         *time.Time       `db:"last_acquired_date" json:"last_acquired_date"`
	FirstAcquiredTransaction *string          `db:"first_acquired_transaction" json:"first_acquired_transaction"`
	LastAcquiredTransaction  *string          `db:"last_acquired_transaction" json:"last_acquired_transaction"`
	MintedToThisWallet       *bool            `db:"minted_to_this_wallet" json:"minted_to_this_wallet"`
	AirdroppedToThisWallet   *bool            `db:"airdropped_to_this_wallet" json:"airdropped_to_this_wallet"`
	SoldToThisWallet         *bool            `db:"sold_to_this_wallet" json:"sold_to_this_wallet"`
	ShouldUpsert             bool             `db:"should_upsert" json:"should_upsert"`
	ShouldDelete             bool             `db:"should_delete" json:"should_delete"`
}

func (q *Queries) ProcessBaseOwnerEntry(ctx context.Context, arg []ProcessBaseOwnerEntryParams) *ProcessBaseOwnerEntryBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.SimplehashKafkaKey,
			a.SimplehashNftID,
			a.KafkaOffset,
			a.KafkaPartition,
			a.KafkaTimestamp,
			a.ContractAddress,
			a.TokenID,
			a.OwnerAddress,
			a.Quantity,
			a.CollectionID,
			a.FirstAcquiredDate,
			a.LastAcquiredDate,
			a.FirstAcquiredTransaction,
			a.LastAcquiredTransaction,
			a.MintedToThisWallet,
			a.AirdroppedToThisWallet,
			a.SoldToThisWallet,
			a.ShouldUpsert,
			a.ShouldDelete,
		}
		batch.Queue(processBaseOwnerEntry, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &ProcessBaseOwnerEntryBatchResults{br, len(arg), false}
}

func (b *ProcessBaseOwnerEntryBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *ProcessBaseOwnerEntryBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const processEthereumOwnerEntry = `-- name: ProcessEthereumOwnerEntry :batchexec
with deletion as (
    delete from ethereum.owners where $19::bool and simplehash_kafka_key = $1
)
insert into ethereum.owners (
    simplehash_kafka_key,
    simplehash_nft_id,
    last_updated,
    kafka_offset,
    kafka_partition,
    kafka_timestamp,
    contract_address,
    token_id,
    owner_address,
    quantity,
    collection_id,
    first_acquired_date,
    last_acquired_date,
    first_acquired_transaction,
    last_acquired_transaction,
    minted_to_this_wallet,
    airdropped_to_this_wallet,
    sold_to_this_wallet
    )
    select
        $1,
        $2,
        now(),
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16,
        $17
    where $18::bool
    on conflict (simplehash_kafka_key) do update
        set simplehash_nft_id = excluded.simplehash_nft_id,
            contract_address = excluded.contract_address,
            token_id = excluded.token_id,
            owner_address = excluded.owner_address,
            quantity = excluded.quantity,
            collection_id = excluded.collection_id,
            first_acquired_date = excluded.first_acquired_date,
            last_acquired_date = excluded.last_acquired_date,
            first_acquired_transaction = excluded.first_acquired_transaction,
            last_acquired_transaction = excluded.last_acquired_transaction,
            minted_to_this_wallet = excluded.minted_to_this_wallet,
            airdropped_to_this_wallet = excluded.airdropped_to_this_wallet,
            sold_to_this_wallet = excluded.sold_to_this_wallet
`

type ProcessEthereumOwnerEntryBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type ProcessEthereumOwnerEntryParams struct {
	SimplehashKafkaKey       string           `db:"simplehash_kafka_key" json:"simplehash_kafka_key"`
	SimplehashNftID          *string          `db:"simplehash_nft_id" json:"simplehash_nft_id"`
	KafkaOffset              *int64           `db:"kafka_offset" json:"kafka_offset"`
	KafkaPartition           *int32           `db:"kafka_partition" json:"kafka_partition"`
	KafkaTimestamp           *time.Time       `db:"kafka_timestamp" json:"kafka_timestamp"`
	ContractAddress          *persist.Address `db:"contract_address" json:"contract_address"`
	TokenID                  pgtype.Numeric   `db:"token_id" json:"token_id"`
	OwnerAddress             *persist.Address `db:"owner_address" json:"owner_address"`
	Quantity                 pgtype.Numeric   `db:"quantity" json:"quantity"`
	CollectionID             *string          `db:"collection_id" json:"collection_id"`
	FirstAcquiredDate        *time.Time       `db:"first_acquired_date" json:"first_acquired_date"`
	LastAcquiredDate         *time.Time       `db:"last_acquired_date" json:"last_acquired_date"`
	FirstAcquiredTransaction *string          `db:"first_acquired_transaction" json:"first_acquired_transaction"`
	LastAcquiredTransaction  *string          `db:"last_acquired_transaction" json:"last_acquired_transaction"`
	MintedToThisWallet       *bool            `db:"minted_to_this_wallet" json:"minted_to_this_wallet"`
	AirdroppedToThisWallet   *bool            `db:"airdropped_to_this_wallet" json:"airdropped_to_this_wallet"`
	SoldToThisWallet         *bool            `db:"sold_to_this_wallet" json:"sold_to_this_wallet"`
	ShouldUpsert             bool             `db:"should_upsert" json:"should_upsert"`
	ShouldDelete             bool             `db:"should_delete" json:"should_delete"`
}

func (q *Queries) ProcessEthereumOwnerEntry(ctx context.Context, arg []ProcessEthereumOwnerEntryParams) *ProcessEthereumOwnerEntryBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.SimplehashKafkaKey,
			a.SimplehashNftID,
			a.KafkaOffset,
			a.KafkaPartition,
			a.KafkaTimestamp,
			a.ContractAddress,
			a.TokenID,
			a.OwnerAddress,
			a.Quantity,
			a.CollectionID,
			a.FirstAcquiredDate,
			a.LastAcquiredDate,
			a.FirstAcquiredTransaction,
			a.LastAcquiredTransaction,
			a.MintedToThisWallet,
			a.AirdroppedToThisWallet,
			a.SoldToThisWallet,
			a.ShouldUpsert,
			a.ShouldDelete,
		}
		batch.Queue(processEthereumOwnerEntry, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &ProcessEthereumOwnerEntryBatchResults{br, len(arg), false}
}

func (b *ProcessEthereumOwnerEntryBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *ProcessEthereumOwnerEntryBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const processEthereumTokenEntry = `-- name: ProcessEthereumTokenEntry :batchexec
with deletion as (
    delete from ethereum.tokens where $34::bool and simplehash_kafka_key = $1
),

contract_insert as (
    insert into ethereum.contracts (address, simplehash_lookup_nft_id)
    select $3::text, $2
    where $33::bool and $3 is not null
    on conflict (address) do nothing
),
    
collection_insert as (
    insert into ethereum.collections (id, simplehash_lookup_nft_id)
    select $20::text, $2
    where $33::bool and $20 is not null
    on conflict (id) do nothing
)

insert into ethereum.tokens (
    simplehash_kafka_key,
    simplehash_nft_id,
    contract_address,
    token_id,
    name,
    description,
    previews,
    image_url,
    video_url,
    audio_url,
    model_url,
    other_url,
    background_color,
    external_url,
    on_chain_created_date,
    status,
    token_count,
    owner_count,
    contract,
    collection_id,
    last_sale,
    first_created,
    rarity,
    extra_metadata,
    image_properties,
    video_properties,
    audio_properties,
    model_properties,
    other_properties,
    last_updated,
    kafka_offset,
    kafka_partition,
    kafka_timestamp
    )
    select
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16,
        $17,
        $18,
        $19,
        $20,
        $21,
        $22,
        $23,
        $24,
        $25,
        $26,
        $27,
        $28,
        $29,
        now(),
        $30,
        $31,
        $32
    where $33::bool
    on conflict (simplehash_kafka_key) do update
        set simplehash_nft_id = excluded.simplehash_nft_id,
            contract_address = excluded.contract_address,
            token_id = excluded.token_id,
            name = excluded.name,
            description = excluded.description,
            previews = excluded.previews,
            image_url = excluded.image_url,
            video_url = excluded.video_url,
            audio_url = excluded.audio_url,
            model_url = excluded.model_url,
            other_url = excluded.other_url,
            background_color = excluded.background_color,
            external_url = excluded.external_url,
            on_chain_created_date = excluded.on_chain_created_date,
            status = excluded.status,
            token_count = excluded.token_count,
            owner_count = excluded.owner_count,
            contract = excluded.contract,
            collection_id = excluded.collection_id,
            last_sale = excluded.last_sale,
            first_created = excluded.first_created,
            rarity = excluded.rarity,
            extra_metadata = excluded.extra_metadata,
            image_properties = excluded.image_properties,
            video_properties = excluded.video_properties,
            audio_properties = excluded.audio_properties,
            model_properties = excluded.model_properties,
            other_properties = excluded.other_properties,
            last_updated = now(),
            kafka_offset = excluded.kafka_offset,
            kafka_partition = excluded.kafka_partition,
            kafka_timestamp = excluded.kafka_timestamp
`

type ProcessEthereumTokenEntryBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type ProcessEthereumTokenEntryParams struct {
	SimplehashKafkaKey string           `db:"simplehash_kafka_key" json:"simplehash_kafka_key"`
	SimplehashNftID    *string          `db:"simplehash_nft_id" json:"simplehash_nft_id"`
	ContractAddress    *persist.Address `db:"contract_address" json:"contract_address"`
	TokenID            pgtype.Numeric   `db:"token_id" json:"token_id"`
	Name               *string          `db:"name" json:"name"`
	Description        *string          `db:"description" json:"description"`
	Previews           pgtype.JSONB     `db:"previews" json:"previews"`
	ImageUrl           *string          `db:"image_url" json:"image_url"`
	VideoUrl           *string          `db:"video_url" json:"video_url"`
	AudioUrl           *string          `db:"audio_url" json:"audio_url"`
	ModelUrl           *string          `db:"model_url" json:"model_url"`
	OtherUrl           *string          `db:"other_url" json:"other_url"`
	BackgroundColor    *string          `db:"background_color" json:"background_color"`
	ExternalUrl        *string          `db:"external_url" json:"external_url"`
	OnChainCreatedDate *time.Time       `db:"on_chain_created_date" json:"on_chain_created_date"`
	Status             *string          `db:"status" json:"status"`
	TokenCount         pgtype.Numeric   `db:"token_count" json:"token_count"`
	OwnerCount         pgtype.Numeric   `db:"owner_count" json:"owner_count"`
	Contract           pgtype.JSONB     `db:"contract" json:"contract"`
	CollectionID       *string          `db:"collection_id" json:"collection_id"`
	LastSale           pgtype.JSONB     `db:"last_sale" json:"last_sale"`
	FirstCreated       pgtype.JSONB     `db:"first_created" json:"first_created"`
	Rarity             pgtype.JSONB     `db:"rarity" json:"rarity"`
	ExtraMetadata      *string          `db:"extra_metadata" json:"extra_metadata"`
	ImageProperties    pgtype.JSONB     `db:"image_properties" json:"image_properties"`
	VideoProperties    pgtype.JSONB     `db:"video_properties" json:"video_properties"`
	AudioProperties    pgtype.JSONB     `db:"audio_properties" json:"audio_properties"`
	ModelProperties    pgtype.JSONB     `db:"model_properties" json:"model_properties"`
	OtherProperties    pgtype.JSONB     `db:"other_properties" json:"other_properties"`
	KafkaOffset        *int64           `db:"kafka_offset" json:"kafka_offset"`
	KafkaPartition     *int32           `db:"kafka_partition" json:"kafka_partition"`
	KafkaTimestamp     *time.Time       `db:"kafka_timestamp" json:"kafka_timestamp"`
	ShouldUpsert       bool             `db:"should_upsert" json:"should_upsert"`
	ShouldDelete       bool             `db:"should_delete" json:"should_delete"`
}

func (q *Queries) ProcessEthereumTokenEntry(ctx context.Context, arg []ProcessEthereumTokenEntryParams) *ProcessEthereumTokenEntryBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.SimplehashKafkaKey,
			a.SimplehashNftID,
			a.ContractAddress,
			a.TokenID,
			a.Name,
			a.Description,
			a.Previews,
			a.ImageUrl,
			a.VideoUrl,
			a.AudioUrl,
			a.ModelUrl,
			a.OtherUrl,
			a.BackgroundColor,
			a.ExternalUrl,
			a.OnChainCreatedDate,
			a.Status,
			a.TokenCount,
			a.OwnerCount,
			a.Contract,
			a.CollectionID,
			a.LastSale,
			a.FirstCreated,
			a.Rarity,
			a.ExtraMetadata,
			a.ImageProperties,
			a.VideoProperties,
			a.AudioProperties,
			a.ModelProperties,
			a.OtherProperties,
			a.KafkaOffset,
			a.KafkaPartition,
			a.KafkaTimestamp,
			a.ShouldUpsert,
			a.ShouldDelete,
		}
		batch.Queue(processEthereumTokenEntry, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &ProcessEthereumTokenEntryBatchResults{br, len(arg), false}
}

func (b *ProcessEthereumTokenEntryBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *ProcessEthereumTokenEntryBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const processZoraOwnerEntry = `-- name: ProcessZoraOwnerEntry :batchexec
with deletion as (
    delete from zora.owners where $19::bool and simplehash_kafka_key = $1
)
insert into zora.owners (
    simplehash_kafka_key,
    simplehash_nft_id,
    last_updated,
    kafka_offset,
    kafka_partition,
    kafka_timestamp,
    contract_address,
    token_id,
    owner_address,
    quantity,
    collection_id,
    first_acquired_date,
    last_acquired_date,
    first_acquired_transaction,
    last_acquired_transaction,
    minted_to_this_wallet,
    airdropped_to_this_wallet,
    sold_to_this_wallet
    )
    select
        $1,
        $2,
        now(),
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16,
        $17
    where $18::bool
    on conflict (simplehash_kafka_key) do update
        set simplehash_nft_id = excluded.simplehash_nft_id,
            contract_address = excluded.contract_address,
            token_id = excluded.token_id,
            owner_address = excluded.owner_address,
            quantity = excluded.quantity,
            collection_id = excluded.collection_id,
            first_acquired_date = excluded.first_acquired_date,
            last_acquired_date = excluded.last_acquired_date,
            first_acquired_transaction = excluded.first_acquired_transaction,
            last_acquired_transaction = excluded.last_acquired_transaction,
            minted_to_this_wallet = excluded.minted_to_this_wallet,
            airdropped_to_this_wallet = excluded.airdropped_to_this_wallet,
            sold_to_this_wallet = excluded.sold_to_this_wallet
`

type ProcessZoraOwnerEntryBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type ProcessZoraOwnerEntryParams struct {
	SimplehashKafkaKey       string           `db:"simplehash_kafka_key" json:"simplehash_kafka_key"`
	SimplehashNftID          *string          `db:"simplehash_nft_id" json:"simplehash_nft_id"`
	KafkaOffset              *int64           `db:"kafka_offset" json:"kafka_offset"`
	KafkaPartition           *int32           `db:"kafka_partition" json:"kafka_partition"`
	KafkaTimestamp           *time.Time       `db:"kafka_timestamp" json:"kafka_timestamp"`
	ContractAddress          *persist.Address `db:"contract_address" json:"contract_address"`
	TokenID                  pgtype.Numeric   `db:"token_id" json:"token_id"`
	OwnerAddress             *persist.Address `db:"owner_address" json:"owner_address"`
	Quantity                 pgtype.Numeric   `db:"quantity" json:"quantity"`
	CollectionID             *string          `db:"collection_id" json:"collection_id"`
	FirstAcquiredDate        *time.Time       `db:"first_acquired_date" json:"first_acquired_date"`
	LastAcquiredDate         *time.Time       `db:"last_acquired_date" json:"last_acquired_date"`
	FirstAcquiredTransaction *string          `db:"first_acquired_transaction" json:"first_acquired_transaction"`
	LastAcquiredTransaction  *string          `db:"last_acquired_transaction" json:"last_acquired_transaction"`
	MintedToThisWallet       *bool            `db:"minted_to_this_wallet" json:"minted_to_this_wallet"`
	AirdroppedToThisWallet   *bool            `db:"airdropped_to_this_wallet" json:"airdropped_to_this_wallet"`
	SoldToThisWallet         *bool            `db:"sold_to_this_wallet" json:"sold_to_this_wallet"`
	ShouldUpsert             bool             `db:"should_upsert" json:"should_upsert"`
	ShouldDelete             bool             `db:"should_delete" json:"should_delete"`
}

func (q *Queries) ProcessZoraOwnerEntry(ctx context.Context, arg []ProcessZoraOwnerEntryParams) *ProcessZoraOwnerEntryBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.SimplehashKafkaKey,
			a.SimplehashNftID,
			a.KafkaOffset,
			a.KafkaPartition,
			a.KafkaTimestamp,
			a.ContractAddress,
			a.TokenID,
			a.OwnerAddress,
			a.Quantity,
			a.CollectionID,
			a.FirstAcquiredDate,
			a.LastAcquiredDate,
			a.FirstAcquiredTransaction,
			a.LastAcquiredTransaction,
			a.MintedToThisWallet,
			a.AirdroppedToThisWallet,
			a.SoldToThisWallet,
			a.ShouldUpsert,
			a.ShouldDelete,
		}
		batch.Queue(processZoraOwnerEntry, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &ProcessZoraOwnerEntryBatchResults{br, len(arg), false}
}

func (b *ProcessZoraOwnerEntryBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *ProcessZoraOwnerEntryBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}
