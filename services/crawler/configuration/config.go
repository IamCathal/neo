package configuration

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

var (
	Logger                 *zap.Logger
	ApplicationStartUpTime time.Time

	WorkerConfig datastructures.WorkerConfig

	rabbitMQUser string
	rabbitMQURL  string
	Queue        amqp.Queue
	Channel      amqp.Channel

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
		LogPath:  os.Getenv("LOG_PATH"),
		NodeIPV4: GetLocalIPAddress(),
	}
	if logFieldsConfig.NodeName == "" || logFieldsConfig.NodeDC == "" ||
		logFieldsConfig.LogPath == "" || logFieldsConfig.NodeIPV4 == "" {

		return datastructures.LoggingFields{}, fmt.Errorf("one or more required environment variables are not set: %v", logFieldsConfig)
	}
	return logFieldsConfig, nil
}

func InitAndSetLogger(logFieldsConfig datastructures.LoggingFields) {
	os.OpenFile(logFieldsConfig.LogPath, os.O_RDONLY|os.O_CREATE, 0666)
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"stdout", logFieldsConfig.LogPath}

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
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/", rabbitMQUser, rabbitMQUser, rabbitMQURL))
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
		"jobsQueue", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		log.Fatal(err)
	}

	Queue = queue
	Channel = *channel
}

func GetLocalIPAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	addrWithNoPort := strings.Split(conn.LocalAddr().(*net.UDPAddr).String(), ":")
	return addrWithNoPort[0]
}

func GetCurrentTimeInMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GetRequestStartTimeInTimeFormat(requestStartTimeString string) int64 {
	requestStartTime, err := strconv.ParseInt(requestStartTimeString, 10, 64)
	if err != nil {
		panic(err)
	}
	return requestStartTime
}
