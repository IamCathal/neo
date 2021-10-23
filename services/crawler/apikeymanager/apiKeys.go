package apikeymanager

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/datastructures"
)

var (
	keyGetLock sync.Mutex
)

func InitApiKeys() {
	APIKeysFromEnv := strings.Split(os.Getenv("STEAM_API_KEYS"), ",")
	for _, APIKey := range APIKeysFromEnv {
		newAPIKey := datastructures.APIKey{
			Key:      APIKey,
			LastUsed: time.Now(),
		}
		configuration.UsableAPIKeys.APIKeys = append(configuration.UsableAPIKeys.APIKeys, newAPIKey)
	}
}

// GetSteamAPIKey gets a steam API key. It picks any steam API key
// stored that has not been used in the last 1000ms to avoid keys
// being used too frequently. If none are found then the function
// waits a short period and tries again until one is returned
func GetSteamAPIKey() string {
	keyGetLock.Lock()
	for {
		for i, usableKey := range configuration.UsableAPIKeys.APIKeys {
			timeSinceLastUsed := time.Now().Sub(usableKey.LastUsed)
			if timeSinceLastUsed > time.Duration(1000*time.Millisecond) {
				configuration.UsableAPIKeys.APIKeys[i].LastUsed = time.Now()
				keyGetLock.Unlock()
				return usableKey.Key
			}
		}
		sleepTimeMs, err := strconv.Atoi(os.Getenv("KEY_SLEEP_TIME"))
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Duration(sleepTimeMs) * time.Millisecond)
	}
}
