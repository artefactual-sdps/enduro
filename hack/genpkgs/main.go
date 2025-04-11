package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/google/uuid"
)

const (
	// How many records we want to generate.
	datasetSize = 1000

	// Size of each batch that we write to the CSV.
	batchSize = 100

	usage = `Usage: genpkgs TYPE

Generate a dataset of packages for testing.

TYPE: "sip" or "aip"
`
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatal(usage)
	}
	if args[0] != "sip" && args[0] != "aip" {
		log.Fatal(usage)
	}

	w := csv.NewWriter(os.Stdout)

	for i := range datasetSize {
		var row []string
		if args[0] == "sip" {
			row = sip(i + 1)
		} else {
			row = aip(i + 1)
		}

		if err := w.Write(row); err != nil {
			log.Println("error writing record to csv:", err)
		}
		if i%batchSize == 0 {
			w.Flush()
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}

func id() string {
	return uuid.New().String()
}

func sip(c int) []string {
	const doneStatus string = "2"
	return []string{
		strconv.Itoa(c),                     // id
		fmt.Sprintf("DPJ-SIP-%s.tar", id()), // name
		id(),                                // aip_id
		doneStatus,                          // status
		"2019-11-21 17:36:10",               // created_at
		"2019-11-21 17:36:11",               // started_at
		"2019-11-21 17:42:12",               // completed_at
	}
}

func aip(i int) []string {
	return []string{
		strconv.Itoa(i),             // id
		fmt.Sprintf("AIP-%s", id()), // name
		id(),                        // aip_id
		"stored",                    // status
		id(),                        // object_key
		"1",                         // location_id
		"2019-11-21 17:43:51",       // created_at
	}
}
