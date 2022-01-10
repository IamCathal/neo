package endpoints

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/neosteamfriendgraphing/common/util"
	"github.com/segmentio/ksuid"
)

var (
	newUserStreamWebsockets []WebsocketConn
	newUserStreamLock       sync.Mutex

	crawlingStatsStreamWebsockets []WebsocketConn
	crawlingStatsStreamLock       sync.Mutex
)

type WebsocketConn struct {
	Ws      *websocket.Conn
	ID      string
	MatchOn string
}

func GetNewUserStreamWebsocketConnections() []WebsocketConn {
	return newUserStreamWebsockets
}
func SetNewUserStreamWebsocketConnections(connections []WebsocketConn) {
	newUserStreamWebsockets = connections
}

func GetCrawlingStatsStreamWebsocketConnections() []WebsocketConn {
	return crawlingStatsStreamWebsockets
}
func SetCrawlingStatsStreamWebsocketConnections(connections []WebsocketConn) {
	crawlingStatsStreamWebsockets = connections
}

func addNewStreamWebsocketConnection(conn WebsocketConn, connections []WebsocketConn, lock *sync.Mutex) []WebsocketConn {
	lock.Lock()
	connections = append(connections, conn)
	configuration.Logger.Sugar().Infof("adding websocket connection %+v to websocket connections", conn)
	lock.Unlock()
	return connections
}

func removeAWebsocketConnection(websocketID string, connections []WebsocketConn, lock *sync.Mutex) ([]WebsocketConn, error) {
	lock.Lock()
	websocketFound := false
	for i, currWebsock := range connections {
		if currWebsock.ID == websocketID {
			websocketFound = true
			connections[i] = connections[len(connections)-1]
			connections = connections[:len(connections)-1]
			lock.Unlock()
			configuration.Logger.Sugar().Infof("removing websocket connection %+v from websocket connections", currWebsock)
		}
	}
	if websocketFound {
		return connections, nil
	}
	return []WebsocketConn{}, fmt.Errorf("failed to remove non existant websocket %s from ws connection list", websocketID)
}

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

	websocketConn := WebsocketConn{
		Ws: ws,
		ID: vars["requestid"],
	}
	newUserStreamWebsockets = addNewStreamWebsocketConnection(websocketConn, newUserStreamWebsockets, &newUserStreamLock)

	err = ws.WriteMessage(1, []byte("HELLO WORLD"))
	if err != nil {
		configuration.Logger.Sugar().Errorf("error writing to websocket connection: %+v", err)
		panic(err)
	}
	time.Sleep(600 * time.Millisecond)
	err = ws.WriteMessage(1, []byte("HELLO EEEE"))
	if err != nil {
		configuration.Logger.Sugar().Errorf("error writing to websocket connection: %+v", err)
		panic(err)
	}

	// go writer(ws, vars["requestid"])
	wsReader(ws, vars["requestid"])
}
