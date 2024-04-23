package main

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	registry "github.com/confluentinc/confluent-kafka-go/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/avro"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mikeydub/go-gallery/db/gen/mirrordb"
	"github.com/mikeydub/go-gallery/env"
	"github.com/mikeydub/go-gallery/kafka-streamer/schema/ethereum"
	"github.com/mikeydub/go-gallery/service/logger"
	"github.com/mikeydub/go-gallery/service/persist"
	"github.com/mikeydub/go-gallery/service/persist/postgres"
	"github.com/mikeydub/go-gallery/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"strings"
	"time"
)

type streamerConfig struct {
	Topic   string
	Batcher batcher
}

func main() {
	setDefaults()
	logger.InitWithGCPDefaults()

	ctx := context.Background()

	// Health endpoint
	router := gin.Default()
	router.GET("/health", util.HealthCheckHandler())

	pgx := postgres.NewPgxClient()
	defer pgx.Close()

	go runStreamer(ctx, pgx)

	err := router.Run(":3000")
	if err != nil {
		err = fmt.Errorf("error running router: %w", err)
		panic(err)
	}
}

func newEthereumOwnerConfig(deserializer *avro.GenericDeserializer, queries *mirrordb.Queries) *streamerConfig {
	parseF := func(ctx context.Context, message *kafka.Message) (mirrordb.ProcessEthereumOwnerEntryParams, error) {
		return parseOwnerMessage(ctx, deserializer, message)
	}

	submitF := func(ctx context.Context, entries []mirrordb.ProcessEthereumOwnerEntryParams) error {
		return submitOwnerBatch(ctx, queries.ProcessEthereumOwnerEntry, entries)
	}

	return &streamerConfig{
		Topic:   "ethereum.owner.v4",
		Batcher: newMessageBatcher(250, time.Second, parseF, submitF),
	}
}

func newBaseOwnerConfig(deserializer *avro.GenericDeserializer, queries *mirrordb.Queries) *streamerConfig {
	parseF := func(ctx context.Context, message *kafka.Message) (mirrordb.ProcessBaseOwnerEntryParams, error) {
		ethereumEntry, err := parseOwnerMessage(ctx, deserializer, message)
		if err != nil {
			return mirrordb.ProcessBaseOwnerEntryParams{}, err
		}

		// All EVM chains (Ethereum, Base, Zora) have the same database and query structure, so we can cast between their parameters
		return mirrordb.ProcessBaseOwnerEntryParams(ethereumEntry), nil
	}

	submitF := func(ctx context.Context, entries []mirrordb.ProcessBaseOwnerEntryParams) error {
		return submitOwnerBatch(ctx, queries.ProcessBaseOwnerEntry, entries)
	}

	return &streamerConfig{
		Topic:   "base.owner.v4",
		Batcher: newMessageBatcher(250, time.Second, parseF, submitF),
	}
}

func newZoraOwnerConfig(deserializer *avro.GenericDeserializer, queries *mirrordb.Queries) *streamerConfig {
	parseF := func(ctx context.Context, message *kafka.Message) (mirrordb.ProcessZoraOwnerEntryParams, error) {
		ethereumEntry, err := parseOwnerMessage(ctx, deserializer, message)
		if err != nil {
			return mirrordb.ProcessZoraOwnerEntryParams{}, err
		}

		// All EVM chains (Ethereum, Base, Zora) have the same database and query structure, so we can cast between their parameters
		return mirrordb.ProcessZoraOwnerEntryParams(ethereumEntry), nil
	}

	submitF := func(ctx context.Context, entries []mirrordb.ProcessZoraOwnerEntryParams) error {
		return submitOwnerBatch(ctx, queries.ProcessZoraOwnerEntry, entries)
	}

	return &streamerConfig{
		Topic:   "zora.owner.v4",
		Batcher: newMessageBatcher(250, time.Second, parseF, submitF),
	}
}

