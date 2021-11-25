package apikeymanager

import (
	"os"
	"testing"
	"time"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestThatValidAPIKeysCanBeExtractedFromEnvFile(t *testing.T) {
	os.Setenv("STEAM_API_KEYS", "Quick,Brown,Fox,Ran")
	os.Setenv("KEY_USAGE_TIMER", "1916")

	InitApiKeys()
	expectedAPIKeys := []string{"Quick", "Brown", "Fox", "Ran"}
	actualAPIKeys := []string{}
	for _, APIKey := range configuration.UsableAPIKeys.APIKeys {
		actualAPIKeys = append(actualAPIKeys, APIKey.Key)
	}

	assert.Equal(t, actualAPIKeys, expectedAPIKeys)
}

func TestGetSteamAPIKeyThrottlesRequests(t *testing.T) {
	keySleepTime := 15
	os.Setenv("STEAM_API_KEYS", "Quick,Brown,Fox,Ran")
	os.Setenv("KEY_USAGE_TIMER", "1916")
	InitApiKeys()

	startTime := time.Now()
	_ = GetSteamAPIKey()
	timeTaken := time.Since(startTime) * time.Millisecond

	assert.GreaterOrEqual(t, timeTaken, time.Duration(keySleepTime)*time.Millisecond)
}
