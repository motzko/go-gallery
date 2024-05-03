// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package multichain

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mikeydub/go-gallery/db/gen/coredb"
	"github.com/mikeydub/go-gallery/service/eth"
	"github.com/mikeydub/go-gallery/service/multichain/custom"
	"github.com/mikeydub/go-gallery/service/multichain/poap"
	"github.com/mikeydub/go-gallery/service/multichain/simplehash"
	"github.com/mikeydub/go-gallery/service/multichain/tezos"
	"github.com/mikeydub/go-gallery/service/multichain/wrapper"
	"github.com/mikeydub/go-gallery/service/persist"
	"github.com/mikeydub/go-gallery/service/persist/postgres"
	"github.com/mikeydub/go-gallery/service/redis"
	"github.com/mikeydub/go-gallery/service/rpc/arweave"
	"github.com/mikeydub/go-gallery/service/rpc/ipfs"
	"github.com/mikeydub/go-gallery/service/task"
	"github.com/mikeydub/go-gallery/service/tokenmanage"
	"net/http"
)

// Injectors from inject.go:

// NewMultichainProvider is a wire injector that sets up a multichain provider instance
// ethClient.Client and task.Client are expensive to initialize so they're passed as an arg.
func NewMultichainProvider(contextContext context.Context, repositories *postgres.Repositories, queries *coredb.Queries, client *ethclient.Client, taskClient *task.Client, cache *redis.Cache) *Provider {
	httpClient := _wireClientValue
	ethereumProvider := ethInjector(contextContext, httpClient, client)
	tezosProvider := tezosInjector(httpClient)
	optimismProvider := optimismInjector(contextContext, httpClient, client)
	arbitrumProvider := arbitrumInjector(contextContext, httpClient, client)
	poapProvider := poapInjector(httpClient)
	zoraProvider := zoraInjector(contextContext, httpClient, client)
	baseProvider := baseInjector(contextContext, httpClient, client)
	polygonProvider := polygonInjector(contextContext, httpClient, client)
	chainProvider := &ChainProvider{
		Ethereum: ethereumProvider,
		Tezos:    tezosProvider,
		Optimism: optimismProvider,
		Arbitrum: arbitrumProvider,
		Poap:     poapProvider,
		Zora:     zoraProvider,
		Base:     baseProvider,
		Polygon:  polygonProvider,
	}
	tokenProcessingSubmitter := tokenProcessingSubmitterInjector(contextContext, taskClient, cache)
	provider := multichainProviderInjector(contextContext, repositories, queries, chainProvider, tokenProcessingSubmitter)
	return provider
}

var (
	_wireClientValue = http.DefaultClient
)

func multichainProviderInjector(ctx context.Context, repos *postgres.Repositories, q *coredb.Queries, chainProvider *ChainProvider, submitter *tokenmanage.TokenProcessingSubmitter) *Provider {
	providerLookup := newProviderLookup(chainProvider)
	provider := &Provider{
		Repos:     repos,
		Queries:   q,
		Chains:    providerLookup,
		Submitter: submitter,
	}
	return provider
}

func customMetadataHandlersInjector(ethCleint *ethclient.Client) *custom.CustomMetadataHandlers {
	shell := ipfs.NewShell()
	client := arweave.NewClient()
	customMetadataHandlers := custom.NewCustomMetadataHandlers(ethCleint, shell, client)
	return customMetadataHandlers
}

func ethInjector(contextContext context.Context, client *http.Client, ethclientClient *ethclient.Client) *EthereumProvider {
	chain := _wireChainValue
	provider := simplehash.NewProvider(chain, client)
	syncPipelineWrapper := ethSyncPipelineInjector(contextContext, client, chain, provider, ethclientClient)
	verifier := ethVerifierInjector(ethclientClient)
	ethereumProvider := ethProviderInjector(contextContext, syncPipelineWrapper, verifier, provider)
	return ethereumProvider
}

var (
	_wireChainValue = persist.ChainETH
)

func ethVerifierInjector(ethClient *ethclient.Client) *eth.Verifier {
	verifier := &eth.Verifier{
		Client: ethClient,
	}
	return verifier
}

