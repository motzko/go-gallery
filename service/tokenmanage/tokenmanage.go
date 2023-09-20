package tokenmanage

import (
	"context"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"

	"github.com/mikeydub/go-gallery/service/limiters"
	"github.com/mikeydub/go-gallery/service/persist"
	"github.com/mikeydub/go-gallery/service/redis"
	"github.com/mikeydub/go-gallery/service/task"
	"github.com/mikeydub/go-gallery/service/throttle"
)

type Manager struct {
	cache           *redis.Cache
	processRegistry *registry
	taskClient      *cloudtasks.Client
	throttle        *throttle.Locker
	// delayer sets the linear delay for retrying tokens up to MaxRetries
	delayer *limiters.KeyRateLimiter
	// maxRetries is the maximum number of times a token can be reenqueued before it is not retried again
	maxRetries int
}

func New(ctx context.Context, taskClient *cloudtasks.Client) *Manager {
	cache := redis.NewCache(redis.TokenManageCache)
	return &Manager{
		cache:           cache,
		processRegistry: &registry{cache},
		taskClient:      taskClient,
		throttle:        throttle.NewThrottleLocker(cache, 30*time.Minute),
	}
}

func NewWithRetries(ctx context.Context, taskClient *cloudtasks.Client, maxRetries int) *Manager {
	m := New(ctx, taskClient)
	m.maxRetries = maxRetries
	m.delayer = limiters.NewKeyRateLimiter(ctx, m.cache, "retry", 2, 1*time.Minute)
	return m
}

// Processing returns true if the token is processing or enqueued.
func (m Manager) Processing(ctx context.Context, token persist.TokenIdentifiers) bool {
	p, _ := m.processRegistry.processing(ctx, token)
	return p
}

// StartProcessing marks a token as processing. It returns a callback that must be called when work on the token is finished in order to mark
// it as finished. If withRetry is true, the callback will attempt to reenqueue the token if an error is passed. attemps is ignored when MaxRetries
// is set to the default value of 0.
func (m Manager) StartProcessing(ctx context.Context, token persist.TokenIdentifiers, attempts int) (error, func(err error) error) {
	err := m.throttle.Lock(ctx, "lock:"+token.String())
	if err != nil {
		return err, nil
	}

	stop := make(chan bool)
	done := make(chan bool)
	tick := time.NewTicker(10 * time.Second)

	go func() {
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				m.processRegistry.keepAlive(ctx, token)
			case <-stop:
				close(done)
				return
			}
		}
	}()

	callback := func(err error) error {
		close(stop)
		<-done
		m.tryRetry(ctx, token, err, attempts)
		m.throttle.Unlock(ctx, "lock:"+token.String())
		return nil
	}

	return nil, callback
}

// SubmitUser enqueues a user's tokens for processing.
func (m Manager) SubmitUser(ctx context.Context, userID persist.DBID, tokenIDs []persist.DBID, tokens []persist.TokenIdentifiers) error {
	m.processRegistry.setManyEnqueue(ctx, tokens)
	message := task.TokenProcessingUserMessage{UserID: userID, TokenIDs: tokenIDs}
	return task.CreateTaskForTokenProcessing(ctx, message, m.taskClient)
}

func (m Manager) tryRetry(ctx context.Context, token persist.TokenIdentifiers, err error, attempts int) error {
	if err == nil || m.maxRetries <= 0 || attempts >= m.maxRetries {
		m.processRegistry.finish(ctx, token)
		return nil
	}

	_, delay, err := m.delayer.ForKey(ctx, token.String())
	if err != nil {
		m.processRegistry.finish(ctx, token)
		return err
	}

	m.processRegistry.setEnqueue(ctx, token)
	message := task.TokenProcessingTokenMessage{Token: token, Attempts: attempts + 1}
	return task.CreateTaskForTokenTokenProcessing(ctx, message, m.taskClient, delay)
}

type registry struct{ c *redis.Cache }

func (r registry) processing(ctx context.Context, token persist.TokenIdentifiers) (bool, error) {
	_, err := r.c.Get(ctx, prefixKey(token))
	return err == nil, err
}

func (r registry) finish(ctx context.Context, token persist.TokenIdentifiers) error {
	return r.c.Delete(ctx, prefixKey(token))
}

func (r registry) setEnqueue(ctx context.Context, token persist.TokenIdentifiers) error {
	return r.setManyEnqueue(ctx, []persist.TokenIdentifiers{token})
}

func (r registry) setManyEnqueue(ctx context.Context, tokens []persist.TokenIdentifiers) error {
	keyValues := make(map[string]any, len(tokens))
	for _, t := range tokens {
		keyValues[prefixKey(t)] = []byte("enqueued")
	}
	return r.c.MSetWithTTL(ctx, keyValues, time.Hour)
}

func (r registry) keepAlive(ctx context.Context, token persist.TokenIdentifiers) error {
	return r.c.Set(ctx, prefixKey(token), []byte("processing"), time.Minute)
}

func prefixKey(t persist.TokenIdentifiers) string { return "inflight:" + t.String() }