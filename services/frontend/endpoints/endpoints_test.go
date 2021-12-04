package endpoints

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	endpoints Endpoints
)

func TestMain(m *testing.M) {
	endpoints = Endpoints{
		ApplicationStartUpTime: time.Now(),
	}

	code := m.Run()

	os.Exit(code)
}

func TestGetAPIStatus(t *testing.T) {
	assert.HTTPStatusCode(t, endpoints.Status, "POST", "/status", nil, 200)
	assert.HTTPBodyContains(t, endpoints.Status, "POST", "/status", nil, "operational")
}
