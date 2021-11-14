package configuration

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/joho/godotenv"
	"github.com/neosteamfriendgraphing/common/util"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

var (
	Logger                 *zap.Logger
	ApplicationStartUpTime time.Time

	WorkerConfig datastructures.WorkerConfig

	Queue   amqp.Queue
	Channel amqp.Channel

	UsableAPIKeys datastructures.APIKeysInUse
)

func InitConfig() error {
	ApplicationStartUpTime = time.Now()
	if err := godotenv.Load(); err != nil {
		return err
	}

	InitAndSetWorkerConfig()
	logConfig, err := LoadLoggingConfig()
	if err != nil {
		log.Fatal(err)
	}

	InitAndSetLogger(logConfig)
	InitRabbitMQConnection()

	return nil
}

func InitAndSetWorkerConfig() {
	workerConfig := datastructures.WorkerConfig{}

	workerAmountFromEnv, _ := strconv.Atoi(os.Getenv("WORKER_AMOUNT"))
	workerConfig.WorkerAmount = workerAmountFromEnv

	WorkerConfig = workerConfig
}

func LoadLoggingConfig() (datastructures.LoggingFields, error) {
	logFieldsConfig := datastructures.LoggingFields{
		NodeName: os.Getenv("NODE_NAME"),
		NodeDC:   os.Getenv("NODE_DC"),
		LogPaths: []string{"stdout", os.Getenv("LOG_PATH")},
		NodeIPV4: util.GetLocalIPAddress(),
	}
	if logFieldsConfig.NodeName == "" || logFieldsConfig.NodeDC == "" ||
		logFieldsConfig.LogPaths[1] == "" || logFieldsConfig.NodeIPV4 == "" {

		return datastructures.LoggingFields{}, fmt.Errorf("one or more required environment variables are not set: %v", logFieldsConfig)
	}
	return logFieldsConfig, nil
}

func InitAndSetLogger(logFieldsConfig datastructures.LoggingFields) {
	os.OpenFile(logFieldsConfig.LogPaths[1], os.O_RDONLY|os.O_CREATE, 0666)
	c := zap.NewProductionConfig()
	c.OutputPaths = logFieldsConfig.LogPaths

	globalLogFields := make(map[string]interface{})
	globalLogFields["nodeName"] = logFieldsConfig.NodeName
	globalLogFields["nodeDC"] = logFieldsConfig.NodeDC
	globalLogFields["nodeIPV4"] = logFieldsConfig.NodeIPV4
	c.InitialFields = globalLogFields

	log, err := c.Build()
	if err != nil {
		panic(err)
	}
	Logger = log
}

func InitRabbitMQConnection() {
	if os.Getenv("RABBITMQ_PASSWORD") == "" ||
		os.Getenv("RABBITMQ_USER") == "" ||
		os.Getenv("RABBITMQ_URL") == "" {
		logMsg := "one or more env vars for connecting to rabbitMQ are not set\n"
		logMsg += fmt.Sprintf("RABBITMQ_PASSWORD is empty: %t", os.Getenv("RABBITMQ_PASSWORD") == "")
		logMsg += fmt.Sprintf("RABBITMQ_USER is empty: %t", os.Getenv("RABBITMQ_USER") == "")
		logMsg += fmt.Sprintf("RABBITMQ_URL is empty: %t", os.Getenv("RABBITMQ_URL") == "")
		Logger.Fatal(logMsg)
		log.Fatal(logMsg)
	}

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s", os.Getenv("RABBITMQ_USER"), os.Getenv("RABBITMQ_PASSWORD"), os.Getenv("RABBITMQ_URL")))
	if err != nil {
		log.Fatal(err)
	}
	// defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	err = channel.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatal(err)
	}

	Queue = queue
	Channel = *channel
	Logger.Info("started rabbitMQ connection")
}
