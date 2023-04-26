package tokenprocessing

import (
	"context"
	"time"

	"cloud.google.com/go/storage"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/everFinance/goar"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/mikeydub/go-gallery/db/gen/coredb"
	"github.com/mikeydub/go-gallery/service/logger"
	"github.com/mikeydub/go-gallery/service/media"
	"github.com/mikeydub/go-gallery/service/multichain"
	"github.com/mikeydub/go-gallery/service/persist"
	"github.com/mikeydub/go-gallery/service/persist/postgres"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type tokenProcessor struct {
	queries       *coredb.Queries
	ethClient     *ethclient.Client
	mc            *multichain.Provider
	ipfsClient    *shell.Shell
	arweaveClient *goar.Client
	stg           *storage.Client
	tokenBucket   string
	tokenRepo     *postgres.TokenGalleryRepository
}

func newTokenProcessor(queries *coredb.Queries, ethClient *ethclient.Client, mc *multichain.Provider, ipfsClient *shell.Shell, arweaveClient *goar.Client, stg *storage.Client, tokenBucket string, tokenRepo *postgres.TokenGalleryRepository) *tokenProcessor {
	return &tokenProcessor{
		queries:       queries,
		ethClient:     ethClient,
		mc:            mc,
		ipfsClient:    ipfsClient,
		arweaveClient: arweaveClient,
		stg:           stg,
		tokenBucket:   tokenBucket,
		tokenRepo:     tokenRepo,
	}
}

type tokenProcessingJob struct {
	tp *tokenProcessor

	id                persist.DBID
	key               string
	token             persist.TokenGallery
	contractAddress   persist.Address
	ownerAddress      persist.Address
	imageKeywords     []string
	animationKeywords []string
	cause             persist.ProcessingCause
	pipelineMetadata  *persist.PipelineMetadata
}

func (tp *tokenProcessor) processTokenPipeline(c context.Context, key string, t persist.TokenGallery, contractAddress, ownerAddress persist.Address, imageKeywords, animationKeywords []string, cause persist.ProcessingCause) error {

	loggerCtx := logger.NewContextWithFields(c, logrus.Fields{
		"tokenDBID":       t.ID,
		"tokenID":         t.TokenID,
		"contractDBID":    t.Contract,
		"contractAddress": contractAddress,
		"chain":           t.Chain,
	})

	ctx, cancel := context.WithTimeout(loggerCtx, time.Minute*10)
	defer cancel()
	job := &tokenProcessingJob{
		id: persist.GenerateID(),

		tp:                tp,
		key:               key,
		token:             t,
		contractAddress:   contractAddress,
		ownerAddress:      ownerAddress,
		imageKeywords:     imageKeywords,
		animationKeywords: animationKeywords,
		cause:             cause,
		pipelineMetadata:  new(persist.PipelineMetadata),
	}

	totalTime := time.Now()
	defer func() {
		logger.For(ctx).Infof("total time for token processing job: %s", time.Since(totalTime))
	}()

	return job.run(ctx)
}

func (tpj *tokenProcessingJob) run(ctx context.Context) error {
	toInsert, err := tpj.createMediaForToken(ctx)
	if err != nil {
		logger.For(ctx).Errorf("error creating media for token: %s", err)
	}

	return tpj.persistMedia(ctx, toInsert)
}

func (tpj *tokenProcessingJob) createMediaForToken(ctx context.Context) (coredb.TokenMedium, error) {
	defer func() {
		if r := recover(); r != nil {
			logger.For(ctx).Errorf("panic in createMediaForToken: %s", r)
		}
	}()

	result := coredb.TokenMedium{
		ID:              persist.GenerateID(),
		Contract:        tpj.token.Contract,
		TokenID:         tpj.token.TokenID,
		Chain:           tpj.token.Chain,
		Active:          true,
		ProcessingJobID: tpj.id,
	}

	result.Metadata = tpj.retrieveMetadata(ctx)

	result.Name, result.Description = tpj.retrieveTokenInfo(ctx, result.Metadata)

	cachedObjects, err := tpj.cacheMediaObjects(ctx, result.Metadata)
	if err != nil {
		return result, err
	}

	result.Media = tpj.createMediaFromCachedObjects(ctx, cachedObjects)

	return result, nil
}

func (tpj *tokenProcessingJob) retrieveMetadata(ctx context.Context) persist.TokenMetadata {
	defer trackStepStatus(&tpj.pipelineMetadata.MetadataRetrieval)

	newMetadata := tpj.token.TokenMetadata

	if len(newMetadata) == 0 || tpj.cause == persist.ProcessingCauseRefresh {
		mcMetadata, err := tpj.tp.mc.GetTokenMetadataByTokenIdentifiers(ctx, tpj.contractAddress, tpj.token.TokenID, tpj.ownerAddress, tpj.token.Chain)
		if err != nil {
			logger.For(ctx).Errorf("error getting metadata from chain: %s", err)
			failStep(&tpj.pipelineMetadata.MetadataRetrieval)
		} else if mcMetadata != nil && len(mcMetadata) > 0 {
			logger.For(ctx).Infof("got metadata from chain: %v", mcMetadata)
			newMetadata = mcMetadata
		}
	}

	if len(newMetadata) == 0 {
		failStep(&tpj.pipelineMetadata.MetadataRetrieval)
	}

	return newMetadata
}

func (tpj *tokenProcessingJob) retrieveTokenInfo(ctx context.Context, metadata persist.TokenMetadata) (string, string) {
	defer trackStepStatus(&tpj.pipelineMetadata.TokenInfoRetrieval)

	name, description := media.FindNameAndDescription(ctx, metadata)

	if name == "" {
		name = tpj.token.Name.String()
	}

	if description == "" {
		description = tpj.token.Description.String()
	}
	return name, description
}

func (tpj *tokenProcessingJob) cacheMediaObjects(ctx context.Context, metadata persist.TokenMetadata) ([]media.CachedMediaObject, error) {
	image, anim := media.KeywordsForChain(tpj.token.Chain, tpj.imageKeywords, tpj.animationKeywords)
	return media.CacheObjectsForMetadata(ctx, metadata, tpj.contractAddress, persist.TokenID(tpj.token.TokenID.String()), tpj.token.TokenURI, tpj.token.Chain, tpj.tp.ipfsClient, tpj.tp.arweaveClient, tpj.tp.stg, tpj.tp.tokenBucket, image, anim)
}

func (tpj *tokenProcessingJob) createMediaFromCachedObjects(ctx context.Context, objects []media.CachedMediaObject) persist.Media {
	defer trackStepStatus(&tpj.pipelineMetadata.CreateMediaFromCachedObjects)
	return media.CreateMediaFromCachedObjects(ctx, tpj.tp.tokenBucket, objects)
}

func (tpj *tokenProcessingJob) isNewMediaPreferable(ctx context.Context, media persist.Media) bool {
	defer trackStepStatus(&tpj.pipelineMetadata.MediaResultComparison)
	return !tpj.token.Media.IsServable() && media.IsServable()
}

func (tpj *tokenProcessingJob) persistMedia(ctx context.Context, tmetadata coredb.TokenMedium) error {
	if !tpj.isNewMediaPreferable(ctx, tmetadata.Media) {
		tmetadata.Active = false
	}

	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return tpj.updateTokenMetadataDB(ctx, tmetadata)
	})
	errGroup.Go(func() error {
		return tpj.updateJobDB(ctx, tmetadata)
	})

	return errGroup.Wait()
}

