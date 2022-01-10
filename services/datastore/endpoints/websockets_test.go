package endpoints

import (
	"fmt"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
)

func TestGetNewUserStreamWebsocketConnections(t *testing.T) {
	newUserStreamWebsockets = []WebsocketConn{}
	expectedWebsocketConnections := []WebsocketConn{
		{
			Ws: &websocket.Conn{},
			ID: ksuid.New().String(),
		},
	}
	newUserStreamWebsockets = expectedWebsocketConnections

	actualWebsocketConnections := GetNewUserStreamWebsocketConnections()

	assert.Equal(t, expectedWebsocketConnections, actualWebsocketConnections)
}

func TestAddNewUserStreamWebsocketConnection(t *testing.T) {
	newUserStreamWebsockets = []WebsocketConn{}
	newWs := WebsocketConn{
		Ws: &websocket.Conn{},
		ID: ksuid.New().String(),
	}
	expectedWebsocketConnections := []WebsocketConn{
		newWs,
	}

	newUserStreamWebsockets = addNewStreamWebsocketConnection(newWs, newUserStreamWebsockets, &newUserStreamLock)

	assert.Equal(t, expectedWebsocketConnections, newUserStreamWebsockets)
}

func TestRemoveNewUserStreamWebsocketConnection(t *testing.T) {
	newUserStreamWebsockets = []WebsocketConn{}
	newWs := WebsocketConn{
		Ws: &websocket.Conn{},
		ID: ksuid.New().String(),
	}
	newUserStreamWebsockets = addNewStreamWebsocketConnection(newWs, newUserStreamWebsockets, &newUserStreamLock)

	newUserStreamWebsockets, err := removeAWebsocketConnection(newWs.ID, newUserStreamWebsockets, &newUserStreamLock)

	assert.Nil(t, err)
	assert.Equal(t, newUserStreamWebsockets, []WebsocketConn{})
}

func TestRemoveNewUserStreamWebsocketConnectionWithNonExistantReturnsAnError(t *testing.T) {
	newUserStreamWebsockets = []WebsocketConn{}
	nonExistantID := "NonExistantID"
	newWs := WebsocketConn{
		Ws: &websocket.Conn{},
		ID: ksuid.New().String(),
	}
	newUserStreamWebsockets = addNewStreamWebsocketConnection(newWs, newUserStreamWebsockets, &newUserStreamLock)

	_, err := removeAWebsocketConnection(nonExistantID, newUserStreamWebsockets, &newUserStreamLock)

	expectedErrorString := fmt.Sprintf("failed to remove non existant websocket %s from ws connection list", nonExistantID)
	assert.EqualError(t, err, expectedErrorString)
}