func ethProviderInjector(ctx context.Context, syncPipeline *wrapper.SyncPipelineWrapper, verifier *eth.Verifier, simplehashProvider *simplehash.Provider) *EthereumProvider {
	ethereumProvider := &EthereumProvider{
		ContractFetcher:                  simplehashProvider,
		ContractsCreatorFetcher:          simplehashProvider,
		TokenDescriptorsFetcher:          simplehashProvider,
		TokenIdentifierOwnerFetcher:      syncPipeline,
		TokenMetadataBatcher:             syncPipeline,
		TokenMetadataFetcher:             syncPipeline,
		TokensByContractWalletFetcher:    syncPipeline,
		TokensByTokenIdentifiersFetcher:  syncPipeline,
		TokensIncrementalContractFetcher: syncPipeline,
		TokensIncrementalOwnerFetcher:    syncPipeline,
		Verifier:                         verifier,
	}
	return ethereumProvider
}

func ethSyncPipelineInjector(ctx context.Context, httpClient *http.Client, chain persist.Chain, simplehashProvider *simplehash.Provider, ethClient *ethclient.Client) *wrapper.SyncPipelineWrapper {
	customMetadataHandlers := customMetadataHandlersInjector(ethClient)
	syncPipelineWrapper := &wrapper.SyncPipelineWrapper{
		Chain:                            chain,
		TokenIdentifierOwnerFetcher:      simplehashProvider,
		TokensIncrementalOwnerFetcher:    simplehashProvider,
		TokensIncrementalContractFetcher: simplehashProvider,
		TokenMetadataBatcher:             simplehashProvider,
		TokensByTokenIdentifiersFetcher:  simplehashProvider,
		TokensByContractWalletFetcher:    simplehashProvider,
		CustomMetadataWrapper:            customMetadataHandlers,
	}
	return syncPipelineWrapper
}

func tezosInjector(client *http.Client) *TezosProvider {
	provider := tezos.NewProvider()
	chain := _wirePersistChainValue
	simplehashProvider := simplehash.NewProvider(chain, client)
	tezosProvider := tezosProviderInjector(provider, simplehashProvider)
	return tezosProvider
}

var (
	_wirePersistChainValue = persist.ChainTezos
)

func tezosProviderInjector(tezosProvider *tezos.Provider, simplehashProvider *simplehash.Provider) *TezosProvider {
	multichainTezosProvider := &TezosProvider{
		ContractFetcher:                  simplehashProvider,
		ContractsCreatorFetcher:          simplehashProvider,
		TokenDescriptorsFetcher:          simplehashProvider,
		TokenIdentifierOwnerFetcher:      simplehashProvider,
		TokenMetadataBatcher:             simplehashProvider,
		TokenMetadataFetcher:             simplehashProvider,
		TokensByContractWalletFetcher:    simplehashProvider,
		TokensByTokenIdentifiersFetcher:  simplehashProvider,
		TokensIncrementalContractFetcher: simplehashProvider,
		TokensIncrementalOwnerFetcher:    simplehashProvider,
		Verifier:                         tezosProvider,
	}
	return multichainTezosProvider
}

func optimismInjector(contextContext context.Context, client *http.Client, ethclientClient *ethclient.Client) *OptimismProvider {
	chain := _wireChainValue2
	provider := simplehash.NewProvider(chain, client)
	syncPipelineWrapper := optimismSyncPipelineInjector(contextContext, client, chain, provider, ethclientClient)
	optimismProvider := optimismProviderInjector(syncPipelineWrapper, provider)
	return optimismProvider
}

var (
	_wireChainValue2 = persist.ChainOptimism
)