func (tpj *tokenProcessingJob) updateTokenMetadataDB(ctx context.Context, tmetadata coredb.TokenMedium) error {
	defer trackStepStatus(&tpj.pipelineMetadata.UpdateTokenMetadataDB)
	if !tmetadata.Active {
		return tpj.tp.queries.InsertTokenMedia(ctx, coredb.InsertTokenMediaParams{
			ID:              tmetadata.ID,
			Contract:        tmetadata.Contract,
			TokenID:         tmetadata.TokenID,
			Chain:           tmetadata.Chain,
			Metadata:        tmetadata.Metadata,
			Media:           tmetadata.Media,
			Name:            tmetadata.Name,
			Description:     tmetadata.Description,
			ProcessingJobID: tmetadata.ProcessingJobID,
		})
	}
	exists, err := tpj.tp.queries.IsExistsTokenMediaByTokenIdentifers(ctx, coredb.IsExistsTokenMediaByTokenIdentifersParams{
		Contract: tpj.token.Contract,
		TokenID:  tpj.token.TokenID,
		Chain:    tpj.token.Chain,
	})
	if err != nil {
		return err
	}

	if exists {
		return tpj.tp.queries.UpdateTokenMediaByTokenIdentifiers(ctx, coredb.UpdateTokenMediaByTokenIdentifiersParams{
			ID:              tmetadata.ID,
			Contract:        tpj.token.Contract,
			TokenID:         tmetadata.TokenID,
			Chain:           tpj.token.Chain,
			Metadata:        tmetadata.Metadata,
			Media:           tmetadata.Media,
			Name:            tmetadata.Name,
			Description:     tmetadata.Description,
			ProcessingJobID: tmetadata.ProcessingJobID,
		})
	}
	err = tpj.tp.queries.InsertTokenMedia(ctx, coredb.InsertTokenMediaParams{
		ID:              tmetadata.ID,
		Contract:        tmetadata.Contract,
		TokenID:         tmetadata.TokenID,
		Chain:           tmetadata.Chain,
		Metadata:        tmetadata.Metadata,
		Media:           tmetadata.Media,
		Name:            tmetadata.Name,
		Description:     tmetadata.Description,
		ProcessingJobID: tmetadata.ProcessingJobID,
	})
	if err != nil {
		return err
	}
	return tpj.tp.queries.UpdateTokenTokenMediaByTokenIdentifiers(ctx, coredb.UpdateTokenTokenMediaByTokenIdentifiersParams{
		TokenMedia: persist.DBIDToNullStr(tmetadata.ID),
		Contract:   tpj.token.Contract,
		TokenID:    tpj.token.TokenID,
		Chain:      tpj.token.Chain,
	})
}

