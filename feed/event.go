package feed

import (
	"context"

	"github.com/mikeydub/go-gallery/db/sqlc"
	"github.com/mikeydub/go-gallery/service/persist"
	"github.com/mikeydub/go-gallery/service/persist/postgres"
	"github.com/mikeydub/go-gallery/service/task"
	"github.com/spf13/viper"
)

type EventBuilder struct {
	eventRepo *postgres.EventRepository
	feedRepo  *postgres.FeedRepository
}

func NewEventBuilder() *EventBuilder {
	queries := sqlc.New(postgres.NewPgxClient())
	return &EventBuilder{
		eventRepo: &postgres.EventRepository{Queries: queries},
		feedRepo:  &postgres.FeedRepository{Queries: queries},
	}
}

func (b *EventBuilder) NewEvent(ctx context.Context, message task.FeedMessage) (*sqlc.FeedEvent, error) {
	event, err := b.eventRepo.Get(ctx, message.ID)

	if err != nil {
		return nil, err
	}

	switch event.Action {
	case persist.ActionUserCreated:
		return b.createUserCreatedEvent(ctx, event)
	case persist.ActionUserFollowedUsers:
		return b.createUserFollowedUsersEvent(ctx, event)
	case persist.ActionCollectorsNoteAddedToToken:
		return b.createCollectorsNoteAddedToTokenEvent(ctx, event)
	case persist.ActionCollectionCreated:
		return b.createCollectionCreatedEvent(ctx, event)
	case persist.ActionCollectorsNoteAddedToCollection:
		return b.createCollectorsNoteAddedToCollectionEvent(ctx, event)
	case persist.ActionTokensAddedToCollection:
		return b.createTokensAddedToCollectionEvent(ctx, event)
	default:
		return nil, persist.ErrUnknownAction{Action: event.Action}
	}
}

func (b *EventBuilder) createUserCreatedEvent(ctx context.Context, event sqlc.Event) (*sqlc.FeedEvent, error) {
	isActive, err := b.eventRepo.WindowActive(ctx, event)

	// more recent events are bufferred
	if err != nil || isActive {
		return nil, err
	}

	feedEvent, err := b.feedRepo.LastEventFrom(ctx, event)
	if err != nil {
		return nil, err
	}

	// only want to store this event type once
	if feedEvent != nil {
		return nil, nil
	}

	return b.feedRepo.Add(ctx, sqlc.FeedEvent{
		ID:        persist.GenerateID(),
		OwnerID:   event.ActorID,
		Action:    event.Action,
		EventTime: event.CreatedAt,
		Data:      persist.FeedEventData{UserBio: event.Data.UserBio},
		EventIds:  persist.DBIDList{event.ID},
	})
}

func (b *EventBuilder) createUserFollowedUsersEvent(ctx context.Context, event sqlc.Event) (*sqlc.FeedEvent, error) {
	isActive, err := b.eventRepo.WindowActive(ctx, event)

	// more recent events are bufferred
	if err != nil || isActive {
		return nil, err
	}

	feedEvent, err := b.feedRepo.LastEventFrom(ctx, event)
	if err != nil {
		return nil, err
	}

	events := []sqlc.Event{event}

	if feedEvent != nil {
		events, err = b.eventRepo.EventsInWindow(ctx, event.ID, viper.GetInt("FEED_WINDOW_SIZE"))
		if err != nil {
			return nil, err
		}
	}

	var followedIDs []persist.DBID
	var followedBack []bool
	var eventIDs []persist.DBID

	for _, event := range events {
		if !event.Data.UserRefollowed {
			followedIDs = append(followedIDs, event.SubjectID)
			followedBack = append(followedBack, event.Data.UserFollowedBack)
			eventIDs = append(eventIDs, event.ID)
		}
	}

	if len(followedIDs) < 1 {
		return nil, nil
	}

	return b.feedRepo.Add(ctx, sqlc.FeedEvent{
		ID:      persist.GenerateID(),
		OwnerID: event.ActorID,
		Action:  event.Action,
		Data: persist.FeedEventData{
			UserFollowedIDs:  followedIDs,
			UserFollowedBack: followedBack,
		},
		EventTime: event.CreatedAt,
		EventIds:  eventIDs,
	})
}

func (b *EventBuilder) createCollectorsNoteAddedToTokenEvent(ctx context.Context, event sqlc.Event) (*sqlc.FeedEvent, error) {
	isActive, err := b.eventRepo.WindowActiveForSubject(ctx, event)

	// more recent events are bufferred
	if err != nil || isActive {
		return nil, err
	}

	// don't present empty notes
	if event.Data.TokenCollectorsNote == "" {
		return nil, nil
	}

	// token should be edited in the context of a collection
	if event.Data.TokenCollectionID == "" {
		return nil, nil
	}

	feedEvent, err := b.feedRepo.LastTokenEventFromEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	// only show if note has changed
	if feedEvent != nil && feedEvent.Data.TokenNewCollectorsNote == event.Data.TokenCollectorsNote {
		return nil, nil
	}

	return b.feedRepo.Add(ctx, sqlc.FeedEvent{
		ID:      persist.GenerateID(),
		OwnerID: event.ActorID,
		Action:  event.Action,
		Data: persist.FeedEventData{
			TokenID:                event.SubjectID,
			TokenCollectionID:      event.Data.TokenCollectionID,
			TokenNewCollectorsNote: event.Data.TokenCollectorsNote,
		},
		EventTime: event.CreatedAt,
		EventIds:  persist.DBIDList{event.ID},
	})
}

