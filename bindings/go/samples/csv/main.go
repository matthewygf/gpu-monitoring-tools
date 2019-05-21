package main

import (
	"encoding/csv"
	"log"
	"os"
)

var data = [][]string{{"line1", "hello readers of"}, {"line2", "golang"}}

func main() {
	fileHandle, err := os.Create("test.csv")
	checkError("File not created", err)
	// until the end of main()
	defer fileHandle.Close()

	writer := csv.NewWriter(fileHandle)
	// until the end of main()
	defer writer.Flush()

	for _, value := range data {
		err := writer.Write(value)
		checkError("cannot write", err)
	}
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
