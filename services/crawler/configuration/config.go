package configuration

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/joho/godotenv"
	commonUtil "github.com/neosteamfriendgraphing/common/util"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

var (
	Logger                 *zap.Logger
	ApplicationStartUpTime time.Time

	WorkerConfig datastructures.WorkerConfig

	Queue           amqp.Queue
	ConsumeChannel  amqp.Channel
	AmqpChannels    []amqp.Channel
	amqlChannelLock sync.Mutex

	UsableAPIKeys datastructures.APIKeysInUse
)

func InitConfig() error {
	ApplicationStartUpTime = time.Now()
	if err := godotenv.Load(); err != nil {
		return err
	}
	commonUtil.EnsureAllEnvVarsAreSet("RABBITMQ_PASSWORD", "RABBITMQ_USER",
		"RABBITMQ_URL", "DATASTORE_INSTANCE", "WORKER_AMOUNT", "STEAM_API_KEYS",
		"KEY_SLEEP_TIME")
	logConfig, err := commonUtil.LoadLoggingConfig()
	if err != nil {
		return commonUtil.MakeErr(err)
	}
	logger := commonUtil.InitLogger(logConfig)
	Logger = logger

	InitAndSetWorkerConfig()
	newQueue, channel := InitRabbitMQConnection()
	Queue = newQueue
	ConsumeChannel = channel
	InitExtraAMQPChannels()

	return nil
}

func InitAndSetWorkerConfig() {
	workerConfig := datastructures.WorkerConfig{}

	workerAmountFromEnv, _ := strconv.Atoi(os.Getenv("WORKER_AMOUNT"))
	workerConfig.WorkerAmount = workerAmountFromEnv

	WorkerConfig = workerConfig
}

func InitRabbitMQConnection() (amqp.Queue, amqp.Channel) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s", os.Getenv("RABBITMQ_USER"), os.Getenv("RABBITMQ_PASSWORD"), os.Getenv("RABBITMQ_URL")))
	if err != nil {
		log.Fatal(commonUtil.MakeErr(err))
	}
	// defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatal(commonUtil.MakeErr(err))
	}
	// defer channel.Close()

	queue, err := channel.QueueDeclare(
		os.Getenv("RABBITMQ_QUEUE_NAME"), // name
		false,                            // durable
		false,                            // delete when unused
		false,                            // exclusive
		false,                            // no-wait
		nil,                              // arguments
	)
	if err != nil {
		log.Fatal(commonUtil.MakeErr(err))
	}
	err = channel.Qos(
		2,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatal(commonUtil.MakeErr(err))
	}

	Logger.Info("started rabbitMQ connection")
	return queue, *channel
}

func InitExtraAMQPChannels() {
	extraChannels := 5
	var waitG sync.WaitGroup
	waitG.Add(extraChannels)

	for i := 0; i < extraChannels; i++ {
		go initAndAddAMQPChannel(&waitG)
	}
	waitG.Wait()
	Logger.Sugar().Infof("initialised %d rabbitMQ channels successfully", len(AmqpChannels))
}

func initAndAddAMQPChannel(waitG *sync.WaitGroup) {
	defer waitG.Done()

	_, newChannel := InitRabbitMQConnection()
	amqlChannelLock.Lock()
	AmqpChannels = append(AmqpChannels, newChannel)
	amqlChannelLock.Unlock()
}
