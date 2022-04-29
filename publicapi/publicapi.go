package publicapi

import (
	"context"
	"fmt"

	"github.com/mikeydub/go-gallery/db/sqlc"

	"cloud.google.com/go/storage"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/everFinance/goar"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/mikeydub/go-gallery/graphql/dataloader"
	"github.com/mikeydub/go-gallery/service/auth"
	"github.com/mikeydub/go-gallery/service/multichain"
	"github.com/mikeydub/go-gallery/service/persist"
	"github.com/mikeydub/go-gallery/util"
	"github.com/mikeydub/go-gallery/validate"
)

const apiContextKey = "publicapi.api"

type PublicAPI struct {
	repos      *persist.Repositories
	queries    *sqlc.Queries
	loaders    *dataloader.Loaders
	validator  *validator.Validate
	Collection *CollectionAPI
	Gallery    *GalleryAPI
	User       *UserAPI
	Nft        *NftAPI
	Address    *AddressApi
}

func AddTo(ctx *gin.Context, repos *persist.Repositories, queries *sqlc.Queries, ethClient *ethclient.Client, ipfsClient *shell.Shell, arweaveClient *goar.Client, storageClient *storage.Client, multichainProvider *multichain.Provider) {
	loaders := dataloader.NewLoaders(ctx, queries)
	validator := newValidator()

	api := &PublicAPI{
		repos:      repos,
		queries:    queries,
		loaders:    loaders,
		validator:  validator,
		Collection: &CollectionAPI{repos: repos, queries: queries, loaders: loaders, validator: validator, ethClient: ethClient},
		Gallery:    &GalleryAPI{repos: repos, queries: queries, loaders: loaders, validator: validator, ethClient: ethClient},
		User:       &UserAPI{repos: repos, queries: queries, loaders: loaders, validator: validator, ethClient: ethClient, ipfsClient: ipfsClient, arweaveClient: arweaveClient, storageClient: storageClient},
		Nft:        &NftAPI{repos: repos, queries: queries, loaders: loaders, validator: validator, ethClient: ethClient, multichainProvider: multichainProvider},
		Address:    &AddressApi{repos: repos, queries: queries, loaders: loaders, validator: validator, ethClient: ethClient},
	}

	ctx.Set(apiContextKey, api)
}

func For(ctx context.Context) *PublicAPI {
	gc := util.GinContextFromContext(ctx)
	return gc.Value(apiContextKey).(*PublicAPI)
}

func newValidator() *validator.Validate {
	v := validator.New()
	validate.RegisterCustomValidators(v)
	return v
}

func getAuthenticatedUser(ctx context.Context) (persist.DBID, error) {
	gc := util.GinContextFromContext(ctx)
	authError := auth.GetAuthErrorFromCtx(gc)

	if authError != nil {
		return "", authError
	}

	userID := auth.GetUserIDFromCtx(gc)
	return userID, nil
}

type validationMap map[string]struct {
	value interface{}
	tag   string
}

func validateFields(validator *validator.Validate, fields validationMap) error {
	validationErr := ErrInvalidInput{}
	foundErrors := false

	for k, v := range fields {
		err := validator.Var(v.value, v.tag)
		if err != nil {
			foundErrors = true
			validationErr.Append(k, err.Error())
		}
	}

	if foundErrors {
		return validationErr
	}

	return nil
}

type ErrInvalidInput struct {
	Parameters []string
	Reasons    []string
}

func (e *ErrInvalidInput) Append(parameter string, reason string) {
	e.Parameters = append(e.Parameters, parameter)
	e.Reasons = append(e.Reasons, reason)
}

func (e ErrInvalidInput) Error() string {
	str := "invalid input:\n"

	for i, _ := range e.Parameters {
		str += fmt.Sprintf("    parameter: %s, reason: %s\n", e.Parameters[i], e.Reasons[i])
	}

	return str
}
