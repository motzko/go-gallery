package alchemy

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/mikeydub/go-gallery/env"
	"github.com/mikeydub/go-gallery/service/multichain"
	"github.com/mikeydub/go-gallery/service/persist"
	"github.com/mikeydub/go-gallery/util"
)

func init() {
	env.RegisterValidation("ALCHEMY_API_KEY", "required")
}

type TokenURI struct {
	Gateway string `json:"gateway"`
	Raw     string `json:"raw"`
}

type Media struct {
	Raw       string `json:"raw"`
	Gateway   string `json:"gateway"`
	Thumbnail string `json:"thumbnail"`
	Format    string `json:"format"`
	Bytes     int    `json:"bytes"`
}
type MetadataAttribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

type Metadata struct {
	Image           string              `json:"image"`
	ExternalURL     string              `json:"external_url"`
	BackgroundColor string              `json:"background_color"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	Attributes      []MetadataAttribute `json:"attributes"`
}

type ContractMetadata struct {
	Name                string `json:"name"`
	Symbol              string `json:"symbol"`
	TotalSupply         string `json:"totalSupply"`
	TokenType           string `json:"tokenType"`
	ContractDeployer    string `json:"contractDeployer"`
	DeployedBlockNumber string `json:"deployedBlockNumber"`
}

type Contract struct {
	Address string `json:"address"`
}

type TokenID string

func (t TokenID) String() string {
	return string(t)
}

func (t TokenID) ToTokenID() persist.TokenID {
	if strings.HasPrefix(t.String(), "0x") {

		return persist.TokenID(strings.TrimPrefix(t.String(), "0x"))
	} else {
		big, ok := new(big.Int).SetString(t.String(), 10)
		if !ok {
			return ""
		}
		return persist.TokenID(fmt.Sprintf("0x%x", big))
	}
}

type TokenMetadata struct {
	TokenType string `json:"tokenType"`
}

type TokenIdentifiers struct {
	TokenID       TokenID          `json:"tokenId"`
	TokenMetadata ContractMetadata `json:"tokenMetadata"`
}

type Token struct {
	Contract         Contract         `json:"contract"`
	ID               TokenIdentifiers `json:"id"`
	Balance          string           `json:"balance"`
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	TokenURI         TokenURI         `json:"owner"`
	Media            []Media          `json:"media"`
	Metadata         Metadata         `json:"metadata"`
	ContractMetadata ContractMetadata `json:"contractMetadata"`
	TimeLastUpdated  time.Time        `json:"timeLastUpdated"`
}

type tokensPaginated interface {
	GetTokensFromResponse(resp *http.Response) ([]Token, error)
	GetNextPageKey() string
}

type getNFTsResponse struct {
	OwnedNFTs  []Token `json:"ownedNFTs"`
	PageKey    string  `json:"pageKey"`
	TotalCount int     `json:"totalCount"`
}

func (r getNFTsResponse) GetTokensFromResponse(resp *http.Response) ([]Token, error) {
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return r.OwnedNFTs, nil
}

func (r getNFTsResponse) GetNextPageKey() string {
	return r.PageKey
}

type getNFTsForCollectionResponse struct {
	NFTs      []Token `json:"nfts"`
	NextToken TokenID `json:"nextToken"`
}

func (r getNFTsForCollectionResponse) GetTokensFromResponse(resp *http.Response) ([]Token, error) {
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return r.NFTs, nil
}

func (r getNFTsForCollectionResponse) GetNextPageKey() string {
	return r.NextToken.String()
}

// Provider is an the struct for retrieving data from the Ethereum blockchain
type Provider struct {
	alchemyAPIURL string
	httpClient    *http.Client
}

// NewProvider creates a new ethereum Provider
func NewProvider(httpClient *http.Client) *Provider {
	return &Provider{
		alchemyAPIURL: env.GetString("ALCHEMY_API_URL"),
		httpClient:    httpClient,
	}
}

// GetBlockchainInfo retrieves blockchain info for ETH
func (d *Provider) GetBlockchainInfo(ctx context.Context) (multichain.BlockchainInfo, error) {
	return multichain.BlockchainInfo{
		Chain:   persist.ChainETH,
		ChainID: 0,
	}, nil
}

func (d *Provider) RefreshToken(context.Context, multichain.ChainAgnosticIdentifiers, persist.Address) error {
	return nil
}

// GetTokensByWalletAddress retrieves tokens for a wallet address on the Ethereum Blockchain
func (d *Provider) GetTokensByWalletAddress(ctx context.Context, addr persist.Address, limit, offset int) ([]multichain.ChainAgnosticToken, []multichain.ChainAgnosticContract, error) {
	url := fmt.Sprintf("%s/getNFTs?owner=%s&withMetadata=true&orderBy=transferTime", d.alchemyAPIURL, addr)
	tokens, err := getNFTsPaginate[getNFTsResponse](ctx, url, "pageSize", "pageKey", limit, offset, "", d.httpClient)
	if err != nil {
		return nil, nil, err
	}

	cTokens, cContracts := alchemyTokensToChainAgnosticTokensForOwner(persist.EthereumAddress(addr), tokens)
	return cTokens, cContracts, nil
}

func getNFTsPaginate[T tokensPaginated](ctx context.Context, baseURL, limitFieldName, pageKeyName string, limit, offset int, pageKey string, httpClient *http.Client) ([]Token, error) {

	tokens := make([]Token, 0, limit)
	url := baseURL
	if limit > 0 && limit < 100 && limitFieldName != "" {
		url = fmt.Sprintf("%s&%s=%d", url, limitFieldName, limit)
	}
	if pageKey != "" && pageKeyName != "" {
		url = fmt.Sprintf("%s&%s=%s", url, pageKeyName, pageKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tokens from alchemy api: %s", resp.Status)
	}

	var tokensResp T
	newTokens, err := tokensResp.GetTokensFromResponse(resp)
	if err != nil {
		return nil, err
	}

	if offset > 0 && offset < 100 {
		if len(newTokens) > offset {
			newTokens = newTokens[offset:]
		} else {
			newTokens = nil
		}
	}
	tokens = append(tokens, newTokens...)

	if tokensResp.GetNextPageKey() != "" {
		if limit > 0 {
			limit -= 100
		}
		if offset > 0 {
			offset -= 100
		}
		newTokens, err := getNFTsPaginate[T](ctx, baseURL, limitFieldName, pageKeyName, limit, offset, tokensResp.GetNextPageKey(), httpClient)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, newTokens...)
	}

	return tokens, nil
}

// GetTokenMetadataByTokenIdentifiers retrieves a token's metadata for a given contract address and token ID
func (d *Provider) GetTokenMetadataByTokenIdentifiers(ctx context.Context, ti multichain.ChainAgnosticIdentifiers, ownerAddress persist.Address) (persist.TokenMetadata, error) {
	tokens, _, err := d.retryGetToken(ctx, ti, false, 0)
	if err != nil {
		return persist.TokenMetadata{}, err
	}

	if len(tokens) == 0 {
		return persist.TokenMetadata{}, fmt.Errorf("no token found for contract address %s and token ID %s", ti.ContractAddress, ti.TokenID)
	}

	token := tokens[0]
	return token.TokenMetadata, nil
}

func (d *Provider) retryGetToken(ctx context.Context, ti multichain.ChainAgnosticIdentifiers, forceRefresh bool, timeout time.Duration) ([]multichain.ChainAgnosticToken, multichain.ChainAgnosticContract, error) {
	if timeout == 0 {
		timeout = (time.Second * 20) / time.Millisecond
	}
	url := fmt.Sprintf("%s/getNFTMetadata?contract=%s&tokenId=%s&tokenUri=0x%sTimeoutInMs=%d&refreshCache=%t", d.alchemyAPIURL, ti.ContractAddress, ti.TokenID, timeout, forceRefresh)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, multichain.ChainAgnosticContract{}, err
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, multichain.ChainAgnosticContract{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if forceRefresh == false {
			return d.retryGetToken(ctx, ti, true, timeout)
		}
		return nil, multichain.ChainAgnosticContract{}, fmt.Errorf("failed to get token metadata from alchemy api: %s", resp.Status)
	}

	// will have most of the fields empty
	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, multichain.ChainAgnosticContract{}, err
	}

	tokens, contracts, err := d.alchemyTokensToChainAgnosticTokens(ctx, []Token{token})
	if err != nil {
		return nil, multichain.ChainAgnosticContract{}, err
	}

	if len(contracts) == 0 {
		return nil, multichain.ChainAgnosticContract{}, fmt.Errorf("failed to get token metadata from alchemy api: %s", resp.Status)
	}

	return tokens, contracts[0], nil
}

// GetTokensByContractAddress retrieves tokens for a contract address on the Ethereum Blockchain
func (d *Provider) GetTokensByContractAddress(ctx context.Context, contractAddress persist.Address, limit, offset int) ([]multichain.ChainAgnosticToken, multichain.ChainAgnosticContract, error) {
	url := fmt.Sprintf("%s/getNFTsForCollection?contractAddress=%s&withMetadata=true&tokenUriTimeoutInMs=20000", d.alchemyAPIURL, contractAddress)
	tokens, err := getNFTsPaginate[getNFTsResponse](ctx, url, "limit", "startToken", limit, offset, "", d.httpClient)
	if err != nil {
		return nil, multichain.ChainAgnosticContract{}, err
	}

	cTokens, cContracts, err := d.alchemyTokensToChainAgnosticTokens(ctx, tokens)
	if err != nil {
		return nil, multichain.ChainAgnosticContract{}, err
	}
	if len(cContracts) == 0 {
		return nil, multichain.ChainAgnosticContract{}, fmt.Errorf("no contract found for contract address %s", contractAddress)
	}
	return cTokens, cContracts[0], nil
}

// GetTokensByTokenIdentifiers retrieves tokens for a token identifiers on the Ethereum Blockchain
func (d *Provider) GetTokensByTokenIdentifiers(ctx context.Context, tokenIdentifiers multichain.ChainAgnosticIdentifiers, limit, offset int) ([]multichain.ChainAgnosticToken, multichain.ChainAgnosticContract, error) {
	return d.retryGetToken(ctx, tokenIdentifiers, false, 0)
}

func (d *Provider) GetTokensByTokenIdentifiersAndOwner(ctx context.Context, tokenIdentifiers multichain.ChainAgnosticIdentifiers, ownerAddress persist.Address) (multichain.ChainAgnosticToken, multichain.ChainAgnosticContract, error) {
	tokens, contract, err := d.retryGetToken(ctx, tokenIdentifiers, false, 0)
	if err != nil {
		return multichain.ChainAgnosticToken{}, multichain.ChainAgnosticContract{}, err
	}

	if len(tokens) == 0 {
		return multichain.ChainAgnosticToken{}, multichain.ChainAgnosticContract{}, fmt.Errorf("no token found for contract address %s and token ID %s", tokenIdentifiers.ContractAddress, tokenIdentifiers.TokenID)
	}

	token, ok := util.FindFirst(tokens, func(t multichain.ChainAgnosticToken) bool {
		return t.OwnerAddress == ownerAddress
	})
	if !ok {
		return multichain.ChainAgnosticToken{}, multichain.ChainAgnosticContract{}, fmt.Errorf("no token found for contract address %s and token ID %s and owner address %s", tokenIdentifiers.ContractAddress, tokenIdentifiers.TokenID, ownerAddress)
	}

	return token, contract, nil
}

type contractMetadataResponse struct {
	Address          persist.EthereumAddress `json:"address"`
	ContractMetadata ContractMetadata        `json:"contractMetadata"`
}

// GetContractByAddress retrieves an ethereum contract by address
func (d *Provider) GetContractByAddress(ctx context.Context, addr persist.Address) (multichain.ChainAgnosticContract, error) {
	url := fmt.Sprintf("%s/getContractMetadata?contract=%s", d.alchemyAPIURL, addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return multichain.ChainAgnosticContract{}, err
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return multichain.ChainAgnosticContract{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return multichain.ChainAgnosticContract{}, fmt.Errorf("failed to get contract metadata from alchemy api: %s", resp.Status)
	}

	var contractMetadataResponse contractMetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&contractMetadataResponse); err != nil {
		return multichain.ChainAgnosticContract{}, err
	}

	return multichain.ChainAgnosticContract{
		Address:        persist.Address(contractMetadataResponse.Address),
		Symbol:         contractMetadataResponse.ContractMetadata.Symbol,
		Name:           contractMetadataResponse.ContractMetadata.Name,
		CreatorAddress: persist.Address(contractMetadataResponse.ContractMetadata.ContractDeployer),
	}, nil

}

func (d *Provider) GetCommunityOwners(ctx context.Context, contractAddress persist.Address, limit, offset int) ([]multichain.ChainAgnosticCommunityOwner, error) {
	owners, err := d.paginateCollectionOwners(ctx, contractAddress, limit, offset, "")
	if err != nil {
		return nil, err
	}
	result := make([]multichain.ChainAgnosticCommunityOwner, 0, limit)

	for _, owner := range owners {
		result = append(result, multichain.ChainAgnosticCommunityOwner{
			Address: persist.Address(owner),
		})
	}

	return result, nil
}

type collectionOwnersResponse struct {
	Owners  []persist.EthereumAddress `json:"owners"`
	PageKey string                    `json:"pageKey"`
}

func (d *Provider) paginateCollectionOwners(ctx context.Context, contractAddress persist.Address, limit, offset int, pagekey string) ([]persist.EthereumAddress, error) {
	allOwners := make([]persist.EthereumAddress, 0, limit)
	url := fmt.Sprintf("%s/getCollectionOwners?contractAddress=%s", d.alchemyAPIURL, contractAddress)
	if pagekey != "" {
		url = fmt.Sprintf("%s&pageKey=%s", url, pagekey)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get collection owners from alchemy api: %s", resp.Status)
	}

	var collectionOwnersResponse collectionOwnersResponse
	if err := json.NewDecoder(resp.Body).Decode(&collectionOwnersResponse); err != nil {
		return nil, err
	}

	if offset > 0 && offset < 50000 {
		if len(collectionOwnersResponse.Owners) > offset {
			collectionOwnersResponse.Owners = collectionOwnersResponse.Owners[offset:]
		} else {
			collectionOwnersResponse.Owners = nil
		}
	}

	if limit > 0 && limit < 50000 {
		if len(collectionOwnersResponse.Owners) > limit {
			collectionOwnersResponse.Owners = collectionOwnersResponse.Owners[:limit]
		}
	}

	allOwners = append(allOwners, collectionOwnersResponse.Owners...)

	if collectionOwnersResponse.PageKey != "" {
		if limit > 0 && limit > 50000 {
			limit -= 50000
		}
		if offset > 0 && offset > 50000 {
			offset -= 50000
		}
		owners, err := d.paginateCollectionOwners(ctx, contractAddress, limit, offset, collectionOwnersResponse.PageKey)
		if err != nil {
			return nil, err
		}
		allOwners = append(collectionOwnersResponse.Owners, owners...)
	}

	return allOwners, nil
}

func (d *Provider) GetOwnedTokensByContract(ctx context.Context, contractAddress persist.Address, ownerAddress persist.Address, limit, offset int) ([]multichain.ChainAgnosticToken, multichain.ChainAgnosticContract, error) {
	url := fmt.Sprintf("%s/getNFTs?owner=%s&contractAddresses[]=%s&withMetadata=true&orderBy=transferTime", d.alchemyAPIURL, ownerAddress, contractAddress)
	tokens, err := getNFTsPaginate[getNFTsResponse](ctx, url, "pageSize", "pageKey", limit, offset, "", d.httpClient)
	if err != nil {
		return nil, multichain.ChainAgnosticContract{}, err
	}

	cTokens, cContracts := alchemyTokensToChainAgnosticTokensForOwner(persist.EthereumAddress(ownerAddress), tokens)

	if len(cContracts) == 0 {
		return nil, multichain.ChainAgnosticContract{}, fmt.Errorf("no contract found for contract address %s", contractAddress)
	}
	return cTokens, cContracts[0], nil
}

func alchemyTokensToChainAgnosticTokensForOwner(owner persist.EthereumAddress, tokens []Token) ([]multichain.ChainAgnosticToken, []multichain.ChainAgnosticContract) {
	chainAgnosticTokens := make([]multichain.ChainAgnosticToken, 0, len(tokens))
	chainAgnosticContracts := make([]multichain.ChainAgnosticContract, 0, len(tokens))
	seenContracts := make(map[persist.Address]bool)
	for _, token := range tokens {
		cToken, cContract := alchemyTokenToChainAgnosticToken(owner, token)
		if _, ok := seenContracts[cContract.Address]; !ok {
			seenContracts[cContract.Address] = true
			chainAgnosticContracts = append(chainAgnosticContracts, cContract)
		}
		chainAgnosticTokens = append(chainAgnosticTokens, cToken)
	}
	return chainAgnosticTokens, chainAgnosticContracts
}

func (d *Provider) alchemyTokensToChainAgnosticTokens(ctx context.Context, tokens []Token) ([]multichain.ChainAgnosticToken, []multichain.ChainAgnosticContract, error) {
	chainAgnosticTokens := make([]multichain.ChainAgnosticToken, 0, len(tokens))
	chainAgnosticContracts := make([]multichain.ChainAgnosticContract, 0, len(tokens))
	seenContracts := make(map[persist.Address]bool)
	for _, token := range tokens {
		owners, err := d.getOwnersForToken(ctx, token)
		if err != nil {
			return nil, nil, err
		}
		for _, owner := range owners {
			cToken, cContract := alchemyTokenToChainAgnosticToken(owner, token)
			if _, ok := seenContracts[cContract.Address]; !ok {
				seenContracts[cContract.Address] = true
				chainAgnosticContracts = append(chainAgnosticContracts, cContract)
			}
			chainAgnosticTokens = append(chainAgnosticTokens, cToken)
		}
	}
	return chainAgnosticTokens, chainAgnosticContracts, nil
}

type ownersResponse struct {
	Owners []persist.EthereumAddress `json:"owners"`
}

func (d *Provider) getOwnersForToken(ctx context.Context, token Token) ([]persist.EthereumAddress, error) {
	url := fmt.Sprintf("%s/getOwnersForToken?contractAddress=%s&tokenId=%s", d.alchemyAPIURL, token.Contract.Address, token.ID.TokenID)
	resp, err := d.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var owners ownersResponse
	if err := json.NewDecoder(resp.Body).Decode(&owners); err != nil {
		return nil, err
	}

	if len(owners.Owners) == 0 {
		return nil, fmt.Errorf("no owners found for token %s-%s", token.ID.TokenID, token.Contract.Address)
	}

	return owners.Owners, nil
}

func alchemyTokenToChainAgnosticToken(owner persist.EthereumAddress, token Token) (multichain.ChainAgnosticToken, multichain.ChainAgnosticContract) {

	var tokenType persist.TokenType
	switch token.ID.TokenMetadata.TokenType {
	case "ERC721":
		tokenType = persist.TokenTypeERC721
	case "ERC1155":
		tokenType = persist.TokenTypeERC1155
	}

	bal, ok := new(big.Int).SetString(token.Balance, 10)
	if !ok {
		bal = big.NewInt(1)
	}

	return multichain.ChainAgnosticToken{
			TokenType:       tokenType,
			Name:            token.Title,
			Description:     token.Metadata.Description,
			TokenURI:        persist.TokenURI(token.TokenURI.Raw),
			TokenMetadata:   alchemyTokenToMetadata(token),
			TokenID:         token.ID.TokenID.ToTokenID(),
			Quantity:        persist.HexString(bal.Text(16)),
			OwnerAddress:    persist.Address(owner),
			ContractAddress: persist.Address(token.Contract.Address),
			ExternalURL:     token.Metadata.ExternalURL,
		}, multichain.ChainAgnosticContract{
			Address:        persist.Address(token.Contract.Address),
			Symbol:         token.ContractMetadata.Symbol,
			Name:           token.ContractMetadata.Name,
			CreatorAddress: persist.Address(token.ContractMetadata.ContractDeployer),
		}
}

func alchemyTokenToMetadata(token Token) persist.TokenMetadata {
	firstWithFormat, ok := util.FindFirst(token.Media, func(m Media) bool {
		return m.Format != ""
	})
	metadata := persist.TokenMetadata{
		"image_url":    token.Metadata.Image,
		"name":         token.Metadata.Name,
		"description":  token.Metadata.Description,
		"external_url": token.Metadata.ExternalURL,
	}

	if ok {
		metadata["media_type"] = formatToMediaType(firstWithFormat.Format)
		metadata["format"] = firstWithFormat.Format
	}
	return metadata
}

func formatToMediaType(format string) persist.MediaType {
	switch format {
	case "jpeg", "png", "image", "jpg", "webp":
		return persist.MediaTypeImage
	case "gif":
		return persist.MediaTypeGIF
	case "video", "mp4", "quicktime":
		return persist.MediaTypeVideo
	case "audio", "mp3", "wav":
		return persist.MediaTypeAudio
	case "pdf":
		return persist.MediaTypePDF
	case "html", "iframe":
		return persist.MediaTypeHTML
	case "svg":
		return persist.MediaTypeSVG
	default:
		return persist.MediaTypeUnknown
	}
}
