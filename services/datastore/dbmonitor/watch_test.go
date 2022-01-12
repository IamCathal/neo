package dbmonitor

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/datastructures"
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	configuration.Logger = logger

	code := m.Run()

	os.Exit(code)
}

func TestGetNewUserStreamWebsocketConnections(t *testing.T) {
	NewUserStreamWebsockets = []WebsocketConn{}
	expectedWebsocketConnections := []WebsocketConn{
		{
			Ws: &websocket.Conn{},
			ID: ksuid.New().String(),
		},
	}
	NewUserStreamWebsockets = expectedWebsocketConnections

	actualWebsocketConnections := GetNewUserStreamWebsocketConnections()

	assert.Equal(t, expectedWebsocketConnections, actualWebsocketConnections)
}

func TestAddNewUserStreamWebsocketConnection(t *testing.T) {
	NewUserStreamWebsockets = []WebsocketConn{}
	newWs := WebsocketConn{
		Ws: &websocket.Conn{},
		ID: ksuid.New().String(),
	}
	expectedWebsocketConnections := []WebsocketConn{
		newWs,
	}

	NewUserStreamWebsockets = AddNewStreamWebsocketConnection(newWs, NewUserStreamWebsockets, &NewUserStreamLock)

	assert.Equal(t, expectedWebsocketConnections, NewUserStreamWebsockets)
}

func TestRemoveNewUserStreamWebsocketConnection(t *testing.T) {
	NewUserStreamWebsockets = []WebsocketConn{}
	newWs := WebsocketConn{
		Ws: &websocket.Conn{},
		ID: ksuid.New().String(),
	}
	NewUserStreamWebsockets = AddNewStreamWebsocketConnection(newWs, NewUserStreamWebsockets, &NewUserStreamLock)

	newUserStreamWebsockets, err := RemoveAWebsocketConnection(newWs.ID, NewUserStreamWebsockets, &NewUserStreamLock)

	assert.Nil(t, err)
	assert.Equal(t, newUserStreamWebsockets, []WebsocketConn{})
}

func TestRemoveNewUserStreamWebsocketConnectionWithNonExistantReturnsAnError(t *testing.T) {
	NewUserStreamWebsockets = []WebsocketConn{}
	nonExistantID := "NonExistantID"
	newWs := WebsocketConn{
		Ws: &websocket.Conn{},
		ID: ksuid.New().String(),
	}
	NewUserStreamWebsockets = AddNewStreamWebsocketConnection(newWs, NewUserStreamWebsockets, &NewUserStreamLock)

	_, err := RemoveAWebsocketConnection(nonExistantID, NewUserStreamWebsockets, &NewUserStreamLock)

	expectedErrorString := fmt.Sprintf("failed to remove non existant websocket %s from ws connection list", nonExistantID)
	assert.EqualError(t, err, expectedErrorString)
}

func TestAddUserEventToMostRecent(t *testing.T) {
	firstUserEvent := datastructures.AddUserEvent{
		SteamID:   "12345",
		CrawlTime: time.Now().Unix(),
	}
	secondUserEvent := datastructures.AddUserEvent{
		SteamID:   "67890",
		CrawlTime: time.Now().Unix(),
	}
	expectedUserEvents := []datastructures.AddUserEvent{
		secondUserEvent,
		firstUserEvent,
	}

	addUserEventToMostRecent(firstUserEvent)
	addUserEventToMostRecent(secondUserEvent)

	assert.Equal(t, expectedUserEvents, LastEightUserEvents)
}

func TestAddUserEventToMostRecentOnlyHoldsTheEightMostRecentEvents(t *testing.T) {
	firstUserEvent := datastructures.AddUserEvent{
		SteamID:   "12345",
		CrawlTime: time.Now().Unix(),
	}

	LastEightUserEvents = []datastructures.AddUserEvent{
		{
			CrawlTime: time.Now().Unix(),
		},
		{
			CrawlTime: time.Now().Unix(),
		},
		{
			CrawlTime: time.Now().Unix(),
		},
		{
			CrawlTime: time.Now().Unix(),
		},
		{
			CrawlTime: time.Now().Unix(),
		},
		{
			CrawlTime: time.Now().Unix(),
		},
		{
			CrawlTime: time.Now().Unix(),
		},
		{
			CrawlTime: time.Now().Unix(),
		},
		{
			CrawlTime: time.Now().Unix(),
		},
	}
	addUserEventToMostRecent(firstUserEvent)

	assert.Len(t, LastEightUserEvents, 8)
}
