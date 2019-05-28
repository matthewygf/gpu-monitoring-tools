package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
)

var usecsv = flag.Bool("csv", false, "csv format")
var path = flag.String("logpath", "processinfo.csv", "file path for logging")

func main() {
	flag.Parse()
	fmt.Printf("%v\n", *path)
	fmt.Printf("%t\n", *usecsv)

	var fileHandle *os.File
	var err error
	var writer *csv.Writer
	if *usecsv {
		fileHandle, err = os.Create(*path)
		checkError("couldn't create", err)
		defer func() {
			fmt.Printf("gonna close \n")
			fileHandle.Close()
		}()
	}

	if fileHandle != nil {
		writer = csv.NewWriter(fileHandle)
		header := []string{"head1", "head2", "head3", "head4", "head5"}
		err := writer.Write(header)
		checkError("couldn't write", err)
		writer.Flush()
	}

	if fileHandle != nil {
		for i := 1; i <= 10; i++ {
			value1 := fmt.Sprintf("v%d", i*1)
			value2 := fmt.Sprintf("v%d", i*2)
			value3 := fmt.Sprintf("v%d", i*3)
			value4 := fmt.Sprintf("v%d", i*4)
			value5 := fmt.Sprintf("v%d", i*5)
			values := []string{value1, value2, value3, value4, value5}
			err := writer.Write(values)
			checkError("couldn't write", err)
			writer.Flush()
		}
	}

	fmt.Printf("i did some stuff \n")
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
