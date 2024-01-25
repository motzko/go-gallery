package publicapi

import (
	"context"
	"crypto/sha256"
	"encoding"
	"encoding/base64"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgtype"
	db "github.com/mikeydub/go-gallery/db/gen/coredb"
	"github.com/mikeydub/go-gallery/env"
	"github.com/mikeydub/go-gallery/event"
	"github.com/mikeydub/go-gallery/graphql/dataloader"
	"github.com/mikeydub/go-gallery/graphql/model"
	"github.com/mikeydub/go-gallery/service/auth"
	"github.com/mikeydub/go-gallery/service/persist"
	"github.com/mikeydub/go-gallery/service/persist/postgres"
	"github.com/mikeydub/go-gallery/util"
	"github.com/mikeydub/go-gallery/validate"
	"net"
	"time"
)

const maxCollectionsPerGallery = 1000

type GalleryAPI struct {
	repos     *postgres.Repositories
	queries   *db.Queries
	loaders   *dataloader.Loaders
	validator *validator.Validate
	ethClient *ethclient.Client
}

func (api GalleryAPI) CreateGallery(ctx context.Context, name, description *string, position string) (db.Gallery, error) {

	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"name":        validate.WithTag(name, "max=200"),
		"description": validate.WithTag(description, "max=1200"),
		"position":    validate.WithTag(position, "required"),
	}); err != nil {
		return db.Gallery{}, err
	}

	userID, err := getAuthenticatedUserID(ctx)
	if err != nil {
		return db.Gallery{}, err
	}

	gallery, err := api.repos.GalleryRepository.Create(ctx, db.GalleryRepoCreateParams{
		GalleryID:   persist.GenerateID(),
		Name:        util.FromPointer(name),
		Description: util.FromPointer(description),
		Position:    position,
		OwnerUserID: userID,
	})
	if err != nil {
		return db.Gallery{}, err
	}

	return gallery, nil
}

