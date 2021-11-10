package endpoints

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

// func TestGetAPIStatus(t *testing.T) {
// 	mockController := &controller.MockCntrInterface{}
// 	endpoints := Endpoints{
// 		mockController,
// 	}

// 	assert.HTTPStatusCode(t, endpoints.Status, "POST", "/status", nil, 200)
// 	assert.HTTPBodyContains(t, endpoints.Status, "POST", "/status", nil, "operational")
// }
