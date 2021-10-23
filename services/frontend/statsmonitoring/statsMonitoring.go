package statsmonitoring

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go"
)

func getCPUUsagePercentage(cpuUsage *float64) {
	for {
		out, err := exec.Command("bash", "-c", "awk '{u=$2+$4; t=$2+$4+$5; if (NR==1){u1=u; t1=t;} else print ($2+$4-u1) * 100 / (t-t1) ; }' <(grep 'cpu ' /proc/stat) <(sleep 1;grep 'cpu ' /proc/stat)").CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}

		cpuUsageFloat, err := strconv.ParseFloat(strings.ReplaceAll(string(out), "\n", ""), 64)
		if err != nil {
			fmt.Printf("Err: %+v\n", err)
			*cpuUsage = 0
		}
		*cpuUsage = cpuUsageFloat
		time.Sleep(3 * time.Second)
	}
}

func getTotalRAM(totalRAM *float64) {
	out, err := exec.Command("bash", "-c", "free -t | awk 'NR == 2 {printf(\"%.2f\", $2/1000000)}'").CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	totalRAMFloat, err := strconv.ParseFloat(strings.ReplaceAll(string(out), "\n", ""), 64)
	if err != nil {
		fmt.Printf("Err: %+v\n", err)
		*totalRAM = 0
	}
	*totalRAM = totalRAMFloat
}

func getRamUsedCurrently(RAMUsedCurrently *float64) {
	for {
		out, err := exec.Command("bash", "-c", "free -t | awk 'NR == 2 {printf(\"%.2f\", $3/1000000)}'").CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}
		RAMUsedCurrentlyFloat, err := strconv.ParseFloat(strings.ReplaceAll(string(out), "\n", ""), 64)
		if err != nil {
			fmt.Printf("Err: %+v\n", err)
			*RAMUsedCurrently = 0
		}
		*RAMUsedCurrently = RAMUsedCurrentlyFloat
		time.Sleep(3 * time.Second)
	}
}

func getTotalDiskSpaceUsed(diskSpaceUsed *float64) {
	for {
		out, err := exec.Command("bash", "-c", "df --output=size --total | awk 'END {print $1}'").CombinedOutput()
		if err != nil {
			fmt.Println(err)
		}
		diskSpaceUsedCurrently, err := strconv.ParseFloat(strings.ReplaceAll(string(out), "\n", ""), 64)
		if err != nil {
			// fmt.Println(err)
			*diskSpaceUsed = 0
		}
		*diskSpaceUsed = diskSpaceUsedCurrently / 1000000
		time.Sleep(3 * time.Second)
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
	go getRamUsedCurrently(&RAMUsedCurrently)

	client := influxdb2.NewClient(os.Getenv("INFLUXDB_URL"), os.Getenv("SYSTEM_STATS_BUCKET_TOKEN"))
	defer client.Close()

	writeAPI := client.WriteAPIBlocking(os.Getenv("ORG"), os.Getenv("SYSTEM_STATS_BUCKET"))
	time.Sleep(5 * time.Second)

	for {
		// fmt.Printf("Current cpu usage is: %v%% - Total RAM: %v Gb - Used now: %v Gb\n", cpuUsagePercentage, totalRAM, RAMUsedCurrently)
		time.Sleep(5 * time.Second)
		point := influxdb2.NewPointWithMeasurement("systemStatusCPU").
			AddTag("system", os.Getenv("NODE_NAME")).
			AddField("cpu", cpuUsagePercentage).
			SetTime(time.Now())
		writeAPI.WritePoint(context.Background(), point)

		point = influxdb2.NewPointWithMeasurement("systemStatusRAM").
			AddTag("system", fmt.Sprintf("%s (%.0f GB)", os.Getenv("NODE_NAME"), math.Floor(totalRAM))).
			AddField("memory", math.Floor((RAMUsedCurrently/totalRAM)*100)).
			SetTime(time.Now())
		writeAPI.WritePoint(context.Background(), point)

		time.Sleep(time.Second * 1)
	}
}