func (api GalleryAPI) UpdateGallery(ctx context.Context, update model.UpdateGalleryInput) (db.Gallery, error) {

	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"galleryID":           validate.WithTag(update.GalleryID, "required"),
		"name":                validate.WithTag(update.Name, "omitempty,max=200"),
		"description":         validate.WithTag(update.Description, "omitempty,max=1200"),
		"deleted_collections": validate.WithTag(update.DeletedCollections, "omitempty,unique"),
		"created_collections": validate.WithTag(update.CreatedCollections, "omitempty,created_collections"),
	}); err != nil {
		return db.Gallery{}, err
	}

	events := make([]db.Event, 0, len(update.CreatedCollections)+len(update.UpdatedCollections)+1)

	curGal, err := api.loaders.GetGalleryByIdBatch.Load(update.GalleryID)
	if err != nil {
		return db.Gallery{}, err
	}

	userID, err := getAuthenticatedUserID(ctx)
	if err != nil {
		return db.Gallery{}, err
	}

	if curGal.OwnerUserID != userID {
		return db.Gallery{}, fmt.Errorf("user %s is not the owner of gallery %s", userID, update.GalleryID)
	}

	tx, err := api.repos.BeginTx(ctx)
	if err != nil {
		return db.Gallery{}, err
	}
	defer tx.Rollback(ctx)

	q := api.queries.WithTx(tx)

	// then delete collections
	if len(update.DeletedCollections) > 0 {
		err = q.DeleteCollections(ctx, util.StringersToStrings(update.DeletedCollections))
		if err != nil {
			return db.Gallery{}, err
		}
	}

	// create collections
	mappedIDs := make(map[persist.DBID]persist.DBID)
	for _, c := range update.CreatedCollections {
		collectionID, err := q.CreateCollection(ctx, db.CreateCollectionParams{
			ID:             persist.GenerateID(),
			Name:           persist.StrPtrToNullStr(&c.Name),
			CollectorsNote: persist.StrPtrToNullStr(&c.CollectorsNote),
			OwnerUserID:    curGal.OwnerUserID,
			GalleryID:      update.GalleryID,
			Layout:         modelToTokenLayout(c.Layout),
			Hidden:         c.Hidden,
			Nfts:           c.Tokens,
			TokenSettings:  modelToTokenSettings(c.TokenSettings),
		})
		if err != nil {
			return db.Gallery{}, err
		}

		events = append(events, db.Event{
			ID:             persist.GenerateID(),
			ActorID:        persist.DBIDToNullStr(userID),
			Action:         persist.ActionCollectionCreated,
			ResourceTypeID: persist.ResourceTypeCollection,
			SubjectID:      collectionID,
			CollectionID:   collectionID,
			GalleryID:      update.GalleryID,
			Data: persist.EventData{
				CollectionTokenIDs:       c.Tokens,
				CollectionCollectorsNote: c.CollectorsNote,
			},
		})

		mappedIDs[c.GivenID] = collectionID
	}

	// update collections

	if len(update.UpdatedCollections) > 0 {
		collEvents, err := updateCollectionsInfoAndTokens(ctx, q, userID, update.GalleryID, update.UpdatedCollections)
		if err != nil {
			return db.Gallery{}, err
		}

		events = append(events, collEvents...)
	}

	// order collections
	for i, c := range update.Order {
		if newID, ok := mappedIDs[c]; ok {
			update.Order[i] = newID
		}
	}

	params := db.UpdateGalleryInfoParams{
		ID: update.GalleryID,
	}

	util.SetConditionalValue(update.Name, &params.Name, &params.NameSet)
	util.SetConditionalValue(update.Description, &params.Description, &params.DescriptionSet)

	err = q.UpdateGalleryInfo(ctx, params)
	if err != nil {
		return db.Gallery{}, err
	}

	if update.Name != nil || update.Description != nil {
		e := db.Event{
			ID:             persist.GenerateID(),
			ActorID:        persist.DBIDToNullStr(userID),
			Action:         persist.ActionGalleryInfoUpdated,
			ResourceTypeID: persist.ResourceTypeGallery,
			GalleryID:      update.GalleryID,
			SubjectID:      update.GalleryID,
		}

		change := false

		if update.Name != nil && *update.Name != curGal.Name {
			e.Data.GalleryName = update.Name
			change = true
		}

		if update.Description != nil && *update.Description != curGal.Description {
			e.Data.GalleryDescription = update.Description
			change = true
		}

		if change {
			events = append(events, e)
		}

	}

	asList := persist.DBIDList(update.Order)

	if len(asList) > 0 {
		err = q.UpdateGalleryCollections(ctx, db.UpdateGalleryCollectionsParams{
			GalleryID:   update.GalleryID,
			Collections: asList,
		})
		if err != nil {
			return db.Gallery{}, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return db.Gallery{}, err
	}

	newGall, err := api.loaders.GetGalleryByIdBatch.Load(update.GalleryID)
	if err != nil {
		return db.Gallery{}, err
	}

	if update.Caption != nil && *update.Caption == "" {
		update.Caption = nil
	}
	err = event.DispatchMany(ctx, events, update.EditID)
	if err != nil {
		return db.Gallery{}, err
	}

	return newGall, nil
}

