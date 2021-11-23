package server

import (
	"context"
	"net/http"
	"time"

	"cloud.google.com/go/pubsub/pstest"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/mikeydub/go-gallery/eth"
	"github.com/mikeydub/go-gallery/memstore"
	"github.com/mikeydub/go-gallery/middleware"
	"github.com/mikeydub/go-gallery/persist"
	"github.com/mikeydub/go-gallery/persist/mongodb"
	"github.com/mikeydub/go-gallery/pubsub"
	"github.com/mikeydub/go-gallery/pubsub/gcp"
	"github.com/mikeydub/go-gallery/util"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

type repositories struct {
	userRepository            persist.UserRepository
	nonceRepository           persist.NonceRepository
	loginRepository           persist.LoginAttemptRepository
	nftRepository             persist.NFTRepository
	tokenRepository           persist.TokenRepository
	collectionRepository      persist.CollectionRepository
	collectionTokenRepository persist.CollectionTokenRepository
	galleryRepository         persist.GalleryRepository
	galleryTokenRepository    persist.GalleryTokenRepository
	historyRepository         persist.OwnershipHistoryRepository
	accountRepository         persist.AccountRepository
	contractRepository        persist.ContractRepository
	backupRepository          persist.BackupRepository
	membershipRepository      persist.MembershipRepository
}

// Init initializes the server
func init() {
	router := CoreInit()

	http.Handle("/", router)
}

// CoreInit initializes core server functionality. This is abstracted
// so the test server can also utilize it
func CoreInit() *gin.Engine {
	log.Info("initializing server...")

	log.SetReportCaller(true)

	setDefaults()

	router := gin.Default()
	router.Use(middleware.HandleCORS(), middleware.ErrLogger())

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		log.Info("registering validation")
		v.RegisterValidation("short_string", shortStringValidator)
		v.RegisterValidation("medium_string", mediumStringValidator)
		v.RegisterValidation("eth_addr", ethValidator)
		v.RegisterValidation("nonce", nonceValidator)
		v.RegisterValidation("signature", signatureValidator)
		v.RegisterValidation("username", usernameValidator)
	}

	return handlersInit(router, newRepos(), newEthClient(), newIPFSShell(), newGCPPubSub())
}

func setDefaults() {
	viper.SetDefault("ENV", "local")
	viper.SetDefault("ALLOWED_ORIGINS", "http://localhost:3000")
	viper.SetDefault("JWT_SECRET", "Test-Secret")
	viper.SetDefault("JWT_TTL", 60*60*24*3)
	viper.SetDefault("PORT", 4000)
	viper.SetDefault("MONGO_URL", "mongodb://localhost:27017/")
	viper.SetDefault("IPFS_URL", "https://ipfs.io")
	viper.SetDefault("GCLOUD_TOKEN_CONTENT_BUCKET", "token-content")
	viper.SetDefault("REDIS_URL", "localhost:6379")
	viper.SetDefault("GOOGLE_APPLICATION_CREDENTIALS", "decrypted/service-key.json")
	viper.SetDefault("CONTRACT_ADDRESS", "0xe01569ca9b39e55bc7c0dfa09f05fa15cb4c7698")
	viper.SetDefault("CONTRACT_INTERACTION_URL", "https://eth-mainnet.alchemyapi.io/v2/lZc9uHY6g2ak1jnEkrOkkopylNJXvE76")
	viper.SetDefault("REQUIRE_NFTS", false)
	viper.SetDefault("ADMIN_PASS", "TEST_ADMIN_PASS")
	viper.SetDefault("MIXPANEL_TOKEN", "")
	viper.SetDefault("MIXPANEL_API_URL", "https://api.mixpanel.com/track")
	viper.SetDefault("SIGNUPS_TOPIC", "user-signup")

	viper.AutomaticEnv()

	if viper.GetString("ENV") != "local" && viper.GetString("ADMIN_PASS") == "TEST_ADMIN_PASS" {
		panic("ADMIN_PASS must be set")
	}
}

