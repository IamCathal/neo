package apikeymanager

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/datastructures"
)

var (
	keyGetLock   sync.Mutex
	keyUsageTime time.Duration
)

// InitApiKeys initialises the structure that manages rate limitted
// access to the steam web API keys
func InitApiKeys() {
	APIKeysFromEnv := strings.Split(os.Getenv("STEAM_API_KEYS"), ",")
	for _, APIKey := range APIKeysFromEnv {
		newAPIKey := datastructures.APIKey{
			Key:      APIKey,
			LastUsed: time.Now(),
		}
		configuration.UsableAPIKeys.APIKeys = append(configuration.UsableAPIKeys.APIKeys, newAPIKey)
	}

	keyTime, err := strconv.Atoi(os.Getenv("KEY_USAGE_TIMER"))
	if err != nil {
		panic(err)
	}
	keyUsageTime = time.Duration(keyTime)
}

// GetSteamAPIKey gets a steam API key. It picks any steam API key
// stored that has not been used in the last $KEY_SLEEP_TIME ms,
// If none are found then the function waits a short period
// and tries again until one is returned.
func GetSteamAPIKey() string {
	keyGetLock.Lock()
	for {
		for i, usableKey := range configuration.UsableAPIKeys.APIKeys {
			timeSinceLastUsed := time.Now().Sub(usableKey.LastUsed)
			if timeSinceLastUsed > keyUsageTime {
				configuration.UsableAPIKeys.APIKeys[i].LastUsed = time.Now()
				keyGetLock.Unlock()
				return usableKey.Key
			}
		}
		time.Sleep(time.Duration(3) * time.Millisecond)
	}
}
