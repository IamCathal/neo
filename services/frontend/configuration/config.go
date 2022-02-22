package configuration

import (
	"log"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/joho/godotenv"
	"github.com/neosteamfriendgraphing/common/util"
	"go.uber.org/zap"
)

var (
	Logger                 *zap.Logger
	ApplicationStartUpTime time.Time
	InfluxDBClient         influxdb2.Client
)

func InitConfig() error {
	ApplicationStartUpTime = time.Now()
	if err := godotenv.Load(); err != nil {
		return err
	}
	if err := util.EnsureAllEnvVarsAreSet("CRAWLER_INSTANCE", "STATIC_CONTENT_DIR_NAME",
		"DATASTORE_INSTANCE"); err != nil {
		log.Fatal(err)
	}
	logConfig, err := util.LoadLoggingConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger := util.InitLogger(logConfig)
	Logger = logger
	InitAndSetInfluxClient()

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
