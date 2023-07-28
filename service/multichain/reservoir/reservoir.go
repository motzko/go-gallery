package reservoir

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/mikeydub/go-gallery/env"
	"github.com/mikeydub/go-gallery/service/logger"
	"github.com/mikeydub/go-gallery/service/multichain"
	"github.com/mikeydub/go-gallery/service/persist"
	"github.com/mikeydub/go-gallery/util"
)

func init() {
	env.RegisterValidation("RESERVOIR_API_KEY", "required")
}

type TokenID string

func (t TokenID) String() string {
	return string(t)
}

func (t TokenID) ToTokenID() persist.TokenID {

	if strings.HasPrefix(t.String(), "0x") {
		big, ok := new(big.Int).SetString(strings.TrimPrefix(t.String(), "0x"), 16)
		if !ok {
			return ""
		}
		return persist.TokenID(big.Text(16))
	}
	big, ok := new(big.Int).SetString(t.String(), 10)
	if !ok {
		return ""
	}
	return persist.TokenID(big.Text(16))

}

type ErrTokenNotFoundByIdentifiers struct {
	ContractAddress persist.Address
	TokenID         TokenID
	OwnerAddress    persist.Address
}

func (e ErrTokenNotFoundByIdentifiers) Error() string {
	return fmt.Sprintf("token not found for contract %s, tokenID %s, owner %s", e.ContractAddress, e.TokenID, e.OwnerAddress)
}

/*
 {
      "token": {
        "contract": "string",
        "tokenId": "string",
        "kind": "string",
        "name": "string",
        "image": "string",
        "imageSmall": "string",
        "imageLarge": "string",
        "metadata": {},
        "supply": 0,
        "remainingSupply": 0,
        "rarityScore": 0,
        "rarityRank": 0,
        "media": "string",
        "collection": {
          "id": "string",
          "name": "string",
          "imageUrl": "string",
          "openseaVerificationStatus": "string",
        }
      },
      "ownership": {
        "tokenCount": "string",
        "acquiredAt": "string"
      }
    }
  ],
  "continuation": "string"
*/

type Token struct {
	Contract   persist.Address       `json:"contract"`
	TokenID    TokenID               `json:"tokenId"`
	Kind       string                `json:"kind"`
	Name       string                `json:"name"`
	Metadata   persist.TokenMetadata `json:"metadata"`
	Media      string                `json:"media"`
	Image      string                `json:"image"`
	ImageSmall string                `json:"imageSmall"`
	ImageLarge string                `json:"imageLarge"`
	Collection Collection            `json:"collection"`
}

type Collection struct {
	ID                        string `json:"id"`
	Name                      string `json:"name"`
	ImageURL                  string `json:"imageUrl"`
	OpenseaVerificationStatus string `json:"openseaVerificationStatus"`
}

type Ownership struct {
	TokenCount string `json:"tokenCount"`
	AcquiredAt string `json:"acquiredAt"`
}

type TokenWithOwnership struct {
	Token     Token     `json:"token"`
	Ownership Ownership `json:"ownership"`
}

type TokensResponse struct {
	Tokens       []TokenWithOwnership `json:"tokens"`
	Continuation string               `json:"continuation"`
}

