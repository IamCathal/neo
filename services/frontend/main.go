package main

import (
	"github.com/neo/frontend/util"
)

func main() {
	logConfig := util.LoadLoggingConfig()
	logger := util.InitLogger(logConfig)
	logger.Info("Hello world!")
}