func optimismProviderInjector(syncPipeline *wrapper.SyncPipelineWrapper, simplehashProvider *simplehash.Provider) *OptimismProvider {
	optimismProvider := &OptimismProvider{
		ContractFetcher:                  simplehashProvider,
		ContractsCreatorFetcher:          simplehashProvider,
		TokenDescriptorsFetcher:          simplehashProvider,
		TokenIdentifierOwnerFetcher:      syncPipeline,
		TokenMetadataBatcher:             syncPipeline,
		TokenMetadataFetcher:             syncPipeline,
		TokensByContractWalletFetcher:    syncPipeline,
		TokensByTokenIdentifiersFetcher:  syncPipeline,
		TokensIncrementalContractFetcher: syncPipeline,
		TokensIncrementalOwnerFetcher:    syncPipeline,
	}
	return optimismProvider
}

func optimismSyncPipelineInjector(ctx context.Context, httpClient *http.Client, chain persist.Chain, simplehashProvider *simplehash.Provider, ethClient *ethclient.Client) *wrapper.SyncPipelineWrapper {
	customMetadataHandlers := customMetadataHandlersInjector(ethClient)
	syncPipelineWrapper := &wrapper.SyncPipelineWrapper{
		Chain:                            chain,
		TokenIdentifierOwnerFetcher:      simplehashProvider,
		TokensIncrementalOwnerFetcher:    simplehashProvider,
		TokensIncrementalContractFetcher: simplehashProvider,
		TokenMetadataBatcher:             simplehashProvider,
		TokensByTokenIdentifiersFetcher:  simplehashProvider,
		TokensByContractWalletFetcher:    simplehashProvider,
		CustomMetadataWrapper:            customMetadataHandlers,
	}
	return syncPipelineWrapper
}

func arbitrumInjector(contextContext context.Context, client *http.Client, ethclientClient *ethclient.Client) *ArbitrumProvider {
	chain := _wireChainValue3
	provider := simplehash.NewProvider(chain, client)
	syncPipelineWrapper := arbitrumSyncPipelineInjector(contextContext, client, chain, provider, ethclientClient)
	arbitrumProvider := arbitrumProviderInjector(syncPipelineWrapper, provider)
	return arbitrumProvider
}

var (
	_wireChainValue3 = persist.ChainArbitrum
)

func arbitrumProviderInjector(syncPipeline *wrapper.SyncPipelineWrapper, simplehashProvider *simplehash.Provider) *ArbitrumProvider {
	arbitrumProvider := &ArbitrumProvider{
		ContractFetcher:                  simplehashProvider,
		ContractsCreatorFetcher:          simplehashProvider,
		TokenDescriptorsFetcher:          simplehashProvider,
		TokenIdentifierOwnerFetcher:      syncPipeline,
		TokenMetadataBatcher:             syncPipeline,
		TokenMetadataFetcher:             syncPipeline,
		TokensByContractWalletFetcher:    syncPipeline,
		TokensByTokenIdentifiersFetcher:  syncPipeline,
		TokensIncrementalContractFetcher: syncPipeline,
		TokensIncrementalOwnerFetcher:    syncPipeline,
	}
	return arbitrumProvider
}

func arbitrumSyncPipelineInjector(ctx context.Context, httpClient *http.Client, chain persist.Chain, simplehashProvider *simplehash.Provider, ethClient *ethclient.Client) *wrapper.SyncPipelineWrapper {
	customMetadataHandlers := customMetadataHandlersInjector(ethClient)
	syncPipelineWrapper := &wrapper.SyncPipelineWrapper{
		Chain:                            chain,
		TokenIdentifierOwnerFetcher:      simplehashProvider,
		TokensIncrementalOwnerFetcher:    simplehashProvider,
		TokensIncrementalContractFetcher: simplehashProvider,
		TokenMetadataBatcher:             simplehashProvider,
		TokensByTokenIdentifiersFetcher:  simplehashProvider,
		TokensByContractWalletFetcher:    simplehashProvider,
		CustomMetadataWrapper:            customMetadataHandlers,
	}
	return syncPipelineWrapper
}

func poapInjector(client *http.Client) *PoapProvider {
	provider := poap.NewProvider(client)
	poapProvider := poapProviderInjector(provider)
	return poapProvider
}

