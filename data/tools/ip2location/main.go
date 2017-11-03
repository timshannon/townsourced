// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.

// This tool imports a ip2location csv file into RethinkDB

package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"git.townsourced.com/townsourced/config"
	"git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/pb"
	"github.com/timshannon/townsourced/app"
	"github.com/timshannon/townsourced/data"
)

const (
	batchSize = 200
)

var (
	flagCFG  = ""
	flagFile = ""
)

func init() {
	flag.StringVar(&flagCFG, "cfg", "./data.cfg", "Location of the database configuration file")
	flag.StringVar(&flagFile, "file", "./dbip-location.csv", "Location of the ip2location csv file")
}

func main() {
	flag.Parse()

	cfg, err := config.LoadOrCreate(flagCFG)
	if err != nil {
		logrus.Fatalf("Error loading %s. ERROR: %v", flagCFG, err)
	}

	dataCfg := data.DefaultConfig()

	err = cfg.ValueToType("data", dataCfg)
	if err != nil {
		logrus.Fatalf("Error reading data config values: %v", err)
	}

	err = data.Init(dataCfg)
	if err != nil {
		logrus.Fatalf("Error initializing Data layer. ERROR: %v", err)
	}

	// truncate table

	err = data.IP2LocationTruncate()
	if err != nil {
		logrus.Fatalf("Error truncating IP2Location table. ERROR: %v", err)
	}

	// parse file
	file, err := os.Open(flagFile)
	defer file.Close()

	if err != nil {
		logrus.Fatalf("Error opening ip2location data file %s. ERROR: %v", flagFile, err)
	}

	total, err := recordCount(file)
	if err != nil {
		logrus.Fatalf("Error getting record count from data file. ERROR: %v", err)
	}

	progress := pb.StartNew(total)

	//reset file reader

	_, err = file.Seek(0, 0)
	if err != nil {
		logrus.Fatalf("Error seeking to beginning of data file. ERROR: %v", err)
	}

	var readErr error
	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	entries := make([]*app.IPLocation, 0, batchSize)

	// import new data
	for {
		entries = entries[:0]

		for i := 0; i < batchSize; i++ {
			var records []string

			records, readErr = reader.Read()
			if readErr != nil {
				if readErr == io.EOF {
					break
				}

				logrus.Errorf("Error reading record from data file. ERROR: %v", readErr)
				continue
			}

			e, err := entry(records)
			if err != nil {
				logrus.Errorf("Error parsing record from data file. ERROR: %v", err)
				continue
			}

			entries = append(entries, e)

			progress.Increment()
		}

		err = data.IP2LocationImport(entries)
		if err != nil {
			logrus.Errorf("Error inserting parsed records from data file. ERROR: %v", err)
			continue
		}

		if readErr == io.EOF {
			break
		}
	}

	progress.FinishPrint(fmt.Sprintf("Import complete %d records imported", progress.Total))
}

func entry(records []string) (*app.IPLocation, error) {
	var latitude, longitude float64
	var err error

	latitude, err = strconv.ParseFloat(records[5], 64)
	if err != nil {
		return nil, err
	}

	longitude, err = strconv.ParseFloat(records[6], 64)
	if err != nil {
		return nil, err
	}

	return &app.IPLocation{
		IPFrom:      data.IPNumber(records[0]),
		IPTo:        data.IPNumber(records[1]),
		CountryCode: records[2],
		RegionName:  records[3],
		CityName:    records[4],
		Latitude:    latitude,
		Longitude:   longitude,
	}, nil
}

func recordCount(r io.Reader) (int, error) {
	buf := make([]byte, 8196)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}

	return count, nil
}