func newRepos() *repositories {

	mgoClient := newMongoClient()
	redisClients := newMemstoreClients()
	return &repositories{
		nonceRepository:           mongodb.NewNonceMongoRepository(mgoClient),
		loginRepository:           mongodb.NewLoginMongoRepository(mgoClient),
		collectionRepository:      mongodb.NewCollectionMongoRepository(mgoClient, redisClients),
		tokenRepository:           mongodb.NewTokenMongoRepository(mgoClient),
		collectionTokenRepository: mongodb.NewCollectionTokenMongoRepository(mgoClient, redisClients),
		galleryTokenRepository:    mongodb.NewGalleryTokenMongoRepository(mgoClient),
		galleryRepository:         mongodb.NewGalleryMongoRepository(mgoClient),
		historyRepository:         mongodb.NewHistoryMongoRepository(mgoClient),
		nftRepository:             mongodb.NewNFTMongoRepository(mgoClient, redisClients),
		userRepository:            mongodb.NewUserMongoRepository(mgoClient),
		accountRepository:         mongodb.NewAccountMongoRepository(mgoClient),
		contractRepository:        mongodb.NewContractMongoRepository(mgoClient),
		backupRepository:          mongodb.NewBackupMongoRepository(mgoClient),
		membershipRepository:      mongodb.NewMembershipMongoRepository(mgoClient),
	}
}

func newMongoClient() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()
	mgoURL := viper.GetString("MONGO_URL")
	if viper.GetString("ENV") != "local" {
		mongoSecretName := viper.GetString("MONGO_SECRET_NAME")
		secret, err := util.AccessSecret(context.Background(), mongoSecretName)
		if err != nil {
			panic(err)
		}
		mgoURL = string(secret)
	}
	logrus.Infof("Connecting to mongo at %s\n", mgoURL)

	mOpts := options.Client().ApplyURI(string(mgoURL))
	mOpts.SetRegistry(mongodb.CustomRegistry)
	mOpts.SetRetryWrites(true)
	mOpts.SetWriteConcern(writeconcern.New(writeconcern.WMajority()))

	mClient, err := mongo.Connect(ctx, mOpts)
	if err != nil {
		panic(err)
	}

	err = mClient.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}

	return mClient
}

func newEthClient() *eth.Client {
	client, err := ethclient.Dial(viper.GetString("CONTRACT_INTERACTION_URL"))
	if err != nil {
		panic(err)
	}
	return eth.NewEthClient(client, viper.GetString("CONTRACT_ADDRESS"))
}

func newMemstoreClients() *memstore.Clients {
	redisURL := viper.GetString("REDIS_URL")
	redisPass := viper.GetString("REDIS_PASS")
	opensea := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: redisPass,
		DB:       int(memstore.OpenseaRDB),
	})
	if err := opensea.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	unassigned := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: redisPass,
		DB:       int(memstore.CollUnassignedRDB),
	})
	if err := unassigned.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return memstore.NewMemstoreClients(opensea, unassigned)
}

func newIPFSShell() *shell.Shell {

	sh := shell.NewShell(viper.GetString("IPFS_URL"))
	sh.SetTimeout(time.Second * 15)
	return sh
}

func newGCPPubSub() pubsub.PubSub {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()
	if viper.GetString("ENV") != "local" {
		pub, err := gcp.NewGCPPubSub(ctx, viper.GetString("GOOGLE_CLOUD_PROJECT"))
		if err != nil {
			panic(err)
		}
		return pub
	}
	srv := pstest.NewServer()
	// Connect to the server without using TLS.
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	// Use the connection when creating a pubsub client.
	client, err := gcp.NewGCPPubSub(ctx, viper.GetString("GOOGLE_PROJECT_ID"), option.WithGRPCConn(conn))
	if err != nil {
		panic(err)
	}

	client.CreateTopic(ctx, viper.GetString("SIGNUPS_TOPIC"))
	return client
}
