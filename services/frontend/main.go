package main

import (
	"log"
	"net"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var (
	nodeName string
	nodeDC   string
	logPath  string
	nodeIPV4 string
)

func GetLocalIPAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	addrWithNoPort := strings.Split(conn.LocalAddr().(*net.UDPAddr).String(), ":")
	return addrWithNoPort[0]
}

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	nodeName = os.Getenv("NODE_NAME")
	nodeDC = os.Getenv("NODE_DC")
	logPath = os.Getenv("LOG_PATH")
	nodeIPV4 = GetLocalIPAddress()
}

func InitLogger() *zap.Logger {
	os.OpenFile(logPath, os.O_RDONLY|os.O_CREATE, 0666)
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"stdout", logPath}

	globalLogFields := make(map[string]interface{})
	globalLogFields["nodeName"] = nodeName
	globalLogFields["nodeDC"] = nodeDC
	globalLogFields["nodeIPV4"] = nodeIPV4
	c.InitialFields = globalLogFields

	log, err := c.Build()
	if err != nil {
		panic(err)
	}
	return log
}

func main() {
	LoadConfig()
	logger := InitLogger()
	logger.Info("Hello world!")
}
