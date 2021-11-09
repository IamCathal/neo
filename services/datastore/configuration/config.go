package configuration

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/neosteamfriendgraphing/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

var (
	Logger                 *zap.Logger
	ApplicationStartUpTime time.Time
	DBClient               *mongo.Client
)

func InitConfig() error {
	ApplicationStartUpTime = time.Now()
	if err := godotenv.Load(); err != nil {
		return err
	}

	err := EnsureAllServiceSpecificEnvVarsAreSet()
	if err != nil {
		return err
	}
	// err := common.EnsureAllDefaultEnvVarsAreSet()
	// if err != nil {
	// 	return err
	// }

	logConfig, err := LoadLoggingConfig()
	if err != nil {
		log.Fatal(err)
	}

	InitAndSetLogger(logConfig)
	InitMongoDBConnection()

	return nil
}

func EnsureAllServiceSpecificEnvVarsAreSet() error {
	resultString := ""
	if os.Getenv("MONGODB_USER") == "" {
		resultString += "MONGODB_USER\n"
	}
	if os.Getenv("MONGODB_PASSWORD") == "" {
		resultString += "MONGODB_PASSWORD\n"
	}
	if os.Getenv("MONGO_INSTANCE_IP") == "" {
		resultString += "MONGO_INSTANCE_IP\n"
	}
	if os.Getenv("DB_NAME") == "" {
		resultString += "DB_NAME\n"
	}
	if os.Getenv("USER_COLLECTION") == "" {
		resultString += "USER_COLLECTION\n"
	}
	if os.Getenv("CRAWLING_STATS_COLLECTION") == "" {
		resultString += "CRAWLING_STATS_COLLECTION\n"
	}

	if resultString != "" {
		return fmt.Errorf("One or more service specific env vars were not set: %s", resultString)
	}
	return nil
}

func LoadLoggingConfig() (common.LoggingFields, error) {
	logFieldsConfig := common.LoggingFields{
		NodeName: os.Getenv("NODE_NAME"),
		NodeDC:   os.Getenv("NODE_DC"),
		LogPaths: []string{"stdout", os.Getenv("LOG_PATH")},
		NodeIPV4: GetLocalIPAddress(),
	}
	if logFieldsConfig.NodeName == "" || logFieldsConfig.NodeDC == "" ||
		logFieldsConfig.LogPaths[1] == "" || logFieldsConfig.NodeIPV4 == "" {

		return common.LoggingFields{}, fmt.Errorf("one or more required environment variables are not set: %v", logFieldsConfig)
	}
	return logFieldsConfig, nil
}

func InitAndSetLogger(logFieldsConfig common.LoggingFields) {
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

func InitMongoDBConnection() {
	mongoDBUser := os.Getenv("MONGODB_USER")
	mongoDBPassword := os.Getenv("MONGODB_PASSWORD")
	mongoInstanceIP := os.Getenv("MONGO_INSTANCE_IP")

	if mongoDBUser == "" || mongoDBPassword == "" || mongoInstanceIP == "" {
		Logger.Fatal("one or more mongoDB env vars are not set")
		log.Fatal("err")
	}

	mongoDBConnectionURL := fmt.Sprintf("mongodb://%s:%s@%s:27017/testdb", mongoDBUser, mongoDBPassword, mongoInstanceIP)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoDBConnectionURL))
	if err != nil {
		Logger.Fatal(fmt.Sprintf("unable to connect to mongoDB with url '%s': %v", mongoDBConnectionURL, err))
		log.Fatal(err)
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		Logger.Fatal(fmt.Sprintf("unable to ping mongoDB: %v", err))
		log.Fatal(err)
	}

	DBClient = client
	Logger.Info("MongoDB connection initialised successfully")
}
