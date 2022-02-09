package endpoints

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/dbmonitor"
	"github.com/gorilla/mux"
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

func gunzip(body io.ReadCloser) ([]byte, error) {
	requestBodyGzipData, err := ioutil.ReadAll(body)
	if err != nil {
		return []byte{}, err
	}
	byteBuffer := bytes.NewBuffer(requestBodyGzipData)
	var reader io.Reader
	reader, err = gzip.NewReader(byteBuffer)
	if err != nil {
		return []byte{}, err
	}

	var resBytes bytes.Buffer
	_, err = resBytes.ReadFrom(reader)
	if err != nil {
		return []byte{}, err
	}
	return resBytes.Bytes(), nil
}

func gzipData(inputbytes []byte) (bytes.Buffer, error) {
	gzippedData := bytes.Buffer{}
	gz := gzip.NewWriter(&gzippedData)
	if _, err := gz.Write(inputbytes); err != nil {
		return bytes.Buffer{}, err
	}
	if err := gz.Close(); err != nil {
		return bytes.Buffer{}, err
	}
	return gzippedData, nil
}

func wsReader(ws dbmonitor.WebsocketConn, streamType string) {
	defer ws.Ws.Close()
	ws.Ws.SetReadLimit(1024)
	ws.Ws.SetPongHandler(func(string) error {
		ws.Ws.SetReadDeadline(time.Now().Add(time.Duration(1 * time.Second)))
		return nil
	})
	for {
		_, _, err := ws.Ws.ReadMessage()
		if err != nil {
			switch streamType {
			case "newuser":
				newUserSteamWebsockets, err := dbmonitor.RemoveAWebsocketConnection(ws.ID, dbmonitor.NewUserStreamWebsockets, &dbmonitor.NewUserStreamLock)
				if err != nil {
					configuration.Logger.Fatal(err.Error())
					panic(err)
				}
				dbmonitor.SetNewUserStreamWebsocketConnections(newUserSteamWebsockets)
				configuration.Logger.Sugar().Infof("websocket %s is exiting", ws.ID)
			case "crawlingstats":
				crawlingStatWebsockets, err := dbmonitor.RemoveAWebsocketConnection(ws.ID, dbmonitor.CrawlingStatsStreamWebsockets, &dbmonitor.CrawlingStatsStreamLock)
				if err != nil {
					configuration.Logger.Fatal(err.Error())
					panic(err)
				}
				dbmonitor.SetCrawlingStatsStreamWebsocketConnections(crawlingStatWebsockets)
				configuration.Logger.Sugar().Infof("websocket %s is exiting", ws.ID)
			}
			break
		}
	}
}
