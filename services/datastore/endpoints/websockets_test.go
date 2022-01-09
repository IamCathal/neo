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

	addNewUserStreamWebsocketConnection(newWs)

	assert.Equal(t, expectedWebsocketConnections, newUserStreamWebsockets)
}

func TestRemoveNewUserStreamWebsocketConnection(t *testing.T) {
	newUserStreamWebsockets = []WebsocketConn{}
	newWs := WebsocketConn{
		Ws: &websocket.Conn{},
		ID: ksuid.New().String(),
	}
	addNewUserStreamWebsocketConnection(newWs)

	err := removeNewUserStreamWebsocketConnection(newWs.ID)

	assert.Nil(t, err)
	assert.Equal(t, GetNewUserStreamWebsocketConnections(), []WebsocketConn{})
}

func TestRemoveNewUserStreamWebsocketConnectionWithNonExistantReturnsAnError(t *testing.T) {
	newUserStreamWebsockets = []WebsocketConn{}
	nonExistantID := "NonExistantID"
	newWs := WebsocketConn{
		Ws: &websocket.Conn{},
		ID: ksuid.New().String(),
	}
	addNewUserStreamWebsocketConnection(newWs)

	err := removeNewUserStreamWebsocketConnection(nonExistantID)

	expectedErrorString := fmt.Sprintf("failed to remove non existant websocket %s from new user stream ws connection list", nonExistantID)
	assert.EqualError(t, err, expectedErrorString)
	assert.Equal(t, newUserStreamWebsockets, GetNewUserStreamWebsocketConnections())
}
