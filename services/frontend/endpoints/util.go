package endpoints

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/IamCathal/neo/services/frontend/configuration"
	"github.com/gorilla/mux"
	"github.com/neosteamfriendgraphing/common/util"
	"go.uber.org/zap"
)

func LogBasicErr(err error, req *http.Request, statusCode int) {
	vars := mux.Vars(req)
	requestStartTime, _ := strconv.ParseInt(vars["requestStartTime"], 10, 64)
	configuration.Logger.Error(fmt.Sprintf("%v", err),
		zap.String("requestID", vars["requestID"]),
		zap.Int("status", http.StatusInternalServerError),
		zap.Int64("duration", util.GetCurrentTimeInMs()-requestStartTime),
		zap.String("path", req.URL.EscapedPath()),
	)
}

func LogBasicInfo(msg string, req *http.Request, statusCode int) {
	vars := mux.Vars(req)
	requestStartTime, _ := strconv.ParseInt(vars["requestStartTime"], 10, 64)
	configuration.Logger.Info(msg,
		zap.String("requestID", vars["requestID"]),
		zap.Int("status", statusCode),
		zap.Int64("duration", util.GetCurrentTimeInMs()-requestStartTime),
		zap.String("path", req.URL.EscapedPath()),
	)
}
