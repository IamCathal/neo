package controller

import (
	"sync"
	"time"

	"github.com/iamcathal/neo/services/crawler/configuration"
	commonUtil "github.com/neosteamfriendgraphing/common/util"
)

var (
	requestMakeLock     sync.Mutex
	TimeBetweenRequests = time.Duration(3 * time.Millisecond)
	lastRequestTime     = time.Now()
	activeRequests      = 0
)

// MakeNetworkGETRequest limits the throughput for network GET
// requests. This is to stop network io timeouts from my router's DNS
// that occured through too many network requests being initiated
// concurrently
//
// This allows roughly one GET request to be initiated at any given
// time but does not wait for the response before allowing another
// to be generated
func MakeNetworkGETRequest(targetURL string) ([]byte, error) {
	startTime := time.Now()
	requestMakeLock.Lock()
	for {
		if time.Since(lastRequestTime) > time.Duration(1*time.Nanosecond) {
			configuration.Logger.Sugar().Infof("waited %v and time since last request is %v\n", time.Since(startTime), time.Since(lastRequestTime))
			lastRequestTime = time.Now()
			// Roughly 1ms after the request is made allow the lock to be unlocked.
			// We only want one request being initiated at a time but don't care about
			// waiting for the response
			go func() {
				time.Sleep(1 * time.Millisecond)
				requestMakeLock.Unlock()
			}()
			res, err := commonUtil.GetAndRead(targetURL)
			return res, err
		}
	}
}
