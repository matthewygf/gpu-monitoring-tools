// modified from NVIDIA/gpu-monitoring-tools/bindings/go/samples/processInfo
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/matthewygf/gpu-monitoring-tools/bindings/go/nvml"
)

var tocsv = flag.Bool("csv", false, "write values to csv instead.")
var filepath = flag.String("logpath", "processinfo.csv", "path to create the csv file.")
var interval = flag.Int("interval", 500, "interval time to run the profiler, in milliseconds")
var once = flag.Bool("once", true, "execute only once and exit with code 0.")

func checkAndPrintErrorNoFormat(message string, err error) {
	if err != nil {
		log.Fatalln(message, err)
	}
}

func main() {
	flag.Parse()
	nvml.Init()
	defer nvml.Shutdown()

	var fileHandle *os.File
	var err error
	var writer *csv.Writer
	if *tocsv {
		fileHandle, err = os.Create(*filepath)
		checkAndPrintErrorNoFormat("Could not create file", err)
		defer func() {
			fileHandle.Close()
		}()
	}

	if fileHandle != nil {
		writer = csv.NewWriter(fileHandle)
		header := []string{"gpu_idx", "pid", "sm_util", "mem_util", "mem_used", "command_name"}
		err := writer.Write(header)
		checkAndPrintErrorNoFormat("could not write to file:", err)
		writer.Flush()
	}

	count, err := nvml.GetDeviceCount()
	if err != nil {
		log.Panicln("Error getting device count:", err)
	}

	var devices []*nvml.Device
	for i := uint(0); i < count; i++ {
		device, err := nvml.NewDevice(i)
		if err != nil {
			log.Panicf("Error getting device %d: %v\n", i, err)
		}
		devices = append(devices, device)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	intervalTime := time.Duration(*interval)
	ticker := time.NewTicker(time.Millisecond * intervalTime)
	defer ticker.Stop()
	if fileHandle == nil {
		fmt.Printf("gpu,pid,sm_util,mem_util,mem_used,name\n")
	}

	for {
		select {
		case <-ticker.C:
			for i, device := range devices {
				processUtils, err := device.GetProcessUtilization()
				if err != nil {
					log.Panicf("Error getting device %d processes utilization %v \n", i, err)
				} else {
					for j := range processUtils {
						if processUtils[j].SmUtil > 0 {
							name, err := device.SystemGetProcessName(processUtils[j].PID)
							if err != nil {
								log.Panicf("Error getting device %d proccess %d name %v \n", i, processUtils[j].PID, err)
							}
							if fileHandle != nil {
								row := []string{
									strconv.FormatInt(int64(i), 10),
									strconv.FormatUint(uint64(processUtils[j].PID), 10),
									strconv.FormatUint(uint64(processUtils[j].SmUtil), 10),
									strconv.FormatUint(uint64(processUtils[j].MemUtil), 10),
									strconv.FormatUint(uint64(processUtils[j].MemUsed), 10),
									name}
								err := writer.Write(row)
								checkAndPrintErrorNoFormat("Could not write row", err)
								writer.Flush()
							} else {
								fmt.Printf("%5v,%5v,%5v,%5v,%v\n",
									i, processUtils[j].PID, processUtils[j].SmUtil, processUtils[j].MemUtil, processUtils[j].MemUsed, name)
							}
						}
					}
				}
			}

			if *once {
				os.Exit(0)
			}
		case <-sigs:
			return
		}
	}
}
