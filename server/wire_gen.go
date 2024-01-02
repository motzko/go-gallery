// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package server

import (
	"context"
	"database/sql"
	"github.com/google/wire"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mikeydub/go-gallery/db/gen/coredb"
	"github.com/mikeydub/go-gallery/service/multichain"
	"github.com/mikeydub/go-gallery/service/multichain/alchemy"
	"github.com/mikeydub/go-gallery/service/multichain/eth"
	"github.com/mikeydub/go-gallery/service/multichain/infura"
	"github.com/mikeydub/go-gallery/service/multichain/opensea"
	"github.com/mikeydub/go-gallery/service/multichain/poap"
	"github.com/mikeydub/go-gallery/service/multichain/reservoir"
	"github.com/mikeydub/go-gallery/service/multichain/tezos"
	"github.com/mikeydub/go-gallery/service/multichain/zora"
	"github.com/mikeydub/go-gallery/service/persist"
	"github.com/mikeydub/go-gallery/service/persist/postgres"
	"github.com/mikeydub/go-gallery/service/redis"
	"github.com/mikeydub/go-gallery/service/rpc"
	"github.com/mikeydub/go-gallery/service/task"
	"github.com/mikeydub/go-gallery/service/tokenmanage"
	"github.com/mikeydub/go-gallery/util"
	"net/http"
)

// Injectors from inject.go:

// NewMultichainProvider is a wire injector that sets up a multichain provider instance
func NewMultichainProvider(ctx context.Context, envFunc func()) (*multichain.Provider, func()) {
	serverEnvInit := setEnv(envFunc)
	db, cleanup := newPqClient(serverEnvInit)
	pool, cleanup2 := newPgxClient(serverEnvInit)
	repositories := postgres.NewRepositories(db, pool)
	queries := newQueries(pool)
	cache := newCommunitiesCache()
	client := task.NewClient(ctx)
	httpClient := _wireClientValue
	serverTokenMetadataCache := newTokenMetadataCache()
	serverEthProviderList := ethProviderSet(serverEnvInit, client, httpClient, serverTokenMetadataCache)
	serverOptimismProviderList := optimismProviderSet(httpClient, serverTokenMetadataCache)
	serverTezosProviderList := tezosProviderSet(serverEnvInit, httpClient)
	serverPoapProviderList := poapProviderSet(serverEnvInit, httpClient)
	serverZoraProviderList := zoraProviderSet(serverEnvInit, httpClient)
	serverBaseProviderList := baseProviderSet(httpClient)
	serverPolygonProviderList := polygonProviderSet(httpClient, serverTokenMetadataCache)
	serverArbitrumProviderList := arbitrumProviderSet(httpClient, serverTokenMetadataCache)
	v := newMultichainSet(serverEthProviderList, serverOptimismProviderList, serverTezosProviderList, serverPoapProviderList, serverZoraProviderList, serverBaseProviderList, serverPolygonProviderList, serverArbitrumProviderList)
	manager := tokenmanage.New(ctx, client, cache)
	submitTokensF := newSubmitBatch(manager)
	provider := &multichain.Provider{
		Repos:        repositories,
		Queries:      queries,
		Cache:        cache,
		Chains:       v,
		SubmitTokens: submitTokensF,
	}
	return provider, func() {
		cleanup2()
		cleanup()
	}
}

var (
	_wireClientValue = &http.Client{Timeout: 0}
)

// ethProviderSet is a wire injector that creates the set of Ethereum providers
func ethProviderSet(serverEnvInit envInit, client *task.Client, httpClient *http.Client, serverTokenMetadataCache *tokenMetadataCache) ethProviderList {
	ethclientClient := rpc.NewEthClient()
	provider := eth.NewProvider(httpClient, ethclientClient, client)
	chain := _wireChainValue
	openseaProvider := opensea.NewProvider(httpClient, chain)
	syncFailureFallbackProvider := ethFallbackProvider(httpClient, serverTokenMetadataCache)
	serverEthProviderList := ethProvidersConfig(provider, openseaProvider, syncFailureFallbackProvider)
	return serverEthProviderList
}

var (
	_wireChainValue = persist.ChainETH
)

// ethProvidersConfig is a wire injector that binds multichain interfaces to their concrete Ethereum implementations
func ethProvidersConfig(indexerProvider *eth.Provider, openseaProvider *opensea.Provider, fallbackProvider multichain.SyncFailureFallbackProvider) ethProviderList {
	serverEthProviderList := ethRequirements(indexerProvider, indexerProvider, openseaProvider, fallbackProvider, fallbackProvider, fallbackProvider, fallbackProvider, indexerProvider, indexerProvider, indexerProvider, indexerProvider)
	return serverEthProviderList
}

