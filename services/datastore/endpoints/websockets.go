package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/dbmonitor"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/neosteamfriendgraphing/common/util"
	"github.com/segmentio/ksuid"
)

func (endpoints *Endpoints) NewUserStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vars["requestid"] = ksuid.New().String()

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			configuration.Logger.Sugar().Errorf("error upgrading websocket connection (handshake error): %+v", err)
			util.SendBasicInvalidResponse(w, r, "unable to upgrade websocket", vars, http.StatusBadRequest)
			return
		}
		configuration.Logger.Sugar().Errorf("error upgrading websocket connection: %+v", err)
		util.SendBasicInvalidResponse(w, r, "unable to upgrade websocket", vars, http.StatusBadRequest)
		return
	}

	websocketConn := dbmonitor.WebsocketConn{
		Ws: ws,
		ID: vars["requestid"],
	}
	websocketConnections := dbmonitor.AddNewStreamWebsocketConnection(websocketConn, dbmonitor.NewUserStreamWebsockets, &dbmonitor.NewUserStreamLock)
	dbmonitor.SetNewUserStreamWebsocketConnections(websocketConnections)

	// Write the 8 most recent user events to have some content
	// visible on page load
	for _, event := range dbmonitor.LastEightUserEvents {
		jsonEvent, err := json.Marshal(event)
		if err != nil {
			configuration.Logger.Sugar().Errorf("failed to marhsal recent user event %+v: %+v", event, err)
			panic(err)
		}
		err = ws.WriteMessage(1, jsonEvent)
		if err != nil {
			configuration.Logger.Sugar().Errorf("error writing to websocket connection: %+v", err)
			panic(err)
		}
	}
	// go writer(ws, vars["requestid"])
	wsReader(websocketConn, "newuser")
}

func (endpoints *Endpoints) CrawlingStatsUpdateStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vars["requestid"] = ksuid.New().String()

	_, err := ksuid.Parse(vars["crawlid"])
	if err != nil {
		util.SendBasicInvalidResponse(w, r, "invalid input", vars, http.StatusBadRequest)
		return
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			configuration.Logger.Sugar().Errorf("error upgrading websocket connection (handshake error): %+v", err)
			util.SendBasicInvalidResponse(w, r, "unable to upgrade websocket", vars, http.StatusBadRequest)
			return
		}
		configuration.Logger.Sugar().Errorf("error upgrading websocket connection: %+v", err)
		util.SendBasicInvalidResponse(w, r, "unable to upgrade websocket", vars, http.StatusBadRequest)
		return
	}

	websocketConn := dbmonitor.WebsocketConn{
		Ws:      ws,
		ID:      vars["requestid"],
		MatchOn: vars["crawlid"],
	}
	websocketConnections := dbmonitor.AddNewStreamWebsocketConnection(websocketConn, dbmonitor.CrawlingStatsStreamWebsockets, &dbmonitor.CrawlingStatsStreamLock)
	dbmonitor.SetCrawlingStatsStreamWebsocketConnections(websocketConnections)

	wsReader(websocketConn, "crawlingstats")
}