func runStreamer(ctx context.Context, pgx *pgxpool.Pool) {
	deserializer, err := newDeserializerFromRegistry()
	if err != nil {
		panic(fmt.Errorf("failed to create Avro deserializer: %w", err))
	}

	queries := mirrordb.New(pgx)

	// Creating multiple configs for each topic allows them to process separate partitions in parallel
	configs := []*streamerConfig{
		newEthereumOwnerConfig(deserializer, queries),
		newEthereumOwnerConfig(deserializer, queries),
		newBaseOwnerConfig(deserializer, queries),
		newBaseOwnerConfig(deserializer, queries),
		newZoraOwnerConfig(deserializer, queries),
		newZoraOwnerConfig(deserializer, queries),
	}

	// If any topic errors more than 10 times in 10 minutes, panic and restart the whole service
	maxRetries := 10
	retryResetInterval := 10 * time.Minute

	for _, config := range configs {
		go func(config *streamerConfig) {
			retriesRemaining := maxRetries
			retriesResetAt := time.Now().Add(retryResetInterval)
			for {
				err := streamTopic(ctx, config)
				if err != nil {
					if time.Now().After(retriesResetAt) {
						retriesRemaining = maxRetries
						retriesResetAt = time.Now().Add(retryResetInterval)
					}

					if retriesRemaining == 0 {
						panic(fmt.Errorf("streaming topic '%s' failed with error: %w", config.Topic, err))
					}

					retriesRemaining--
					config.Batcher.Reset()
					logger.For(ctx).Warnf("streaming topic '%s' failed with error: %v, restarting...", config.Topic, err)
				}
			}
		}(config)
	}
}

func newDeserializerFromRegistry() (*avro.GenericDeserializer, error) {
	registryURL := env.GetString("SIMPLEHASH_REGISTRY_URL")
	registryAPIKey := env.GetString("SIMPLEHASH_REGISTRY_API_KEY")
	registrySecretKey := env.GetString("SIMPLEHASH_REGISTRY_SECRET_KEY")

	registryClient, err := registry.NewClient(registry.NewConfigWithAuthentication(registryURL, registryAPIKey, registrySecretKey))

	deserializer, err := avro.NewGenericDeserializer(registryClient, serde.ValueSerde, avro.NewDeserializerConfig())
	if err != nil {
		return nil, err
	}

	return deserializer, nil
}

type messageStats struct {
	interval       time.Duration
	nextReportTime time.Time
	inserts        int64
	updates        int64
	deletes        int64
}

func newMessageStats(reportingInterval time.Duration) *messageStats {
	return &messageStats{
		interval:       reportingInterval,
		nextReportTime: time.Now().Add(reportingInterval),
	}
}

func (m *messageStats) Update(ctx context.Context, msg *kafka.Message) {
	actionType, err := getActionType(msg)
	if err != nil {
		logger.For(ctx).Errorf("failed to get action type for message: %v", msg)
		return
	}

	switch actionType {
	case "insert":
		m.inserts++
	case "update":
		m.updates++
	case "delete":
		m.deletes++
	default:
		logger.For(ctx).Errorf("invalid action type %s for message %v", actionType, msg)
	}

	if time.Now().After(m.nextReportTime) {
		logger.For(ctx).Infof("processed %d messages in the last %s (%d inserts, %d updates, and %d deletes)", m.inserts+m.updates+m.deletes, m.interval, m.inserts, m.updates, m.deletes)
		m.inserts = 0
		m.updates = 0
		m.deletes = 0
		m.nextReportTime = time.Now().Add(m.interval)
	}
}

type batcher interface {
	Reset()
	Add(context.Context, *kafka.Message) error
	IsReady() bool
	Submit(context.Context, *kafka.Consumer) error
}

type messageBatcher[T any] struct {
	maxSize         int
	timeoutDuration time.Duration
	parseF          func(context.Context, *kafka.Message) (T, error)
	submitF         func(context.Context, []T) error

	entries     []T
	nextTimeout time.Time
}

func newMessageBatcher[T any](maxSize int, timeout time.Duration, parseF func(context.Context, *kafka.Message) (T, error), submitF func(context.Context, []T) error) *messageBatcher[T] {
	return &messageBatcher[T]{
		maxSize:         maxSize,
		timeoutDuration: timeout,
		parseF:          parseF,
		submitF:         submitF,
	}
}

func (b *messageBatcher[T]) Reset() {
	b.entries = []T{}
	b.nextTimeout = time.Time{}
}

func (b *messageBatcher[T]) Add(ctx context.Context, msg *kafka.Message) error {
	t, err := b.parseF(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	b.entries = append(b.entries, t)
	b.nextTimeout = time.Now().Add(b.timeoutDuration)
	return nil
}

func (b *messageBatcher[T]) IsReady() bool {
	if len(b.entries) == 0 {
		return false
	}

	return len(b.entries) >= b.maxSize || time.Now().After(b.nextTimeout)
}

func (b *messageBatcher[T]) Submit(ctx context.Context, c *kafka.Consumer) error {
	if len(b.entries) == 0 {
		return nil
	}

	err := b.submitF(ctx, b.entries)
	if err != nil {
		return fmt.Errorf("failed to submit batch: %w", err)
	}

	_, err = c.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit offsets: %w", err)
	}

	b.entries = []T{}
	return nil
}