// tezosProviderSet is a wire injector that creates the set of Tezos providers
func tezosProviderSet(serverEnvInit envInit, client *http.Client) tezosProviderList {
	syncWithContractEvalFallbackProvider := tezosFallbackProvider(serverEnvInit, client)
	serverTezosProviderList := tezosProvidersConfig(syncWithContractEvalFallbackProvider)
	return serverTezosProviderList
}

// tezosProvidersConfig is a wire injector that binds multichain interfaces to their concrete Tezos implementations
func tezosProvidersConfig(tezosProvider multichain.SyncWithContractEvalFallbackProvider) tezosProviderList {
	serverTezosProviderList := tezosRequirements(tezosProvider, tezosProvider, tezosProvider, tezosProvider, tezosProvider, tezosProvider)
	return serverTezosProviderList
}

// optimismProviderSet is a wire injector that creates the set of Optimism providers
func optimismProviderSet(client *http.Client, serverTokenMetadataCache *tokenMetadataCache) optimismProviderList {
	chain := _wirePersistChainValue
	provider := newAlchemyProvider(client, chain, serverTokenMetadataCache)
	openseaProvider := opensea.NewProvider(client, chain)
	serverOptimismProviderList := optimismProvidersConfig(provider, openseaProvider)
	return serverOptimismProviderList
}

var (
	_wirePersistChainValue = persist.ChainOptimism
)

// optimismProvidersConfig is a wire injector that binds multichain interfaces to their concrete Optimism implementations
func optimismProvidersConfig(alchemyProvider *alchemy.Provider, openseaProvider *opensea.Provider) optimismProviderList {
	serverOptimismProviderList := optimismRequirements(alchemyProvider, alchemyProvider, alchemyProvider, alchemyProvider)
	return serverOptimismProviderList
}

// arbitrumProviderSet is a wire injector that creates the set of Arbitrum providers
func arbitrumProviderSet(client *http.Client, serverTokenMetadataCache *tokenMetadataCache) arbitrumProviderList {
	chain := _wireChainValue2
	provider := newAlchemyProvider(client, chain, serverTokenMetadataCache)
	openseaProvider := opensea.NewProvider(client, chain)
	serverArbitrumProviderList := arbitrumProvidersConfig(provider, openseaProvider)
	return serverArbitrumProviderList
}

var (
	_wireChainValue2 = persist.ChainArbitrum
)

// arbitrumProvidersConfig is a wire injector that binds multichain interfaces to their concrete Arbitrum implementations
func arbitrumProvidersConfig(alchemyProvider *alchemy.Provider, openseaProvider *opensea.Provider) arbitrumProviderList {
	serverArbitrumProviderList := arbitrumRequirements(alchemyProvider, alchemyProvider, alchemyProvider, alchemyProvider, alchemyProvider)
	return serverArbitrumProviderList
}

// poapProviderSet is a wire injector that creates the set of POAP providers
func poapProviderSet(serverEnvInit envInit, client *http.Client) poapProviderList {
	provider := poap.NewProvider(client)
	serverPoapProviderList := poapProvidersConfig(provider)
	return serverPoapProviderList
}

// poapProvidersConfig is a wire injector that binds multichain interfaces to their concrete POAP implementations
func poapProvidersConfig(poapProvider *poap.Provider) poapProviderList {
	serverPoapProviderList := poapRequirements(poapProvider, poapProvider, poapProvider, poapProvider, poapProvider)
	return serverPoapProviderList
}

// zoraProviderSet is a wire injector that creates the set of zora providers
func zoraProviderSet(serverEnvInit envInit, client *http.Client) zoraProviderList {
	provider := zora.NewProvider(client)
	serverZoraProviderList := zoraProvidersConfig(provider)
	return serverZoraProviderList
}

// zoraProvidersConfig is a wire injector that binds multichain interfaces to their concrete zora implementations
func zoraProvidersConfig(zoraProvider *zora.Provider) zoraProviderList {
	serverZoraProviderList := zoraRequirements(zoraProvider, zoraProvider, zoraProvider, zoraProvider, zoraProvider, zoraProvider, zoraProvider, zoraProvider)
	return serverZoraProviderList
}

func baseProviderSet(client *http.Client) baseProviderList {
	chain := _wireChainValue3
	provider := reservoir.NewProvider(chain, client)
	serverBaseProviderList := baseProvidersConfig(provider)
	return serverBaseProviderList
}

var (
	_wireChainValue3 = persist.ChainBase
)

// baseProvidersConfig is a wire injector that binds multichain interfaces to their concrete base implementations
func baseProvidersConfig(baseProvider *reservoir.Provider) baseProviderList {
	serverBaseProviderList := baseRequirements(baseProvider, baseProvider, baseProvider)
	return serverBaseProviderList
}