// Provider is an the struct for retrieving data from the Ethereum blockchain
type Provider struct {
	chain      persist.Chain
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

// NewProvider creates a new ethereum Provider
func NewProvider(chain persist.Chain, httpClient *http.Client) *Provider {
	var apiURL string
	switch chain {
	case persist.ChainETH:
		apiURL = "https://api.reservoir.tools"
	case persist.ChainOptimism:
		apiURL = "https://api-optimism.reservoir.tools"
	case persist.ChainPolygon:
		apiURL = "https://api-polygon.reservoir.tools"
	case persist.ChainArbitrum:
		apiURL = "https://api-arbitrum-nova.reservoir.tools"
	case persist.ChainZora:
		apiURL = "https://api-zora.reservoir.tools"
	case persist.ChainBase:
		apiURL = "https://api-base.reservoir.tools"
	}

	if apiURL == "" {
		panic(fmt.Sprintf("no reservoir api url set for chain %d", chain))
	}

	apiKey := env.GetString("RESERVOIR_API_KEY")

	return &Provider{
		apiURL:     apiURL,
		apiKey:     apiKey,
		chain:      chain,
		httpClient: httpClient,
	}
}

// GetBlockchainInfo retrieves blockchain info for ETH
func (d *Provider) GetBlockchainInfo(ctx context.Context) (multichain.BlockchainInfo, error) {
	chainID := 0
	switch d.chain {
	case persist.ChainOptimism:
		chainID = 10
	case persist.ChainPolygon:
		chainID = 137
	case persist.ChainArbitrum:
		chainID = 42161
	case persist.ChainZora:
		chainID = 7777777
	case persist.ChainBase:
		chainID = 84531
	}
	return multichain.BlockchainInfo{
		Chain:   d.chain,
		ChainID: chainID,
	}, nil
}

func (d *Provider) GetTokensByWalletAddress(ctx context.Context, addr persist.Address, limit, offset int) ([]multichain.ChainAgnosticToken, []multichain.ChainAgnosticContract, error) {
	tokens, err := d.getTokensByWalletAddressPaginate(ctx, addr, "")
	if err != nil {
		return nil, nil, err
	}
	t, c := tokensWithOwnershipToAgnosticTokens(tokens, addr)
	return t, c, nil
}

func (d *Provider) getTokensByWalletAddressPaginate(ctx context.Context, addr persist.Address, pageKey string) ([]TokenWithOwnership, error) {
	u := fmt.Sprintf("%s/users/%s/tokens/v7", d.apiURL, addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("x-api-key", d.apiKey)
	q := req.URL.Query()
	q.Add("limit", fmt.Sprintf("%d", 200))
	q.Add("sort_by", "acquiredAt")
	if pageKey != "" {
		q.Add("continuation", pageKey)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	var result TokensResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	tokens := result.Tokens
	if len(tokens) == 200 && result.Continuation != "" {
		nextPageTokens, err := d.getTokensByWalletAddressPaginate(ctx, addr, result.Continuation)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, nextPageTokens...)
	}

	return tokens, nil
}
func (d *Provider) GetTokenByTokenIdentifiersAndOwner(ctx context.Context, tokenIdentifiers multichain.ChainAgnosticIdentifiers, ownerAddress persist.Address) (multichain.ChainAgnosticToken, multichain.ChainAgnosticContract, error) {
	u := fmt.Sprintf("%s/users/%s/tokens/v7", d.apiURL, ownerAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return multichain.ChainAgnosticToken{}, multichain.ChainAgnosticContract{}, err
	}

	req.Header.Add("x-api-key", d.apiKey)
	q := req.URL.Query()
	q.Add("limit", fmt.Sprintf("%d", 1))
	q.Add("tokens", fmt.Sprintf("%s:%s", tokenIdentifiers.ContractAddress, tokenIdentifiers.TokenID.Base10String()))

	req.URL.RawQuery = q.Encode()

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return multichain.ChainAgnosticToken{}, multichain.ChainAgnosticContract{}, err
	}

	var result TokensResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return multichain.ChainAgnosticToken{}, multichain.ChainAgnosticContract{}, err
	}

	if len(result.Tokens) == 0 {
		return multichain.ChainAgnosticToken{}, multichain.ChainAgnosticContract{}, ErrTokenNotFoundByIdentifiers{ContractAddress: tokenIdentifiers.ContractAddress, TokenID: TokenID(tokenIdentifiers.TokenID.Base10String()), OwnerAddress: ownerAddress}
	}

	t, c := tokensWithOwnershipToAgnosticTokens(result.Tokens, ownerAddress)
	if len(t) == 0 || len(c) == 0 {
		return multichain.ChainAgnosticToken{}, multichain.ChainAgnosticContract{}, ErrTokenNotFoundByIdentifiers{ContractAddress: tokenIdentifiers.ContractAddress, TokenID: TokenID(tokenIdentifiers.TokenID.Base10String()), OwnerAddress: ownerAddress}
	}
	return t[0], c[0], nil
}

/*
{
  "animation_url": null,
  "external_app_url": "https://l2marathon.com/",
  "id": "1",
  "image_url": "https://l2marathon.com/nft/ultra-runner.jpg",
  "is_unique": true,
  "metadata": {
    "description": "Complete the Layer2 Ultra Marathon by bridging Ultra Runner ONFTs with LayerZero.",
    "external_url": "https://l2marathon.com/",
    "image": "https://l2marathon.com/nft/ultra-runner.jpg",
    "name": "Ultra Runner #1"
  },
  "owner": {
    "hash": "0x5137eEDb91A5f2cC68e86DcA15AD5C2b541654F8",
    "implementation_name": null,
    "is_contract": true,
    "is_verified": null,
    "name": null,
    "private_tags": [],
    "public_tags": [],
    "watchlist_names": []
  },
  "token": {
    "address": "0x5137eEDb91A5f2cC68e86DcA15AD5C2b541654F8",
    "circulating_market_cap": null,
    "decimals": null,
    "exchange_rate": null,
    "holders": "610",
    "icon_url": null,
    "name": "UltraMarathon",
    "symbol": "UltraRunner",
    "total_supply": null,
    "type": "ERC-721"
  }
}
*/

