package apikeymanager

import (
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/iamcathal/neo/services/crawler/configuration"
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

func TestThatValidAPIKeysCanBeExtractedFromEnvFile(t *testing.T) {
	os.Setenv("STEAM_API_KEYS", "Quick,Brown,Fox,Ran")
	os.Setenv("KEY_USAGE_TIMER", "1916")

	var waitG sync.WaitGroup
	waitG.Add(1)
	InitApiKeys(&waitG)
	waitG.Wait()

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
	var waitG sync.WaitGroup
	waitG.Add(1)
	InitApiKeys(&waitG)
	waitG.Wait()

	startTime := time.Now()
	_ = GetSteamAPIKey()
	timeTaken := time.Since(startTime) * time.Millisecond

	assert.GreaterOrEqual(t, timeTaken, time.Duration(keySleepTime)*time.Millisecond)
}