func streamTopic(ctx context.Context, config *streamerConfig) error {
	ctx = logger.NewContextWithFields(ctx, logrus.Fields{"topic": config.Topic})

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  env.GetString("SIMPLEHASH_BOOTSTRAP_SERVERS"),
		"group.id":           env.GetString("SIMPLEHASH_GROUP_ID"),
		"sasl.username":      env.GetString("SIMPLEHASH_API_KEY"),
		"sasl.password":      env.GetString("SIMPLEHASH_API_SECRET"),
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
		"security.protocol":  "SASL_SSL",
		"sasl.mechanisms":    "PLAIN",
	})

	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	defer c.Close()

	batch := config.Batcher
	var rebalanceErr error
	rebalanceErrPtr := &rebalanceErr

	rebalanceCb := func(c *kafka.Consumer, event kafka.Event) error {
		switch e := event.(type) {
		case kafka.AssignedPartitions:
			err := c.Assign(e.Partitions)
			if err != nil {
				err = fmt.Errorf("failed to assign partitions: %w", err)
				*rebalanceErrPtr = err
				return err
			}
		case kafka.RevokedPartitions:
			err := batch.Submit(ctx, c)
			if err != nil {
				*rebalanceErrPtr = err
				return err
			}
			err = c.Unassign()
			if err != nil {
				err = fmt.Errorf("failed to unassign partitions: %w", err)
				*rebalanceErrPtr = err
				return err
			}
		}
		return nil
	}

	err = c.SubscribeTopics([]string{config.Topic}, rebalanceCb)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", config.Topic, err)
	}

	stats := newMessageStats(time.Hour)
	readTimeout := 100 * time.Millisecond

	for {
		msg, err := c.ReadMessage(readTimeout)

		// rebalanceCb is triggered by ReadMessage, so we want to check immediately after ReadMessage to see if
		// the rebalance failed and we need to restart
		if rebalanceErr != nil {
			return fmt.Errorf("rebalance failed: %w", rebalanceErr)
		}

		if err != nil {
			if kafkaErr, ok := util.ErrorAs[kafka.Error](err); ok {
				if kafkaErr.Code() == kafka.ErrTimedOut {
					if batch.IsReady() {
						err := batch.Submit(ctx, c)
						if err != nil {
							return fmt.Errorf("failed to submit batch: %w", err)
						}
					}
					continue
				}
			}

			return fmt.Errorf("error reading message: %w", err)
		}

		err = batch.Add(ctx, msg)
		if err != nil {
			return fmt.Errorf("failed to add message to batch: %w", err)
		}

		if batch.IsReady() {
			err := batch.Submit(ctx, c)
			if err != nil {
				return fmt.Errorf("failed to submit batch: %w", err)
			}
		}

		stats.Update(ctx, msg)
	}
}

func setDefaults() {
	viper.SetDefault("ENV", "local")
	viper.SetDefault("POSTGRES_HOST", "0.0.0.0")
	viper.SetDefault("POSTGRES_PORT", 5432)
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "")
	viper.SetDefault("POSTGRES_DB", "postgres")
	viper.SetDefault("SENTRY_DSN", "")
	viper.SetDefault("GAE_VERSION", "")
	viper.SetDefault("SENTRY_TRACES_SAMPLE_RATE", 0.2)

	viper.AutomaticEnv()

	if env.GetString("ENV") != "local" {
		logger.For(nil).Info("running in non-local environment, skipping environment configuration")
	} else {
		fi := "local"
		if len(os.Args) > 1 {
			fi = os.Args[1]
		}
		envFile := util.ResolveEnvFile("kafka-streamer", fi)
		util.LoadEncryptedEnvFile(envFile)
	}

	if env.GetString("ENV") != "local" {
		util.VarNotSetTo("SENTRY_DSN", "")
	}
}

func parseNumeric(s *string) (pgtype.Numeric, error) {
	if s == nil {
		return pgtype.Numeric{Status: pgtype.Null}, nil
	}

	n := pgtype.Numeric{}
	err := n.Set(*s)
	if err != nil {
		err = fmt.Errorf("failed to parse numeric '%s': %w", *s, err)
		return n, err
	}

	return n, nil
}

func parseNftID(nftID string) (contractAddress persist.Address, tokenID pgtype.Numeric, err error) {
	// NftID is in the format: chain.contract_address.token_id
	parts := strings.Split(nftID, ".")
	if len(parts) != 3 {
		return "", pgtype.Numeric{}, fmt.Errorf("invalid nft_id: %s", nftID)
	}

	id, err := parseNumeric(&parts[2])
	if err != nil {
		return "", pgtype.Numeric{}, fmt.Errorf("failed to parse tokenID: %w", err)
	}

	// TODO: Use the chain to map to one of our chains and normalize the address accordingly.
	// For now, just assume Ethereum and convert to lowercase.
	return persist.Address(strings.ToLower(parts[1])), id, nil
}