func (api GalleryAPI) PublishGallery(ctx context.Context, update model.PublishGalleryInput) error {

	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"galleryID": validate.WithTag(update.GalleryID, "required"),
		"editID":    validate.WithTag(update.EditID, "required"),
	}); err != nil {
		return err
	}

	_, err := publishEventGroup(ctx, update.EditID, persist.ActionGalleryUpdated, update.Caption)
	if err != nil {
		return err
	}

	return nil
}
func updateCollectionsInfoAndTokens(ctx context.Context, q *db.Queries, actor, gallery persist.DBID, update []*model.UpdateCollectionInput) ([]db.Event, error) {

	events := make([]db.Event, 0)

	dbids, err := util.Map(update, func(u *model.UpdateCollectionInput) (string, error) {
		return u.Dbid.String(), nil
	})
	if err != nil {
		return nil, err
	}

	collectorNotes, err := util.Map(update, func(u *model.UpdateCollectionInput) (string, error) {
		return u.CollectorsNote, nil
	})
	if err != nil {
		return nil, err
	}

	layouts, err := util.Map(update, func(u *model.UpdateCollectionInput) (pgtype.JSONB, error) {
		return persist.ToJSONB(modelToTokenLayout(u.Layout))
	})
	if err != nil {
		return nil, err
	}

	tokenSettings, err := util.Map(update, func(u *model.UpdateCollectionInput) (pgtype.JSONB, error) {
		settings := modelToTokenSettings(u.TokenSettings)
		return persist.ToJSONB(settings)
	})
	if err != nil {
		return nil, err
	}

	hiddens, err := util.Map(update, func(u *model.UpdateCollectionInput) (bool, error) {
		return u.Hidden, nil
	})
	if err != nil {
		return nil, err
	}

	names, err := util.Map(update, func(u *model.UpdateCollectionInput) (string, error) {
		return u.Name, nil
	})
	if err != nil {
		return nil, err
	}

	for _, collection := range update {
		curCol, err := q.GetCollectionById(ctx, collection.Dbid)
		if err != nil {
			return nil, err
		}

		// add event if collectors note updated
		if collection.CollectorsNote != "" && collection.CollectorsNote != curCol.CollectorsNote.String {
			events = append(events, db.Event{
				ActorID:        persist.DBIDToNullStr(actor),
				ResourceTypeID: persist.ResourceTypeCollection,
				SubjectID:      collection.Dbid,
				Action:         persist.ActionCollectorsNoteAddedToCollection,
				CollectionID:   collection.Dbid,
				GalleryID:      gallery,
				Data: persist.EventData{
					CollectionCollectorsNote: collection.CollectorsNote,
				},
			})
		}
	}

	err = q.UpdateCollectionsInfo(ctx, db.UpdateCollectionsInfoParams{
		Ids:             dbids,
		Names:           names,
		CollectorsNotes: collectorNotes,
		Layouts:         layouts,
		TokenSettings:   tokenSettings,
		Hidden:          hiddens,
	})
	if err != nil {
		return nil, err
	}

	for _, collection := range update {
		curTokens, err := q.GetCollectionTokensByCollectionID(ctx, collection.Dbid)
		if err != nil {
			return nil, err
		}

		err = q.UpdateCollectionTokens(ctx, db.UpdateCollectionTokensParams{
			ID:   collection.Dbid,
			Nfts: collection.Tokens,
		})
		if err != nil {
			return nil, err
		}

		diff := util.Difference(curTokens, collection.Tokens)

		if len(diff) > 0 {
			events = append(events, db.Event{
				ResourceTypeID: persist.ResourceTypeCollection,
				SubjectID:      collection.Dbid,
				Action:         persist.ActionTokensAddedToCollection,
				ActorID:        persist.DBIDToNullStr(actor),
				CollectionID:   collection.Dbid,
				GalleryID:      gallery,
				Data: persist.EventData{
					CollectionTokenIDs: diff,
				},
			})
		}
	}
	return events, nil
}

func (api GalleryAPI) DeleteGallery(ctx context.Context, galleryID persist.DBID) error {

	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"galleryID": validate.WithTag(galleryID, "required"),
	}); err != nil {
		return err
	}

	userID, err := getAuthenticatedUserID(ctx)
	if err != nil {
		return err
	}

	err = api.repos.GalleryRepository.Delete(ctx, db.GalleryRepoDeleteParams{
		GalleryID:   galleryID,
		OwnerUserID: userID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (api GalleryAPI) GetGalleryById(ctx context.Context, galleryID persist.DBID) (*db.Gallery, error) {
	// Validate
	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"galleryID": validate.WithTag(galleryID, "required"),
	}); err != nil {
		return nil, err
	}

	gallery, err := api.loaders.GetGalleryByIdBatch.Load(galleryID)
	if err != nil {
		return nil, err
	}

	return &gallery, nil
}

