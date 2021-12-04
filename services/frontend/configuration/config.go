package configuration

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/neosteamfriendgraphing/common/util"
	"go.uber.org/zap"
)

var (
	Logger                 *zap.Logger
	ApplicationStartUpTime time.Time
)

func InitConfig() error {
	ApplicationStartUpTime = time.Now()
	if err := godotenv.Load(); err != nil {
		return err
	}
	if err := util.EnsureAllEnvVarsAreSet("CRAWLER_INSTANCE", "STATIC_CONTENT_DIR_NAME",
		"DATASTORE_URL"); err != nil {
		log.Fatal(err)
	}
	logConfig, err := util.LoadLoggingConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger := util.InitLogger(logConfig)
	Logger = logger

	return nil
}