func parseTimestamp(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}

	ts, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp %s: %w", *s, err)
	}

	return &ts, nil
}

func parseOwnerMessage(ctx context.Context, deserializer *avro.GenericDeserializer, msg *kafka.Message) (mirrordb.ProcessEthereumOwnerEntryParams, error) {
	key := string(msg.Key)

	owner := ethereum.Owner{}
	err := deserializer.DeserializeInto(*msg.TopicPartition.Topic, msg.Value, &owner)
	if err != nil {
		return mirrordb.ProcessEthereumOwnerEntryParams{}, fmt.Errorf("failed to deserialize owner message with key %s: %w", key, err)
	}

	actionType, err := getActionType(msg)
	if err != nil {
		err = fmt.Errorf("failed to get action type for msg: %v", msg)
		return mirrordb.ProcessEthereumOwnerEntryParams{}, err
	}

	contractAddress, tokenID, err := parseNftID(owner.Nft_id)
	if err != nil {
		return mirrordb.ProcessEthereumOwnerEntryParams{}, fmt.Errorf("error parsing NftID: %w", err)
	}

	// TODO: Same as above, we'll want to normalize the address based on the chain at some point
	walletAddress := strings.ToLower(owner.Owner_address)

	quantity, err := parseNumeric(owner.Quantity)
	if err != nil {
		return mirrordb.ProcessEthereumOwnerEntryParams{}, err
	}

	firstAcquiredDate, err := parseTimestamp(owner.First_acquired_date)
	if err != nil {
		err = fmt.Errorf("failed to parse First_acquired_date: %w", err)
		return mirrordb.ProcessEthereumOwnerEntryParams{}, err
	}

	lastAcquiredDate, err := parseTimestamp(owner.Last_acquired_date)
	if err != nil {
		err = fmt.Errorf("failed to parse Last_acquired_date: %w", err)
		return mirrordb.ProcessEthereumOwnerEntryParams{}, err
	}

	var params mirrordb.ProcessEthereumOwnerEntryParams

	if actionType == "delete" {
		params = mirrordb.ProcessEthereumOwnerEntryParams{
			ShouldDelete:       true,
			SimplehashKafkaKey: key,

			// pgtype.Numeric defaults to 'undefined' instead of 'null', so we actually need to set these explicitly
			// or else we'll get an error when pgx tries to encode them
			TokenID:  pgtype.Numeric{Status: pgtype.Null},
			Quantity: pgtype.Numeric{Status: pgtype.Null},
		}
	} else {
		if actionType == "insert" || actionType == "update" {
			params = mirrordb.ProcessEthereumOwnerEntryParams{
				ShouldUpsert:             true,
				SimplehashKafkaKey:       key,
				SimplehashNftID:          &owner.Nft_id,
				KafkaOffset:              util.ToPointer(int64(msg.TopicPartition.Offset)),
				KafkaPartition:           util.ToPointer(msg.TopicPartition.Partition),
				KafkaTimestamp:           util.ToPointer(msg.Timestamp),
				ContractAddress:          &contractAddress,
				TokenID:                  tokenID,
				OwnerAddress:             util.ToPointer(persist.Address(walletAddress)),
				Quantity:                 quantity,
				CollectionID:             owner.Collection_id,
				FirstAcquiredDate:        firstAcquiredDate,
				LastAcquiredDate:         lastAcquiredDate,
				FirstAcquiredTransaction: owner.First_acquired_transaction,
				LastAcquiredTransaction:  owner.Last_acquired_transaction,
				MintedToThisWallet:       owner.Minted_to_this_wallet,
				AirdroppedToThisWallet:   owner.Airdropped_to_this_wallet,
				SoldToThisWallet:         owner.Sold_to_this_wallet,
			}
		} else {
			return params, fmt.Errorf("invalid action type: %s", actionType)
		}
	}

	return params, nil
}

type queryBatchExecuter interface {
	Exec(func(i int, e error))
	Close() error
}

func submitOwnerBatch[TBatch queryBatchExecuter, TEntries any](ctx context.Context, queryF func(context.Context, []TEntries) TBatch, entries []TEntries) error {
	b := queryF(ctx, entries)
	defer b.Close()

	var err error
	b.Exec(func(i int, e error) {
		if e != nil {
			err = fmt.Errorf("failed to process owner entry: %w", e)
		}
	})

	return err
}

func getActionType(msg *kafka.Message) (string, error) {
	for _, h := range msg.Headers {
		if h.Key == "modType" {
			return strings.ToLower(string(h.Value)), nil
		}
	}

	return "", fmt.Errorf("no action type found in message headers")
}