func (api GalleryAPI) GetViewerGalleryById(ctx context.Context, galleryID persist.DBID) (*db.Gallery, error) {
	// Validate
	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"galleryID": validate.WithTag(galleryID, "required"),
	}); err != nil {
		return nil, err
	}

	userID, err := getAuthenticatedUserID(ctx)

	if err != nil {
		return nil, persist.ErrGalleryNotFound{ID: galleryID}
	}

	gallery, err := api.loaders.GetGalleryByIdBatch.Load(galleryID)
	if err != nil {
		return nil, err
	}

	if userID != gallery.OwnerUserID {
		return nil, persist.ErrGalleryNotFound{ID: galleryID}
	}

	return &gallery, nil
}

func (api GalleryAPI) GetGalleryByCollectionId(ctx context.Context, collectionID persist.DBID) (*db.Gallery, error) {
	// Validate
	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"collectionID": validate.WithTag(collectionID, "required"),
	}); err != nil {
		return nil, err
	}

	gallery, err := api.loaders.GetGalleryByCollectionIdBatch.Load(collectionID)
	if err != nil {
		return nil, err
	}

	return &gallery, nil
}

func (api GalleryAPI) GetGalleriesByUserId(ctx context.Context, userID persist.DBID) ([]db.Gallery, error) {
	// Validate
	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"userID": validate.WithTag(userID, "required"),
	}); err != nil {
		return nil, err
	}

	galleries, err := api.loaders.GetGalleriesByUserIdBatch.Load(userID)
	if err != nil {
		return nil, err
	}

	return galleries, nil
}

func (api GalleryAPI) GetTokenPreviewsByGalleryID(ctx context.Context, galleryID persist.DBID) ([]db.TokenMedia, error) {
	// Validate
	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"galleryID": validate.WithTag(galleryID, "required"),
	}); err != nil {
		return nil, err
	}

	medias, err := api.loaders.GetGalleryTokenMediasByGalleryIDBatch.Load(galleryID)
	if err != nil {
		return nil, err
	}

	return medias, nil
}

func (api GalleryAPI) UpdateGalleryCollections(ctx context.Context, galleryID persist.DBID, collections []persist.DBID) error {
	// Validate
	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"galleryID":   validate.WithTag(galleryID, "required"),
		"collections": validate.WithTag(collections, fmt.Sprintf("required,unique,max=%d", maxCollectionsPerGallery)),
	}); err != nil {
		return err
	}

	userID, err := getAuthenticatedUserID(ctx)
	if err != nil {
		return err
	}

	update := persist.GalleryTokenUpdateInput{Collections: collections}

	err = api.repos.GalleryRepository.Update(ctx, galleryID, userID, update)
	if err != nil {
		return err
	}

	return nil
}

func (api GalleryAPI) UpdateGalleryInfo(ctx context.Context, galleryID persist.DBID, name, description *string) error {
	// Validate
	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"galleryID":   validate.WithTag(galleryID, "required"),
		"name":        validate.WithTag(name, "max=200"),
		"description": validate.WithTag(description, "max=1200"),
	}); err != nil {
		return err
	}

	var nullName, nullDesc string
	var nameSet, descSet bool

	if name != nil {
		nullName = *name
		nameSet = true
	}
	if description != nil {
		nullDesc = *description
		descSet = true
	}

	err := api.queries.UpdateGalleryInfo(ctx, db.UpdateGalleryInfoParams{
		ID:             galleryID,
		Name:           nullName,
		Description:    nullDesc,
		NameSet:        nameSet,
		DescriptionSet: descSet,
	})
	if err != nil {
		return err
	}
	return nil
}

func (api GalleryAPI) UpdateGalleryHidden(ctx context.Context, galleryID persist.DBID, hidden bool) (db.Gallery, error) {
	// Validate
	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"galleryID": validate.WithTag(galleryID, "required"),
	}); err != nil {
		return db.Gallery{}, err
	}

	gallery, err := api.queries.UpdateGalleryHidden(ctx, db.UpdateGalleryHiddenParams{
		ID:     galleryID,
		Hidden: hidden,
	})
	if err != nil {
		return db.Gallery{}, err
	}

	return gallery, nil
}

