package controller

import (
	"fmt"
	"sync"
	"time"

	commonUtil "github.com/neosteamfriendgraphing/common/util"
)

var (
	requestMakeLock     sync.Mutex
	TimeBetweenRequests = time.Duration(3 * time.Millisecond)
	lastRequestTime     time.Time
	activeRequests      = 0
)

// MakeNetworkGETRequest limits the throughput for network GET
// requests. This is to stop network io timeouts that occured
// through too many network requests being made and executed
// concurrently
func MakeNetworkGETRequest(targetURL string) ([]byte, error) {
	// startTime := time.Now()
	for {
		if activeRequests < 10 {
			activeRequests++
			// fmt.Printf("waited %v\n", time.Since(startTime))
			res, err := commonUtil.GetAndRead(targetURL)
			activeRequests--
			return res, err
		} else {
			fmt.Printf("FULL\n\tFULL\n\t\tFULL\n\t\t\tFULL\n\t\t\t\tFULL\n\t\n")
			fmt.Printf("FULL\n\tFULL\n\t\tFULL\n\t\t\tFULL\n\t\t\t\tFULL\n\t\n")
			fmt.Printf("FULL\n\tFULL\n\t\tFULL\n\t\t\tFULL\n\t\t\t\tFULL\n\t\n")
			fmt.Printf("FULL\n\tFULL\n\t\tFULL\n\t\t\tFULL\n\t\t\t\tFULL\n\t\n")
		}
	}
}
