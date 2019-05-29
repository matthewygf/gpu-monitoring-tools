package main

import (
	"strconv"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/matthewygf/gpu-monitoring-tools/bindings/go/nvml"
)

var tocsv = flag.Bool("csv", false, "write values to csv instead.")
var filepath = flag.String("logpath", "processinfo.csv", "path to create the csv file.")
var interval = flag.Int("interval", 1, "interval time to run the profiler")

const (
	// PINFOHEADER are headers
	PINFOHEADER = `# gpu   pid   type  mem  Command
# Idx     #   C/G   MiB  name`
)

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
		header := []string{"gpu_idx", "pid", "sm_util", "mem_util", "command_name"}
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

	intervalTime := *interval
	ticker := time.NewTicker(time.Second * intervalTime)
	defer ticker.Stop()
	var row []string
	if fileHandle == nil {
		fmt.Printf("gpu,pid,sm_util,mem_util")
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
								row = {
									strconv.FormatInt(i), 
									strconv.FormatUint(processUtils[j].PID), 
									strconv.FormatUint(processUtils[j].SmUtil),
									strconv.FormatUint(processUtils[j].MemUtil),
									name
								}
							} else {
								fmt.Printf("%5v,%5v,%5v,%5v\n",
								i, processUtils[j].PID, processUtils[j].SmUtil, processUtils[j].MemUtil, name)
							}
						}
					}
				}

				
				// pInfo, err := device.GetAllRunningProcesses()
				
				// if len(pInfo) == 0 {
				// 	fmt.Printf("%5v %5s %5s %5s %-5s\n", i, "-", "-", "-", "-")
				// }

				// fmt.Printf("%5v %5v %5v %5v %-5v\n",
				// 	i, pInfo[j].PID, pInfo[j].Type, pInfo[j].MemoryUsed, pInfo[j].Name)

				// if fileHandle != nil {
				// 	// TODO: all the values, and implement deviceAccountingStats.
				// 	row = { strconv.FormatInt(i)}
				// }
				// }
			}
		case <-sigs:
			return
		}
	}
}