func (api GalleryAPI) UpdateGalleryPositions(ctx context.Context, positions []*model.GalleryPositionInput) error {
	// Validate
	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"positions": validate.WithTag(positions, "required,min=1"),
	}); err != nil {
		return err
	}

	user, err := getAuthenticatedUserID(ctx)
	if err != nil {
		return err
	}

	ids := make([]string, len(positions))
	pos := make([]string, len(positions))
	for i, position := range positions {
		ids[i] = position.GalleryID.String()
		pos[i] = position.Position
	}

	tx, err := api.repos.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := api.queries.WithTx(tx)

	err = q.UpdateGalleryPositions(ctx, db.UpdateGalleryPositionsParams{
		GalleryIds:  ids,
		Positions:   pos,
		OwnerUserID: user,
	})
	if err != nil {
		return err
	}

	areDuplicates, err := q.UserHasDuplicateGalleryPositions(ctx, user)
	if err != nil {
		return err
	}
	if areDuplicates {
		return fmt.Errorf("gallery positions are not unique for user %s", user)
	}

	return tx.Commit(ctx)
}

func (api GalleryAPI) ViewGallery(ctx context.Context, galleryID persist.DBID) (db.Gallery, error) {
	// Validate
	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"galleryID": validate.WithTag(galleryID, "required"),
	}); err != nil {
		return db.Gallery{}, err
	}

	gallery, err := api.loaders.GetGalleryByIdBatch.Load(galleryID)
	if err != nil {
		return db.Gallery{}, err
	}

	gc := util.MustGetGinContext(ctx)

	if auth.GetUserAuthedFromCtx(gc) {
		userID, err := getAuthenticatedUserID(ctx)
		if err != nil {
			return db.Gallery{}, err
		}

		if gallery.OwnerUserID != userID {
			// only view gallery if the user hasn't already viewed it in this most recent notification period

			err = event.Dispatch(ctx, db.Event{
				ActorID:        persist.DBIDToNullStr(userID),
				ResourceTypeID: persist.ResourceTypeGallery,
				SubjectID:      galleryID,
				Action:         persist.ActionViewedGallery,
				GalleryID:      galleryID,
			})
			if err != nil {
				return gallery, err
			}
		}
	} else {
		err := event.Dispatch(ctx, db.Event{
			ResourceTypeID: persist.ResourceTypeGallery,
			SubjectID:      galleryID,
			Action:         persist.ActionViewedGallery,
			GalleryID:      galleryID,
			ExternalID:     persist.StrPtrToNullStr(getExternalID(ctx)),
		})
		if err != nil {
			return gallery, err
		}
	}

	return gallery, nil
}