func poapProviderInjector(poapProvider *poap.Provider) *PoapProvider {
	multichainPoapProvider := &PoapProvider{
		TokenDescriptorsFetcher:       poapProvider,
		TokenMetadataFetcher:          poapProvider,
		TokensIncrementalOwnerFetcher: poapProvider,
		TokenIdentifierOwnerFetcher:   poapProvider,
	}
	return multichainPoapProvider
}

func zoraInjector(contextContext context.Context, client *http.Client, ethclientClient *ethclient.Client) *ZoraProvider {
	chain := _wireChainValue4
	provider := simplehash.NewProvider(chain, client)
	syncPipelineWrapper := zoraSyncPipelineInjector(contextContext, client, chain, provider, ethclientClient)
	zoraProvider := zoraProviderInjector(syncPipelineWrapper, provider)
	return zoraProvider
}

var (
	_wireChainValue4 = persist.ChainZora
)

func zoraProviderInjector(syncPipeline *wrapper.SyncPipelineWrapper, simplehashProvider *simplehash.Provider) *ZoraProvider {
	zoraProvider := &ZoraProvider{
		ContractFetcher:                  simplehashProvider,
		ContractsCreatorFetcher:          simplehashProvider,
		TokenDescriptorsFetcher:          simplehashProvider,
		TokenIdentifierOwnerFetcher:      syncPipeline,
		TokenMetadataBatcher:             syncPipeline,
		TokenMetadataFetcher:             syncPipeline,
		TokensByContractWalletFetcher:    syncPipeline,
		TokensByTokenIdentifiersFetcher:  syncPipeline,
		TokensIncrementalContractFetcher: syncPipeline,
		TokensIncrementalOwnerFetcher:    syncPipeline,
	}
	return zoraProvider
}

func zoraSyncPipelineInjector(ctx context.Context, httpClient *http.Client, chain persist.Chain, simplehashProvider *simplehash.Provider, ethClient *ethclient.Client) *wrapper.SyncPipelineWrapper {
	customMetadataHandlers := customMetadataHandlersInjector(ethClient)
	syncPipelineWrapper := &wrapper.SyncPipelineWrapper{
		Chain:                            chain,
		TokenIdentifierOwnerFetcher:      simplehashProvider,
		TokensIncrementalOwnerFetcher:    simplehashProvider,
		TokensIncrementalContractFetcher: simplehashProvider,
		TokenMetadataBatcher:             simplehashProvider,
		TokensByTokenIdentifiersFetcher:  simplehashProvider,
		TokensByContractWalletFetcher:    simplehashProvider,
		CustomMetadataWrapper:            customMetadataHandlers,
	}
	return syncPipelineWrapper
}

func baseInjector(contextContext context.Context, client *http.Client, ethclientClient *ethclient.Client) *BaseProvider {
	chain := _wireChainValue5
	provider := simplehash.NewProvider(chain, client)
	syncPipelineWrapper := baseSyncPipelineInjector(contextContext, client, chain, provider, ethclientClient)
	baseProvider := baseProvidersInjector(syncPipelineWrapper, provider, ethclientClient)
	return baseProvider
}

var (
	_wireChainValue5 = persist.ChainBase
)

func baseProvidersInjector(syncPipeline *wrapper.SyncPipelineWrapper, simplehashProvider *simplehash.Provider, ethClient *ethclient.Client) *BaseProvider {
	baseProvider := &BaseProvider{
		ContractFetcher:                  simplehashProvider,
		ContractsCreatorFetcher:          simplehashProvider,
		TokenDescriptorsFetcher:          simplehashProvider,
		TokenIdentifierOwnerFetcher:      syncPipeline,
		TokenMetadataBatcher:             syncPipeline,
		TokenMetadataFetcher:             syncPipeline,
		TokensByContractWalletFetcher:    syncPipeline,
		TokensByTokenIdentifiersFetcher:  syncPipeline,
		TokensIncrementalContractFetcher: syncPipeline,
		TokensIncrementalOwnerFetcher:    syncPipeline,
	}
	return baseProvider
}

