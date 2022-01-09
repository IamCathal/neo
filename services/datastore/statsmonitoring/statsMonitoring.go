package statsmonitoring

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
)

func getCPUUsagePercentage(cpuUsage *float64) {
	for {
		before, err := cpu.Get()
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Duration(1) * time.Second)
		after, err := cpu.Get()
		if err != nil {
			log.Fatal(err)
		}
		total := float64(after.Total - before.Total)

		*cpuUsage = float64(after.User-before.User) / total * 100
		time.Sleep(8 * time.Second)
	}
}

func getTotalRAM(totalRAM *float64) {
	memory, err := memory.Get()
	if err != nil {
		log.Fatal(err)
	}
	memTotalMB := float64(((memory.Total / 1000) / 1000))
	*totalRAM = memTotalMB
}

func GetCurrentRAMUsage(ramUsedNow *float64) {
	for {
		memory, err := memory.Get()
		if err != nil {
			log.Fatal(err)
		}
		memUsedMB := float64(((memory.Used / 1000) / 1000))
		*ramUsedNow = memUsedMB
		time.Sleep(8 * time.Second)
	}
}

func CollectAndShipStats() {

	var cpuUsagePercentage float64 = 0
	var totalRAM float64 = 0
	var RAMUsedCurrently float64 = 0

	// Get the total ram available once
	getTotalRAM(&totalRAM)
	// Get live CPU and RAM usage stats on a regular interval
	go getCPUUsagePercentage(&cpuUsagePercentage)
	go GetCurrentRAMUsage(&RAMUsedCurrently)

	client := influxdb2.NewClient(os.Getenv("INFLUXDB_URL"), os.Getenv("SYSTEM_STATS_BUCKET_TOKEN"))
	defer client.Close()

	writeAPI := client.WriteAPIBlocking(os.Getenv("ORG"), os.Getenv("SYSTEM_STATS_BUCKET"))
	time.Sleep(5 * time.Second)

	for {
		point := influxdb2.NewPointWithMeasurement("systemStatusCPU").
			AddTag("system", os.Getenv("NODE_NAME")).
			AddField("cpu", cpuUsagePercentage).
			SetTime(time.Now())
		writeAPI.WritePoint(context.Background(), point)

		memUsagePercentage := float64(RAMUsedCurrently) / float64(totalRAM)

		point = influxdb2.NewPointWithMeasurement("systemStatusRAM").
			AddTag("system", fmt.Sprintf("%s (%.0f GB)", os.Getenv("NODE_NAME"), totalRAM/1000)).
			AddField("memory", math.Floor(memUsagePercentage*100)).
			SetTime(time.Now())
		writeAPI.WritePoint(context.Background(), point)
		time.Sleep(10 * time.Second)
	}
}
