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
		fileHandle, err = os.Create("test.csv")
		checkError("couldn't create", err)
		defer func() {
			fmt.Printf("gonna close \n")
			fileHandle.Close()
		}()
	}

	if fileHandle != nil {
		writer = csv.NewWriter(fileHandle)
		header := []string{"1", "2", "3", "4", "5"}
		err := writer.Write(header)
		checkError("couldn't write", err)
		writer.Flush()
	}

	fmt.Printf("i did some stuff \n")
	// fileHandle, err := os.Create("test.csv")
	// checkError("File not created", err)
	// // until the end of main()
	// defer fileHandle.Close()

	// writer := csv.NewWriter(fileHandle)
	// // until the end of main()
	// defer writer.Flush()

	// for _, value := range data {
	// 	err := writer.Write(value)
	// 	checkError("cannot write", err)
	// }
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
