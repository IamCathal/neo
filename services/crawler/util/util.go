package util

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	commonUtil "github.com/neosteamfriendgraphing/common/util"
	"go.uber.org/zap"
)

func LogBasicInfo(msg string, req *http.Request, statusCode int) {
	vars := mux.Vars(req)
	requestStartTime, _ := strconv.ParseInt(vars["requestStartTime"], 10, 64)
	configuration.Logger.Info(msg,
		zap.String("requestID", vars["requestID"]),
		zap.Int("status", http.StatusInternalServerError),
		zap.Int64("duration", commonUtil.GetCurrentTimeInMs()-requestStartTime),
		zap.String("path", req.URL.EscapedPath()),
	)
}

func LogBasicErr(err error, req *http.Request, statusCode int) {
	vars := mux.Vars(req)
	requestStartTime, _ := strconv.ParseInt(vars["requestStartTime"], 10, 64)
	configuration.Logger.Error(fmt.Sprintf("%v", err),
		zap.String("requestID", vars["requestID"]),
		zap.Int("status", http.StatusInternalServerError),
		zap.Int64("duration", commonUtil.GetCurrentTimeInMs()-requestStartTime),
		zap.String("path", req.URL.EscapedPath()),
	)
}

func LogBasicFatal(err error, req *http.Request, statusCode int) {
	vars := mux.Vars(req)
	requestStartTime, _ := strconv.ParseInt(vars["requestStartTime"], 10, 64)
	configuration.Logger.Panic(fmt.Sprintf("%v", err),
		zap.String("requestID", vars["requestID"]),
		zap.Int("status", http.StatusInternalServerError),
		zap.Int64("duration", commonUtil.GetCurrentTimeInMs()-requestStartTime),
		zap.String("path", req.URL.EscapedPath()),
	)
}

func JobIsNotLevelOneAndNotMax(job datastructures.Job) bool {
	return (job.MaxLevel != 1) || (job.CurrentLevel < job.MaxLevel)
}
