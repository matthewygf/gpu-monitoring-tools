package main

import (
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
var filepath = flag.String("logpath", "processinfo.log", "write the values to the logfile")

const (
	// PINFOHEADER are headers
	PINFOHEADER = `# gpu   pid   type  mem  Command
# Idx     #   C/G   MiB  name`
)

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func getFileHandle(create bool, path string) (*File, error) {
	if create {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err := os.MkdirAll(path, os.FileMode(0777))
			checkError("cannot create directories", err)
			fileHandle, err := os.Create(path)
			checkError("cannot create file", err)
		} else {
			fileHandle, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0777)
			checkError("cannot open file", err)
		}
	} else {
		fileHandle := nil
		err := nil
	}
	return fileHandle, err
}

func getCsvWriter(f *File) *Writer {
	if f != nil {
		return csv.NewWriter(f)
	}
	return nil
}

func main() {
	flag.Parse()
	nvml.Init()
	defer nvml.Shutdown()

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

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	fmt.Println(PINFOHEADER)

	fileHandle, err := getFileHandle(tocsv, logpath)
	writer := getCsvWriter(fileHandle)
	for {
		select {
		case <-ticker.C:
			for i, device := range devices {
				pInfo, err := device.GetAllRunningProcesses()
				if err != nil {
					log.Panicf("Error getting device %d processes: %v\n", i, err)
				}
				if len(pInfo) == 0 {
					fmt.Printf("%5v %5s %5s %5s %-5s\n", i, "-", "-", "-", "-")
				}

				for j := range pInfo {
					fmt.Printf("%5v %5v %5v %5v %-5v\n",
						i, pInfo[j].PID, pInfo[j].Type, pInfo[j].MemoryUsed, pInfo[j].Name)
					if writer != nil {
						data := [5]string{i, pInfo[j].PID, pInfo[j].Type, pInfo[j].MemoryUsed, pInfo[j].Name}
						writer.Write(data)
					}
					err := writer.Flush()
					checkError("cannot write", err)
				}

			}
		case <-sigs:
			return
		}
	}
	if writer != nil {
		fileHandle.close()
		writer.close()
	}
}
