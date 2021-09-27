package main

import (
	"github.com/IamCathal/neo/services/frontend/util"
)

func main() {
	logConfig := util.LoadLoggingConfig()
	logger := util.InitLogger(logConfig)
	logger.Info("Hello world!")
}
