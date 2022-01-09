package endpoints

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
)

var (
	newUserStreamWebsockets []WebsocketConn
	newUserStreamLock       sync.Mutex
)

type WebsocketConn struct {
	Ws *websocket.Conn
	ID string
}

func GetNewUserStreamWebsocketConnections() []WebsocketConn {
	return newUserStreamWebsockets
}

func addNewUserStreamWebsocketConnection(conn WebsocketConn) {
	newUserStreamLock.Lock()
	newUserStreamWebsockets = append(newUserStreamWebsockets, conn)
	configuration.Logger.Sugar().Infof("adding websocket connection %s to addNewUserStream websocket connections", conn.ID)
	newUserStreamLock.Unlock()
}

func removeNewUserStreamWebsocketConnection(websocketID string) error {
	newUserStreamLock.Lock()
	websocketFound := false
	for i, currWebsock := range newUserStreamWebsockets {
		if currWebsock.ID == websocketID {
			websocketFound = true
			newUserStreamWebsockets[i] = newUserStreamWebsockets[len(newUserStreamWebsockets)-1]
			newUserStreamWebsockets = newUserStreamWebsockets[:len(newUserStreamWebsockets)-1]
			newUserStreamLock.Unlock()
			configuration.Logger.Sugar().Infof("removing websocket connection %s to addNewUserStream websocket connections", currWebsock.ID)
		}
	}
	if websocketFound {
		return nil
	}
	return fmt.Errorf("failed to remove non existant websocket %s from new user stream ws connection list", websocketID)
}

func (endpoints *Endpoints) NewUserStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vars["requestid"] = ksuid.New().String()

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			configuration.Logger.Sugar().Errorf("error upgrading websocket connection (handshake error): %+v", err)
			panic(err)
		}
		configuration.Logger.Sugar().Errorf("error upgrading websocket connection: %+v", err)
		panic(err)
	}

	websocketConn := WebsocketConn{
		Ws: ws,
		ID: vars["requestid"],
	}
	addNewUserStreamWebsocketConnection(websocketConn)

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