func (b *EventBuilder) createCollectionCreatedEvent(ctx context.Context, event sqlc.Event) (*sqlc.FeedEvent, error) {
	isActive, err := b.eventRepo.WindowActiveForSubject(ctx, event)

	// more recent events are bufferred
	if err != nil || isActive {
		return nil, err
	}

	// don't show empty collections
	if len(event.Data.CollectionTokenIDs) < 1 {
		return nil, nil
	}

	return b.feedRepo.Add(ctx, sqlc.FeedEvent{
		ID:      persist.GenerateID(),
		OwnerID: event.ActorID,
		Action:  event.Action,
		Data: persist.FeedEventData{
			CollectionID:                event.SubjectID,
			CollectionTokenIDs:          event.Data.CollectionTokenIDs,
			CollectionNewTokenIDs:       event.Data.CollectionTokenIDs,
			CollectionNewCollectorsNote: event.Data.CollectionCollectorsNote,
		},
		EventTime: event.CreatedAt,
		EventIds:  persist.DBIDList{event.ID},
	})
}

func (b *EventBuilder) createCollectorsNoteAddedToCollectionEvent(ctx context.Context, event sqlc.Event) (*sqlc.FeedEvent, error) {
	isActive, err := b.eventRepo.WindowActiveForSubject(ctx, event)

	// more recent events are bufferred
	if err != nil || isActive {
		return nil, err
	}

	// don't present empty notes
	if event.Data.CollectionCollectorsNote == "" {
		return nil, nil
	}

	feedEvent, err := b.feedRepo.LastCollectionEventFromEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	// only show if note has changed
	if feedEvent != nil && feedEvent.Data.CollectionNewCollectorsNote == event.Data.CollectionCollectorsNote {
		return nil, nil
	}

	return b.feedRepo.Add(ctx, sqlc.FeedEvent{
		ID:      persist.GenerateID(),
		OwnerID: event.ActorID,
		Action:  event.Action,
		Data: persist.FeedEventData{
			CollectionID:                event.SubjectID,
			CollectionNewCollectorsNote: event.Data.CollectionCollectorsNote,
		},
		EventTime: event.CreatedAt,
		EventIds:  persist.DBIDList{event.ID},
	})
}

func (b *EventBuilder) createTokensAddedToCollectionEvent(ctx context.Context, event sqlc.Event) (*sqlc.FeedEvent, error) {
	isActive, err := b.eventRepo.WindowActiveForSubject(ctx, event)

	// more recent events are bufferred
	if err != nil || isActive {
		return nil, err
	}

	// don't show empty collections
	if len(event.Data.CollectionTokenIDs) < 1 {
		return nil, nil
	}

	feedEvent, err := b.feedRepo.LastCollectionEventFromEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	createEvent, err := b.feedRepo.LastCollectionEvent(ctx,
		event.ActorID, persist.ActionCollectionCreated, event.SubjectID, event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	var added []persist.DBID

	if feedEvent != nil {
		// compare against last token added event
		added = newTokens(event.Data.CollectionTokenIDs, feedEvent.Data.CollectionTokenIDs)
	} else if createEvent != nil {
		// compare against the create collection event
		added = newTokens(event.Data.CollectionTokenIDs, createEvent.Data.CollectionTokenIDs)
	} else {
		// don't have the create event for whatever reason, so treat all tokens as new
		added = event.Data.CollectionTokenIDs
	}

	// only show if new tokens were added
	if len(added) < 1 {
		return nil, nil
	}

	return b.feedRepo.Add(ctx, sqlc.FeedEvent{
		ID:      persist.GenerateID(),
		OwnerID: event.ActorID,
		Action:  event.Action,
		Data: persist.FeedEventData{
			CollectionID:          event.SubjectID,
			CollectionTokenIDs:    event.Data.CollectionTokenIDs,
			CollectionNewTokenIDs: added,
		},
		EventTime: event.CreatedAt,
		EventIds:  persist.DBIDList{event.ID},
	})
}

func newTokens(tokens []persist.DBID, otherTokens []persist.DBID) []persist.DBID {
	newTokens := make([]persist.DBID, 0)

	for _, token := range tokens {
		var exists bool

		for _, otherToken := range otherTokens {
			if token == otherToken {
				exists = true
				break
			}
		}

		if !exists {
			newTokens = append(newTokens, token)
		}
	}

	return newTokens
}
