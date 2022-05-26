package persist

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"

	"github.com/lib/pq"
)

var errWalletValueNoID = fmt.Errorf("wallet value has no ID")

// Wallet represents an address on any chain
type Wallet struct {
	ID           DBID            `json:"id"`
	Version      NullInt64       `json:"version"`
	CreationTime CreationTime    `json:"created_at"`
	Deleted      NullBool        `json:"-"`
	LastUpdated  LastUpdatedTime `json:"last_updated"`

	Address    Address    `json:"address"`
	Chain      Chain      `json:"chain"`
	WalletType WalletType `json:"wallet_type"`
}

// WalletType is the type of wallet used to sign a message
type WalletType int

type WalletList []Wallet

// AddressValue represents the value of an address
type Address string

type AddressDetails struct {
	AddressValue Address `json:"address"`
	Chain        Chain   `json:"chain"`
}

const (
	// WalletTypeEOA represents an externally owned account (regular wallet address)
	WalletTypeEOA WalletType = iota
	// WalletTypeGnosis represents a smart contract gnosis safe
	WalletTypeGnosis
)

// WalletRepository represents a repository for interacting with persisted wallets
type WalletRepository interface {
	GetByAddressDetails(context.Context, Address, Chain) (Wallet, error)
	Insert(context.Context, Address, Chain, WalletType) (DBID, error)
}

func (l WalletList) Value() (driver.Value, error) {
	return pq.Array(l).Value()
}

// Scan implements the Scanner interface for the AddressList type
func (l *WalletList) Scan(value interface{}) error {
	return pq.Array(l).Scan(value)
}

// Scan implements the Scanner interface for the Wallet type
func (w *Wallet) Scan(value interface{}) error {
	if value == nil {
		*w = Wallet{}
		return nil
	}
	*w = Wallet{ID: DBID(string(value.([]uint8)))}
	return nil
}

// Value implements the database/sql driver Valuer interface for the Wallet type
func (w Wallet) Value() (driver.Value, error) {
	if w.ID == "" {
		return "", nil
	}
	return w.ID.String(), nil
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (wa *WalletType) UnmarshalGQL(v interface{}) error {
	n, ok := v.(int)
	if !ok {
		return fmt.Errorf("Chain must be an int")
	}

	*wa = WalletType(n)
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (wa WalletType) MarshalGQL(w io.Writer) {
	w.Write([]byte{uint8(wa)})
}

func (n Address) String() string {
	return string(n)
}

// Value implements the database/sql driver Valuer interface for the NullString type
func (n Address) Value() (driver.Value, error) {
	if n.String() == "" {
		return "", nil
	}
	return strings.ToValidUTF8(strings.ReplaceAll(n.String(), "\\u0000", ""), ""), nil
}

// Scan implements the database/sql Scanner interface for the NullString type
func (n *Address) Scan(value interface{}) error {
	if value == nil {
		*n = Address("")
		return nil
	}
	*n = Address(value.([]uint8))
	return nil
}

// ErrWalletNotFoundByAddressDetails is an error type for when a wallet is not found by address and chain unique combination
type ErrWalletNotFoundByAddressDetails struct {
	Address Address
	Chain   Chain
}

// ErrWalletNotFoundByAddress is an error type for when a wallet is not found by address's ID
type ErrWalletNotFoundByAddress struct {
	Address DBID
}

func (e ErrWalletNotFoundByAddressDetails) Error() string {
	return fmt.Sprintf("wallet not found by address details: %s | chain: %s", e.Address, e.Chain)
}

func (e ErrWalletNotFoundByAddress) Error() string {
	return fmt.Sprintf("wallet not found by address ID: %s", e.Address)
}