type BlockScoutTokenResponse struct {
	AnimationURL   string                `json:"animation_url"`
	ExternalAppURL string                `json:"external_app_url"`
	ID             string                `json:"id"`
	ImageURL       string                `json:"image_url"`
	IsUnique       bool                  `json:"is_unique"`
	Metadata       persist.TokenMetadata `json:"metadata"`
}

func (d *Provider) GetTokenMetadataByTokenIdentifiers(ctx context.Context, ti multichain.ChainAgnosticIdentifiers) (persist.TokenMetadata, error) {
	meta, err := d.reservoirMetadata(ctx, ti)
	if err == nil {
		return meta, nil
	}
	logger.For(ctx).Infof("reservoir metadata error: %s", err)
	meta, err = d.blockScoutMetadata(ctx, ti)
	if err != nil {
		logger.For(ctx).Infof("blockscout metadata error: %s", err)
		return nil, err
	}

	return meta, nil
}

func (d *Provider) reservoirMetadata(ctx context.Context, ti multichain.ChainAgnosticIdentifiers) (persist.TokenMetadata, error) {
	u := fmt.Sprintf("%s/tokens/v6", d.apiURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("x-api-key", d.apiKey)
	q := req.URL.Query()
	q.Add("tokens", fmt.Sprintf("%s:%s", ti.ContractAddress, ti.TokenID.Base10String()))

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	var res TokensResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	if len(res.Tokens) == 0 {
		return nil, ErrTokenNotFoundByIdentifiers{ContractAddress: ti.ContractAddress, TokenID: TokenID(ti.TokenID.Base10String())}
	}
	meta := res.Tokens[0].Token.Metadata
	if _, ok := util.FindFirstFieldFromMap(meta, "image", "image_url", "imageURL").(string); !ok {
		meta["image_url"] = res.Tokens[0].Token.Image
	}
	return meta, nil
}

func (d *Provider) blockScoutMetadata(ctx context.Context, ti multichain.ChainAgnosticIdentifiers) (persist.TokenMetadata, error) {
	u := fmt.Sprintf("https://base.blockscout.com/api/v2/tokens/%s/instances/%s", ti.ContractAddress, ti.TokenID.Base10String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	var res BlockScoutTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	if len(res.Metadata) == 0 {
		return nil, ErrTokenNotFoundByIdentifiers{ContractAddress: ti.ContractAddress, TokenID: TokenID(ti.TokenID.Base10String())}
	}

	if _, ok := util.FindFirstFieldFromMap(res.Metadata, "image", "image_url", "imageURL").(string); !ok {
		res.Metadata["image_url"] = res.ImageURL
	}

	if _, ok := util.FindFirstFieldFromMap(res.Metadata, "animation", "animation_url").(string); !ok {
		res.Metadata["animation_url"] = res.AnimationURL
	}

	return res.Metadata, nil
}

func tokensWithOwnershipToAgnosticTokens(tokens []TokenWithOwnership, ownerAddress persist.Address) ([]multichain.ChainAgnosticToken, []multichain.ChainAgnosticContract) {
	ts := []multichain.ChainAgnosticToken{}
	cs := []multichain.ChainAgnosticContract{}
	for _, t := range tokens {
		var tokenType persist.TokenType
		switch t.Token.Kind {
		case "erc721":
			tokenType = persist.TokenTypeERC721
		case "erc1155":
			tokenType = persist.TokenTypeERC1155
		case "erc20":
			tokenType = persist.TokenTypeERC20
		}
		var tokenQuantity persist.HexString
		b, ok := new(big.Int).SetString(t.Ownership.TokenCount, 10)
		if !ok {
			b, ok = new(big.Int).SetString(t.Ownership.TokenCount, 16)
			if !ok {
				tokenQuantity = persist.HexString("1")
			} else {
				tokenQuantity = persist.HexString(b.Text(16))
			}
		} else {
			tokenQuantity = persist.HexString(b.Text(16))
		}
		description, _ := t.Token.Metadata["description"].(string)
		ts = append(ts, multichain.ChainAgnosticToken{
			ContractAddress: t.Token.Contract,
			Descriptors: multichain.ChainAgnosticTokenDescriptors{
				Name:        t.Token.Name,
				Description: description,
			},
			TokenType:     tokenType,
			TokenID:       t.Token.TokenID.ToTokenID(),
			Quantity:      tokenQuantity,
			OwnerAddress:  ownerAddress,
			TokenMetadata: t.Token.Metadata,
			FallbackMedia: persist.FallbackMedia{
				ImageURL: persist.NullString(t.Token.Image),
			},
		})
		cs = append(cs, multichain.ChainAgnosticContract{
			Address: t.Token.Contract,
			Descriptors: multichain.ChainAgnosticContractDescriptors{
				Name:            t.Token.Collection.Name,
				ProfileImageURL: t.Token.Collection.ImageURL,
			},
		})
	}
	return ts, cs
}