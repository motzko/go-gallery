package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/mikeydub/go-gallery/service/persist"
)

// UserRepository represents a user repository in the postgres database
type UserRepository struct {
	db                  *sql.DB
	updateInfoStmt      *sql.Stmt
	existsByAddressStmt *sql.Stmt
	createStmt          *sql.Stmt
	getByIDStmt         *sql.Stmt
	getByAddressStmt    *sql.Stmt
	getByUsernameStmt   *sql.Stmt
	deleteStmt          *sql.Stmt
	addAddressStmt      *sql.Stmt
	removeAddressStmt   *sql.Stmt
}

// NewUserRepository creates a new postgres repository for interacting with users
func NewUserRepository(db *sql.DB) *UserRepository {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	updateInfoStmt, err := db.PrepareContext(ctx, `UPDATE users SET USERNAME = $2, USERNAME_IDEMPOTENT = $3, LAST_UPDATED = $4 WHERE ID = $1;`)
	checkNoErr(err)

	existsByAddressStmt, err := db.PrepareContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE ADDRESSES @> ARRAY[$1]:: varchar[] AND DELETED = false);`)
	checkNoErr(err)

	createStmt, err := db.PrepareContext(ctx, `INSERT INTO users (ID, DELETED, VERSION, USERNAME, USERNAME_IDEMPOTENT, ADDRESSES) VALUES ($1, $2, $3, $4, $5, $6) RETURNING ID;`)
	checkNoErr(err)

	getByIDStmt, err := db.PrepareContext(ctx, `SELECT ID,DELETED,VERSION,USERNAME,USERNAME_IDEMPOTENT,ADDRESSES,CREATED_AT,LAST_UPDATED FROM users WHERE ID = $1 AND DELETED = false;`)
	checkNoErr(err)

	getByAddressStmt, err := db.PrepareContext(ctx, `SELECT ID,DELETED,VERSION,USERNAME,USERNAME_IDEMPOTENT,ADDRESSES,CREATED_AT,LAST_UPDATED FROM users WHERE ADDRESSES @> ARRAY[$1]:: varchar[] AND DELETED = false;`)
	checkNoErr(err)

	getByUsernameStmt, err := db.PrepareContext(ctx, `SELECT ID,DELETED,VERSION,USERNAME,USERNAME_IDEMPOTENT,ADDRESSES,CREATED_AT,LAST_UPDATED FROM users WHERE USERNAME_IDEMPOTENT = $1 AND DELETED = false;`)
	checkNoErr(err)

	deleteStmt, err := db.PrepareContext(ctx, `UPDATE users SET DELETED = TRUE WHERE ID = $1;`)
	checkNoErr(err)

	addAddressStmt, err := db.PrepareContext(ctx, `UPDATE users SET ADDRESSES = ADDRESSES || $2 WHERE ID = $1;`)
	checkNoErr(err)

	removeAddressStmt, err := db.PrepareContext(ctx, `UPDATE users u SET ADDRESSES = array_remove(u.ADDRESSES, $2::varchar) WHERE u.ID = $1 AND $2 = ANY(u.ADDRESSES);`)
	checkNoErr(err)

	return &UserRepository{db: db, updateInfoStmt: updateInfoStmt, existsByAddressStmt: existsByAddressStmt, createStmt: createStmt, getByIDStmt: getByIDStmt, getByAddressStmt: getByAddressStmt, getByUsernameStmt: getByUsernameStmt, deleteStmt: deleteStmt, addAddressStmt: addAddressStmt, removeAddressStmt: removeAddressStmt}
}

// UpdateByID updates the user with the given ID
func (u *UserRepository) UpdateByID(pCtx context.Context, pID persist.DBID, pUpdate interface{}) error {
	switch pUpdate.(type) {
	case persist.UserUpdateInfoInput:
		update := pUpdate.(persist.UserUpdateInfoInput)
		res, err := u.updateInfoStmt.ExecContext(pCtx, pID, update.Username, strings.ToLower(update.UsernameIdempotent.String()), update.LastUpdated)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return persist.ErrUserNotFoundByID{ID: pID}
		}
	default:
		return fmt.Errorf("unsupported update type: %T", pUpdate)
	}

	return nil

}

// ExistsByAddress checks if a user exists with the given address
func (u *UserRepository) ExistsByAddress(pCtx context.Context, pAddress persist.Address) (bool, error) {

	res, err := u.existsByAddressStmt.QueryContext(pCtx, pAddress)
	if err != nil {
		return false, err
	}
	defer res.Close()
	var exists bool
	for res.Next() {
		err = res.Scan(&exists)
		if err != nil {
			return false, err
		}
	}

	if err = res.Err(); err != nil {
		return false, err
	}

	return exists, nil
}

// Create creates a new user
func (u *UserRepository) Create(pCtx context.Context, pUser persist.User) (persist.DBID, error) {

	var id persist.DBID
	err := u.createStmt.QueryRowContext(pCtx, persist.GenerateID(), pUser.Deleted, pUser.Version, pUser.Username, pUser.UsernameIdempotent, pq.Array(pUser.Addresses)).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

// GetByID gets the user with the given ID
func (u *UserRepository) GetByID(pCtx context.Context, pID persist.DBID) (persist.User, error) {

	user := persist.User{}
	err := u.getByIDStmt.QueryRowContext(pCtx, pID).Scan(&user.ID, &user.Deleted, &user.Version, &user.Username, &user.UsernameIdempotent, pq.Array(&user.Addresses), &user.CreationTime, &user.LastUpdated)
	if err != nil {
		if err == sql.ErrNoRows {
			return persist.User{}, persist.ErrUserNotFoundByID{ID: pID}
		}
		return persist.User{}, err
	}
	return user, nil
}

// GetByAddress gets the user with the given address in their list of addresses
func (u *UserRepository) GetByAddress(pCtx context.Context, pAddress persist.Address) (persist.User, error) {

	res, err := u.getByAddressStmt.QueryContext(pCtx, pAddress)
	if err != nil {
		return persist.User{}, err
	}
	defer res.Close()

	var user persist.User
	for res.Next() {
		err = res.Scan(&user.ID, &user.Deleted, &user.Version, &user.Username, &user.UsernameIdempotent, pq.Array(&user.Addresses), &user.CreationTime, &user.LastUpdated)
		if err != nil {
			return persist.User{}, err
		}
	}

	if err = res.Err(); err != nil {
		if err == sql.ErrNoRows {
			return persist.User{}, persist.ErrUserNotFoundByAddress{Address: pAddress}
		}
		return persist.User{}, err
	}

	return user, nil

}

// GetByUsername gets the user with the given username
func (u *UserRepository) GetByUsername(pCtx context.Context, pUsername string) (persist.User, error) {

	res, err := u.getByUsernameStmt.QueryContext(pCtx, strings.ToLower(pUsername))
	if err != nil {
		return persist.User{}, err
	}
	defer res.Close()

	var user persist.User
	for res.Next() {
		err = res.Scan(&user.ID, &user.Deleted, &user.Version, &user.Username, &user.UsernameIdempotent, pq.Array(&user.Addresses), &user.CreationTime, &user.LastUpdated)
		if err != nil {
			return persist.User{}, err
		}
	}

	if err = res.Err(); err != nil {
		if err == sql.ErrNoRows {
			return persist.User{}, persist.ErrUserNotFoundByUsername{Username: pUsername}
		}
		return persist.User{}, err
	}

	return user, nil

}

// Delete deletes the user with the given ID
func (u *UserRepository) Delete(pCtx context.Context, pID persist.DBID) error {

	res, err := u.deleteStmt.ExecContext(pCtx, pID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return persist.ErrUserNotFoundByID{ID: pID}
	}

	return nil
}

// AddAddresses adds the given addresses to the user with the given ID
func (u *UserRepository) AddAddresses(pCtx context.Context, pID persist.DBID, pAddresses []persist.Address) error {

	res, err := u.addAddressStmt.ExecContext(pCtx, pID, pq.Array(pAddresses))
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return persist.ErrUserNotFoundByID{ID: pID}
	}
	return nil
}

// RemoveAddresses removes the given addresses from the user with the given ID
func (u *UserRepository) RemoveAddresses(pCtx context.Context, pID persist.DBID, pAddresses []persist.Address) error {
	for _, address := range pAddresses {
		res, err := u.removeAddressStmt.ExecContext(pCtx, pID, address)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return persist.ErrUserNotFoundByAddress{Address: address}
		}
	}

	return nil
}