// polygonProviderSet is a wire injector that creates the set of polygon providers
func polygonProviderSet(client *http.Client, serverTokenMetadataCache *tokenMetadataCache) polygonProviderList {
	chain := _wireChainValue4
	provider := newAlchemyProvider(client, chain, serverTokenMetadataCache)
	reservoirProvider := reservoir.NewProvider(chain, client)
	serverPolygonProviderList := polygonProvidersConfig(provider, reservoirProvider)
	return serverPolygonProviderList
}

var (
	_wireChainValue4 = persist.ChainPolygon
)

// polygonProvidersConfig is a wire injector that binds multichain interfaces to their concrete Polygon implementations
func polygonProvidersConfig(alchemyProvider *alchemy.Provider, reservoirProvider *reservoir.Provider) polygonProviderList {
	serverPolygonProviderList := polygonRequirements(alchemyProvider, alchemyProvider, alchemyProvider, reservoirProvider)
	return serverPolygonProviderList
}

func ethFallbackProvider(httpClient *http.Client, cache *tokenMetadataCache) multichain.SyncFailureFallbackProvider {
	chain := _wireChainValue5
	provider := newAlchemyProvider(httpClient, chain, cache)
	infuraProvider := infura.NewProvider(httpClient)
	syncFailureFallbackProvider := multichain.SyncFailureFallbackProvider{
		Primary:  provider,
		Fallback: infuraProvider,
	}
	return syncFailureFallbackProvider
}

var (
	_wireChainValue5 = persist.ChainETH
)

func tezosFallbackProvider(e envInit, httpClient *http.Client) multichain.SyncWithContractEvalFallbackProvider {
	provider := tezos.NewProvider(httpClient)
	tezosObjktProvider := tezos.NewObjktProvider()
	v := tezosTokenEvalFunc()
	syncWithContractEvalFallbackProvider := multichain.SyncWithContractEvalFallbackProvider{
		Primary:  provider,
		Fallback: tezosObjktProvider,
		Eval:     v,
	}
	return syncWithContractEvalFallbackProvider
}

// inject.go:

// envInit is a type returned after setting up the environment
// Adding envInit as a dependency to a provider will ensure that the environment is set up prior
// to calling the provider
type envInit struct{}

type ethProviderList []any

type tezosProviderList []any

type optimismProviderList []any

type poapProviderList []any

type zoraProviderList []any

type baseProviderList []any

type polygonProviderList []any

type arbitrumProviderList []any

type tokenMetadataCache redis.Cache

// dbConnSet is a wire provider set for initializing a postgres connection
var dbConnSet = wire.NewSet(
	newPqClient,
	newPgxClient,
	newQueries,
)

func setEnv(f func()) envInit {
	f()
	return envInit{}
}

func newPqClient(e envInit) (*sql.DB, func()) {
	pq := postgres.MustCreateClient()
	return pq, func() { pq.Close() }
}

func newPgxClient(envInit) (*pgxpool.Pool, func()) {
	pgx := postgres.NewPgxClient()
	return pgx, func() { pgx.Close() }
}

func newQueries(p *pgxpool.Pool) *coredb.Queries {
	return coredb.New(p)
}

// ethRequirements is the set of provider interfaces required for Ethereum
func ethRequirements(
	nr multichain.NameResolver,
	v multichain.Verifier,
	tof multichain.TokensOwnerFetcher,
	toc multichain.TokensContractFetcher,
	tiof multichain.TokensIncrementalOwnerFetcher,
	ticf multichain.TokensIncrementalContractFetcher,
	cf multichain.ContractsFetcher,
	cr multichain.ContractRefresher,
	tmf multichain.TokenMetadataFetcher,
	tcof multichain.ContractsOwnerFetcher,
	tdf multichain.TokenDescriptorsFetcher,
) ethProviderList {
	return ethProviderList{nr, v, tof, toc, tiof, ticf, cf, cr, tmf, tcof, tdf}
}

// tezosRequirements is the set of provider interfaces required for Tezos
func tezosRequirements(
	tof multichain.TokensOwnerFetcher,
	tiof multichain.TokensIncrementalOwnerFetcher,
	ticf multichain.TokensIncrementalContractFetcher,
	toc multichain.TokensContractFetcher,
	tmf multichain.TokenMetadataFetcher,
	tcof multichain.ContractsOwnerFetcher,
) tezosProviderList {
	return tezosProviderList{tof, tiof, ticf, toc, tmf, tcof}
}

// optimismRequirements is the set of provider interfaces required for Optimism
func optimismRequirements(
	tof multichain.TokensOwnerFetcher,
	tiof multichain.TokensIncrementalOwnerFetcher,
	toc multichain.TokensContractFetcher,
	tmf multichain.TokenMetadataFetcher,
) optimismProviderList {
	return optimismProviderList{tof, toc, tiof, tmf}
}

