package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mikeydub/go-gallery/service/persist"
)

// ContractRepository represents a contract repository in the postgres database
type ContractRepository struct {
	db                  *sql.DB
	getByAddressStmt    *sql.Stmt
	updateByAddressStmt *sql.Stmt
	ownedByAddressStmt  *sql.Stmt
	mostRecentBlockStmt *sql.Stmt
}

// NewContractRepository creates a new postgres repository for interacting with contracts
func NewContractRepository(db *sql.DB) *ContractRepository {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	getByAddressStmt, err := db.PrepareContext(ctx, `SELECT ID,VERSION,CREATED_AT,LAST_UPDATED,ADDRESS,SYMBOL,NAME,LATEST_BLOCK,OWNER_ADDRESS FROM contracts WHERE ADDRESS = $1 AND DELETED = false;`)
	checkNoErr(err)

	updateByAddressStmt, err := db.PrepareContext(ctx, `UPDATE contracts SET NAME = $2, SYMBOL = $3, OWNER_ADDRESS = $4, CREATOR_ADDRESS = $5, LATEST_BLOCK = $6, LAST_UPDATED = now() WHERE ADDRESS = $1;`)
	checkNoErr(err)

	ownedByAddressStmt, err := db.PrepareContext(ctx, `SELECT ID,VERSION,CREATED_AT,LAST_UPDATED,ADDRESS,SYMBOL,NAME,LATEST_BLOCK,OWNER_ADDRESS FROM contracts WHERE OWNER_ADDRESS = $1 AND DELETED = false;`)
	checkNoErr(err)

	mostRecentBlockStmt, err := db.PrepareContext(ctx, `SELECT MAX(LATEST_BLOCK) FROM contracts;`)
	checkNoErr(err)

	return &ContractRepository{
		db:                  db,
		getByAddressStmt:    getByAddressStmt,
		updateByAddressStmt: updateByAddressStmt,
		ownedByAddressStmt:  ownedByAddressStmt,
		mostRecentBlockStmt: mostRecentBlockStmt,
	}
}

// GetByAddress returns the contract with the given address
func (c *ContractRepository) GetByAddress(pCtx context.Context, pAddress persist.EthereumAddress) (persist.Contract, error) {
	contract := persist.Contract{}
	err := c.getByAddressStmt.QueryRowContext(pCtx, pAddress).Scan(&contract.ID, &contract.Version, &contract.CreationTime, &contract.LastUpdated, &contract.Address, &contract.Symbol, &contract.Name, &contract.LatestBlock, &contract.OwnerAddress)
	if err != nil {
		if err == sql.ErrNoRows {
			return persist.Contract{}, persist.ErrContractNotFoundByAddress{
				Address: persist.Address(pAddress),
				Chain:   persist.ChainETH,
			}
		}
		return persist.Contract{}, err
	}

	return contract, nil
}

// BulkUpsert bulk upserts the contracts by address
func (c *ContractRepository) BulkUpsert(pCtx context.Context, pContracts []persist.Contract) error {
	if len(pContracts) == 0 {
		return nil
	}
	pContracts = removeDuplicateContracts(pContracts)
	sqlStr := `INSERT INTO contracts (ID,VERSION,ADDRESS,SYMBOL,NAME,LATEST_BLOCK,OWNER_ADDRESS,CREATOR_ADDRESS) VALUES `
	vals := make([]interface{}, 0, len(pContracts)*8)
	for i, contract := range pContracts {
		sqlStr += generateValuesPlaceholders(8, i*8, nil)
		vals = append(vals, persist.GenerateID(), contract.Version, contract.Address, contract.Symbol, contract.Name, contract.LatestBlock, contract.OwnerAddress, contract.CreatorAddress)
		sqlStr += ","
	}
	sqlStr = sqlStr[:len(sqlStr)-1]
	sqlStr += ` ON CONFLICT (ADDRESS) DO UPDATE SET SYMBOL = EXCLUDED.SYMBOL,NAME = EXCLUDED.NAME,LATEST_BLOCK = EXCLUDED.LATEST_BLOCK,OWNER_ADDRESS = EXCLUDED.OWNER_ADDRESS, CREATOR_ADDRESS = EXCLUDED.CREATOR_ADDRESS;`
	_, err := c.db.ExecContext(pCtx, sqlStr, vals...)
	if err != nil {
		return fmt.Errorf("error bulk upserting contracts: %v - SQL: %s -- VALS: %+v", err, sqlStr, vals)
	}

	return nil
}

// UpdateByAddress updates the given contract's metadata fields by its address field.
func (c *ContractRepository) UpdateByAddress(ctx context.Context, addr persist.EthereumAddress, up persist.ContractUpdateInput) error {
	if _, err := c.updateByAddressStmt.ExecContext(ctx, addr, up.Name, up.Symbol, up.OwnerAddress, up.CreatorAddress, up.LatestBlock); err != nil {
		return err
	}
	return nil
}

// GetContractsOwnedByAddress returns all contracts owned by the given address
func (c *ContractRepository) GetContractsOwnedByAddress(ctx context.Context, addr persist.EthereumAddress) ([]persist.Contract, error) {
	rows, err := c.ownedByAddressStmt.QueryContext(ctx, addr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contracts := []persist.Contract{}
	for rows.Next() {
		contract := persist.Contract{}
		if err := rows.Scan(&contract.ID, &contract.Version, &contract.CreationTime, &contract.LastUpdated, &contract.Address, &contract.Symbol, &contract.Name, &contract.LatestBlock, &contract.OwnerAddress); err != nil {
			return nil, err
		}
		contracts = append(contracts, contract)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return contracts, nil
}

// MostRecentBlock returns the most recent block number of any token
func (c *ContractRepository) MostRecentBlock(pCtx context.Context) (persist.BlockNumber, error) {
	var blockNumber persist.BlockNumber
	err := c.mostRecentBlockStmt.QueryRowContext(pCtx).Scan(&blockNumber)
	if err != nil {
		return 0, err
	}
	return blockNumber, nil
}

func removeDuplicateContracts(pContracts []persist.Contract) []persist.Contract {
	if len(pContracts) == 0 {
		return pContracts
	}
	unique := map[persist.EthereumAddress]bool{}
	result := make([]persist.Contract, 0, len(pContracts))
	for _, v := range pContracts {
		if unique[v.Address] {
			continue
		}
		result = append(result, v)
		unique[v.Address] = true
	}
	return result
}