func (api GalleryAPI) GetGalleriesDisplayingCommunityID(ctx context.Context, communityID persist.DBID, before *string, after *string,
	first *int, last *int) ([]db.Gallery, [][]persist.Media, [][]time.Time, PageInfo, error) {
	// Validate
	if err := validate.ValidateFields(api.validator, validate.ValidationMap{
		"communityID": validate.WithTag(communityID, "required"),
	}); err != nil {
		return nil, nil, nil, PageInfo{}, err
	}

	if err := validatePaginationParams(api.validator, first, last); err != nil {
		return nil, nil, nil, PageInfo{}, err
	}

	queryFunc := func(params floatIDPagingParams) ([]db.GetGalleriesDisplayingCommunityIDPaginateBatchRow, error) {
		return api.loaders.GetGalleriesDisplayingCommunityIDPaginateBatch.Load(db.GetGalleriesDisplayingCommunityIDPaginateBatchParams{
			CommunityID:        communityID,
			Limit:              params.Limit,
			CurBeforeRelevance: params.CursorBeforeFloat,
			CurBeforeID:        params.CursorBeforeID,
			CurAfterRelevance:  params.CursorAfterFloat,
			CurAfterID:         params.CursorAfterID,
			PagingForward:      params.PagingForward,
		})
	}

	countFunc := func() (int, error) {
		total, err := api.loaders.CountGalleriesDisplayingCommunityIDBatch.Load(communityID)
		return int(total), err
	}

	cursorFunc := func(r db.GetGalleriesDisplayingCommunityIDPaginateBatchRow) (float64, persist.DBID, error) {
		return r.Relevance, r.Gallery.ID, nil
	}

	paginator := floatIDPaginator[db.GetGalleriesDisplayingCommunityIDPaginateBatchRow]{
		QueryFunc:  queryFunc,
		CursorFunc: cursorFunc,
		CountFunc:  countFunc,
	}

	rows, pageInfo, err := paginator.paginate(before, after, first, last)
	if err != nil {
		return nil, nil, nil, PageInfo{}, err
	}

	galleries := make([]db.Gallery, len(rows))
	medias := make([][]persist.Media, len(rows))
	mediasLastUpdated := make([][]time.Time, len(rows))
	for i, row := range rows {
		galleries[i] = row.Gallery
		medias[i], mediasLastUpdated[i] = getPreviewMediaForCommunityGallery(ctx, row.CommunityTokenIds, row.CommunityMedias, row.CommunityMediaLastUpdated, row.AllTokenIds, row.AllMedias, row.AllMediaLastUpdated)
	}

	return galleries, medias, mediasLastUpdated, pageInfo, nil
}

// getPreviewMediaForCommunityGallery returns a slice of media that contains all community media first, then all other media
func getPreviewMediaForCommunityGallery(ctx context.Context, communityTokenIDs []persist.DBID, communityMedia []persist.Media, communityLastUpdated []time.Time,
	allTokenIDs []persist.DBID, allMedia []persist.Media, allLastUpdated []time.Time) ([]persist.Media, []time.Time) {
	media := make([]persist.Media, 0, len(communityTokenIDs)+len(allTokenIDs))
	lastUpdated := make([]time.Time, 0, len(communityTokenIDs)+len(allTokenIDs))

	communityIDMap := make(map[persist.DBID]bool)
	for _, id := range communityTokenIDs {
		communityIDMap[id] = true
	}

	for i, m := range communityMedia {
		media = append(media, m)
		lastUpdated = append(lastUpdated, communityLastUpdated[i])
	}

	for i, m := range allMedia {
		if _, ok := communityIDMap[allTokenIDs[i]]; !ok {
			media = append(media, m)
			lastUpdated = append(lastUpdated, allLastUpdated[i])
		}
	}

	return media, lastUpdated
}

func getExternalID(ctx context.Context) *string {
	gc := util.MustGetGinContext(ctx)
	if ip := net.ParseIP(gc.ClientIP()); ip != nil && !ip.IsPrivate() {
		hash := sha256.New()
		hash.Write([]byte(env.GetString("BACKEND_SECRET") + ip.String()))
		res, _ := hash.(encoding.BinaryMarshaler).MarshalBinary()
		externalID := base64.StdEncoding.EncodeToString(res)
		return &externalID
	}
	return nil
}

func modelToTokenLayout(u *model.CollectionLayoutInput) persist.TokenLayout {
	sectionLayout := make([]persist.CollectionSectionLayout, len(u.SectionLayout))
	for i, layout := range u.SectionLayout {
		sectionLayout[i] = persist.CollectionSectionLayout{
			Columns:    persist.NullInt32(layout.Columns),
			Whitespace: layout.Whitespace,
		}
	}
	return persist.TokenLayout{
		Sections:      persist.StandardizeCollectionSections(u.Sections),
		SectionLayout: sectionLayout,
	}
}

func modelToTokenSettings(u []*model.CollectionTokenSettingsInput) map[persist.DBID]persist.CollectionTokenSettings {
	settings := make(map[persist.DBID]persist.CollectionTokenSettings)
	for _, tokenSetting := range u {
		settings[tokenSetting.TokenID] = persist.CollectionTokenSettings{RenderLive: tokenSetting.RenderLive, HighDefinition: tokenSetting.HighDefinition}
	}
	return settings
}
