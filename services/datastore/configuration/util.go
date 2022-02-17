package configuration

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/neosteamfriendgraphing/common/util"
)

func GetLocalIPAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(util.MakeErr(err))
	}
	defer conn.Close()
	addrWithNoPort := strings.Split(conn.LocalAddr().(*net.UDPAddr).String(), ":")
	return addrWithNoPort[0]
}

func GetCurrentTimeInMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GetRequestStartTimeInTimeFormat(requestStartTimeString string) int64 {
	requestStartTime, err := strconv.ParseInt(requestStartTimeString, 10, 64)
	if err != nil {
		panic(util.MakeErr(err))
	}
	return requestStartTime
}

func isEnvSet(envName string) (string, bool) {
	value, exists := os.LookupEnv(envName)
	return value, exists
}
