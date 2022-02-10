package configuration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/util"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

var (
	Logger                 *zap.Logger
	ApplicationStartUpTime time.Time
	DBClient               *mongo.Client
	InfluxDBClient         influxdb2.Client
	SQLClient              *sql.DB

	OVERWRITE_USERS_BEYOND int64
)

func InitConfig() error {
	ApplicationStartUpTime = time.Now()
	if err := godotenv.Load(); err != nil {
		return err
	}

	err := util.EnsureAllEnvVarsAreSet("MONGODB_USER",
		"MONGODB_PASSWORD", "MONGO_INSTANCE_IP", "DB_NAME",
		"USER_COLLECTION", "CRAWLING_STATS_COLLECTION",
		"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB",
		"POSTGRES_INSTANCE_IP", "SHORTEST_DISTANCE_COLLECTION")
	if err != nil {
		return err
	}

	logConfig, err := LoadLoggingConfig()
	if err != nil {
		log.Fatal(err)
	}

	InitAndSetLogger(logConfig)
	InitMongoDBConnection()
	InitAndSetInfluxClient()
	InitSQLDBConnection()

	fmt.Printf("the val is: %+v\n", OVERWRITE_USERS_BEYOND)
	return nil
}

func InitAndSetInfluxClient() {
	client := influxdb2.NewClientWithOptions(
		os.Getenv("INFLUXDB_URL"),
		os.Getenv("SYSTEM_STATS_BUCKET_TOKEN"),
		influxdb2.DefaultOptions().SetBatchSize(10))
	InfluxDBClient = client
	Logger.Info("InfluxDB client initialied successfully")
}

func LoadLoggingConfig() (common.LoggingFields, error) {
	logFieldsConfig := common.LoggingFields{
		NodeName: os.Getenv("NODE_NAME"),
		NodeDC:   os.Getenv("NODE_DC"),
		LogPaths: []string{"stdout", os.Getenv("LOG_PATH")},
		NodeIPV4: GetLocalIPAddress(),
		Service:  os.Getenv("SERVICE"),
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
	globalLogFields["service"] = logFieldsConfig.Service
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

	mongoDBConnectionURL := fmt.Sprintf("mongodb://%s:%s@%s:27017/maindb?authSource=maindb&readPreference=primary&directConnection=true", mongoDBUser, mongoDBPassword, mongoInstanceIP)

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

func InitSQLDBConnection() {
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDB := os.Getenv("POSTGRES_DB")
	postgresInstanceIP := os.Getenv("POSTGRES_INSTANCE_IP")

	connURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		postgresUser, postgresPassword, postgresInstanceIP, postgresDB)

	db, err := sql.Open("postgres", connURL)
	if err != nil {
		logMsg := fmt.Sprintf("couldn't open connection to SQL db: %+v", err)
		Logger.Fatal(logMsg)
		panic(logMsg)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		logMsg := fmt.Sprintf("couldn't ping sql db: %+v", err)
		Logger.Fatal(logMsg)
		panic(logMsg)
	}

	SQLClient = db
	Logger.Info("SQL connection initialised successfully")
}