// arbitrumRequirements is the set of provider interfaces required for Arbitrum
func arbitrumRequirements(
	tof multichain.TokensOwnerFetcher,
	tiof multichain.TokensIncrementalOwnerFetcher,
	toc multichain.TokensContractFetcher,
	tmf multichain.TokenMetadataFetcher,
	tdf multichain.TokenDescriptorsFetcher,
) arbitrumProviderList {
	return arbitrumProviderList{tof, toc, tiof, tmf, tdf}
}

// poapRequirements is the set of provider interfaces required for POAP
func poapRequirements(
	nr multichain.NameResolver,
	tof multichain.TokensOwnerFetcher,
	tiof multichain.TokensIncrementalOwnerFetcher,
	toc multichain.TokensContractFetcher,
	tmf multichain.TokenMetadataFetcher,
) poapProviderList {
	return poapProviderList{nr, tof, tiof, toc, tmf}
}

// zoraRequirements is the set of provider interfaces required for zora
func zoraRequirements(
	nr multichain.ContractsFetcher,
	tof multichain.TokensOwnerFetcher,
	tiof multichain.TokensIncrementalOwnerFetcher,
	ticf multichain.TokensIncrementalContractFetcher,
	toc multichain.TokensContractFetcher,
	tcof multichain.ContractsOwnerFetcher,
	tmf multichain.TokenMetadataFetcher,
	tdf multichain.TokenDescriptorsFetcher,
) zoraProviderList {
	return zoraProviderList{nr, tof, tiof, ticf, toc, tcof, tmf, tdf}
}

// zoraRequirements is the set of provider interfaces required for zora
func baseRequirements(
	tof multichain.TokensOwnerFetcher,
	tiof multichain.TokensIncrementalOwnerFetcher,
	tdf multichain.TokenDescriptorsFetcher,
) baseProviderList {
	return baseProviderList{tof, tiof, tdf}
}

// polygonRequirements is the set of provider interfaces required for Polygon
func polygonRequirements(
	tof multichain.TokensOwnerFetcher,
	tiof multichain.TokensIncrementalOwnerFetcher,
	toc multichain.TokensContractFetcher,
	tmf multichain.TokenMetadataFetcher,
) polygonProviderList {
	return polygonProviderList{tof, tiof, toc, tmf}
}

// dedupe removes duplicate providers based on provider ID
func dedupe(providers []any) []any {
	seen := map[string]bool{}
	deduped := []any{}
	for _, p := range providers {
		if id := p.(multichain.Configurer).GetBlockchainInfo().ProviderID; !seen[id] {
			seen[id] = true
			deduped = append(deduped, p)
		}
	}
	return deduped
}

// newMultichain is a wire provider that creates a multichain provider
func newMultichainSet(
	ethProviders ethProviderList,
	optimismProviders optimismProviderList,
	tezosProviders tezosProviderList,
	poapProviders poapProviderList,
	zoraProviders zoraProviderList,
	baseProviders baseProviderList,
	polygonProviders polygonProviderList,
	arbitrumProviders arbitrumProviderList,
) map[persist.Chain][]any {
	chainToProviders := map[persist.Chain][]any{}
	chainToProviders[persist.ChainETH] = dedupe(ethProviders)
	chainToProviders[persist.ChainOptimism] = dedupe(optimismProviders)
	chainToProviders[persist.ChainTezos] = dedupe(tezosProviders)
	chainToProviders[persist.ChainPOAP] = dedupe(poapProviders)
	chainToProviders[persist.ChainZora] = dedupe(zoraProviders)
	chainToProviders[persist.ChainBase] = dedupe(baseProviders)
	chainToProviders[persist.ChainPolygon] = dedupe(polygonProviders)
	chainToProviders[persist.ChainArbitrum] = dedupe(arbitrumProviders)
	return chainToProviders
}

func tezosTokenEvalFunc() func(multichain.ChainAgnosticToken) bool {
	return func(t multichain.ChainAgnosticToken) bool {
		return tezos.IsFxHashSigned(t.ContractAddress, t.Descriptors.Name) && tezos.ContainsTezosKeywords(t)
	}
}

func newAlchemyProvider(httpClient *http.Client, chain persist.Chain, cache *tokenMetadataCache) *alchemy.Provider {
	c := redis.Cache(*cache)
	return alchemy.NewProvider(chain, httpClient, util.ToPointer(c))
}

func newCommunitiesCache() *redis.Cache {
	return redis.NewCache(redis.CommunitiesCache)
}

func newTokenMetadataCache() *tokenMetadataCache {
	cache := redis.NewCache(redis.TokenProcessingMetadataCache)
	return util.ToPointer(tokenMetadataCache(*cache))
}

func newSubmitBatch(tm *tokenmanage.Manager) multichain.SubmitTokensF {
	return tm.SubmitBatch
}