func baseSyncPipelineInjector(ctx context.Context, httpClient *http.Client, chain persist.Chain, simplehashProvider *simplehash.Provider, ethClient *ethclient.Client) *wrapper.SyncPipelineWrapper {
	customMetadataHandlers := customMetadataHandlersInjector(ethClient)
	syncPipelineWrapper := &wrapper.SyncPipelineWrapper{
		Chain:                            chain,
		TokenIdentifierOwnerFetcher:      simplehashProvider,
		TokensIncrementalOwnerFetcher:    simplehashProvider,
		TokensIncrementalContractFetcher: simplehashProvider,
		TokenMetadataBatcher:             simplehashProvider,
		TokensByTokenIdentifiersFetcher:  simplehashProvider,
		TokensByContractWalletFetcher:    simplehashProvider,
		CustomMetadataWrapper:            customMetadataHandlers,
	}
	return syncPipelineWrapper
}

func polygonInjector(contextContext context.Context, client *http.Client, ethclientClient *ethclient.Client) *PolygonProvider {
	chain := _wireChainValue6
	provider := simplehash.NewProvider(chain, client)
	syncPipelineWrapper := polygonSyncPipelineInjector(contextContext, client, chain, provider, ethclientClient)
	polygonProvider := polygonProvidersInjector(syncPipelineWrapper, provider, ethclientClient)
	return polygonProvider
}

var (
	_wireChainValue6 = persist.ChainPolygon
)

func polygonProvidersInjector(syncPipeline *wrapper.SyncPipelineWrapper, simplehashProvider *simplehash.Provider, ethClient *ethclient.Client) *PolygonProvider {
	polygonProvider := &PolygonProvider{
		ContractFetcher:                  simplehashProvider,
		ContractsCreatorFetcher:          simplehashProvider,
		TokenDescriptorsFetcher:          simplehashProvider,
		TokenIdentifierOwnerFetcher:      syncPipeline,
		TokenMetadataBatcher:             syncPipeline,
		TokenMetadataFetcher:             simplehashProvider,
		TokensByContractWalletFetcher:    syncPipeline,
		TokensByTokenIdentifiersFetcher:  syncPipeline,
		TokensIncrementalContractFetcher: syncPipeline,
		TokensIncrementalOwnerFetcher:    syncPipeline,
	}
	return polygonProvider
}

func polygonSyncPipelineInjector(ctx context.Context, httpClient *http.Client, chain persist.Chain, simplehashProvider *simplehash.Provider, ethClient *ethclient.Client) *wrapper.SyncPipelineWrapper {
	customMetadataHandlers := customMetadataHandlersInjector(ethClient)
	syncPipelineWrapper := &wrapper.SyncPipelineWrapper{
		Chain:                            chain,
		TokenIdentifierOwnerFetcher:      simplehashProvider,
		TokensIncrementalOwnerFetcher:    simplehashProvider,
		TokensIncrementalContractFetcher: simplehashProvider,
		TokenMetadataBatcher:             simplehashProvider,
		TokensByTokenIdentifiersFetcher:  simplehashProvider,
		TokensByContractWalletFetcher:    simplehashProvider,
		CustomMetadataWrapper:            customMetadataHandlers,
	}
	return syncPipelineWrapper
}

func tokenProcessingSubmitterInjector(contextContext context.Context, client *task.Client, cache *redis.Cache) *tokenmanage.TokenProcessingSubmitter {
	registry := &tokenmanage.Registry{
		Cache: cache,
	}
	tokenProcessingSubmitter := &tokenmanage.TokenProcessingSubmitter{
		TaskClient: client,
		Registry:   registry,
	}
	return tokenProcessingSubmitter
}

// inject.go:

// New chains must be added here
func newProviderLookup(p *ChainProvider) ProviderLookup {
	return ProviderLookup{persist.ChainETH: p.Ethereum, persist.ChainTezos: p.Tezos, persist.ChainOptimism: p.Optimism, persist.ChainArbitrum: p.Arbitrum, persist.ChainPOAP: p.Poap, persist.ChainZora: p.Zora, persist.ChainBase: p.Base, persist.ChainPolygon: p.Polygon}
}