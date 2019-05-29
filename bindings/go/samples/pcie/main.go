// modified from NVIDIA/gpu-monitoring-tools/bindings/go/samples/dmon

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
var interval = flag.Int("interval", 1, "interval time to run the profiler")

func main() {
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
		// bar1 in MiB
		// throughput are both in MB
		header := []string{"gpu_idx", "bar1_used", "pcie_read", "pcie_write"}
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
	if fileHandle == nil {
		fmt.Printf("gpu_idx,bar1_used,pcie_read,pcie_write")
	}
	for {
		select {
		case <-ticker.C:
			for i, device := range devices {
				st, err := device.Status()
				if err != nil {
					log.Panicf("Error getting device %d status: %v\n", i, err)
				}
				if fileHandle != nil {
					row := []string{
						strconv.FormatInt(i),
						strconv.FormatUint(*st.PCI.BAR1Used, 10),
						strconv.FormatUint(*st.PCI.Throughput.RX, 10),
						strconv.FormatUint(*st.PCI.Throughput.TX, 10)}
				} else {
					fmt.Printf("%5d,%5d,%5d,%5d\n",
						i, *st.PCI.BAR1Used, *st.PCI.Throughput.RX, *st.PCI.Throughput.TX)
				}
			}
		case <-sigs:
			return
		}
	}
}