func (tpj *tokenProcessingJob) updateJobDB(ctx context.Context, tmetadata coredb.TokenMedium) error {
	defer trackStepStatus(&tpj.pipelineMetadata.UpdateJobDB)
	p := persist.TokenProperties{}
	if tmetadata.Metadata != nil && len(tmetadata.Metadata) > 0 {
		p.HasMetadata = true
	}
	if tmetadata.Media.MediaType.IsValid() && tmetadata.Media.MediaURL != "" {
		p.HasPrimaryMedia = true
	}
	if tmetadata.Media.ThumbnailURL != "" {
		p.HasThumbnail = true
	}
	if tmetadata.Media.LivePreviewURL != "" {
		p.HasLiveRender = true
	}
	if tmetadata.Media.Dimensions.Valid() {
		p.HasDimensions = true
	}

	return tpj.tp.queries.InsertTokenProcessingJob(ctx, coredb.InsertTokenProcessingJobParams{
		ID:               tpj.id,
		TokenProperties:  p,
		PipelineMetadata: *tpj.pipelineMetadata,
		ProcessingCause:  tpj.cause,
		ProcessorVersion: "",
	})
}

func trackStepStatus(status *persist.PipelineStepStatus) func() {
	if status == nil {
		started := persist.PipelineStepStatusStarted
		status = &started
	}
	*status = persist.PipelineStepStatusStarted
	return func() {
		if *status == persist.PipelineStepStatusError {
			return
		}
		*status = persist.PipelineStepStatusSuccess
	}

}

func failStep(status *persist.PipelineStepStatus) {
	if status == nil {
		errored := persist.PipelineStepStatusError
		status = &errored
	}
	*status = persist.PipelineStepStatusError
}
