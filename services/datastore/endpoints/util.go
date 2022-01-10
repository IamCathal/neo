package endpoints

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func LogBasicErr(err error, req *http.Request, statusCode int) {
	vars := mux.Vars(req)
	requestStartTime, _ := strconv.ParseInt(vars["requestStartTime"], 10, 64)
	configuration.Logger.Error(fmt.Sprintf("%v", err),
		zap.String("requestID", vars["requestID"]),
		zap.Int("status", http.StatusInternalServerError),
		zap.Int64("duration", configuration.GetCurrentTimeInMs()-requestStartTime),
		zap.String("path", req.URL.EscapedPath()),
	)
}

func LogBasicInfo(msg string, req *http.Request, statusCode int) {
	vars := mux.Vars(req)
	requestStartTime, _ := strconv.ParseInt(vars["requestStartTime"], 10, 64)
	configuration.Logger.Info(msg,
		zap.String("requestID", vars["requestID"]),
		zap.Int("status", statusCode),
		zap.Int64("duration", configuration.GetCurrentTimeInMs()-requestStartTime),
		zap.String("path", req.URL.EscapedPath()),
	)
}

func wsReader(ws *websocket.Conn, requestID string) {
	defer ws.Close()
	ws.SetReadLimit(1024)
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(time.Duration(1 * time.Second)))
		return nil
	})
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			newUserSteamWebsockets, err := removeAWebsocketConnection(requestID, newUserStreamWebsockets, &newUserStreamLock)
			if err != nil {
				configuration.Logger.Fatal(err.Error())
				panic(err)
			}
			SetNewUserStreamWebsocketConnections(newUserSteamWebsockets)
			break
		}
	}
	configuration.Logger.Sugar().Infof("websocket %s is exiting", requestID)
}
