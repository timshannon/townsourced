// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

// This tool runs through all towns in the configured database and makes sure that they are properly indexed in
// elasticsearch

package main

import (
	"fmt"

	"git.townsourced.com/townsourced/config"
	"git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/pb"
	"git.townsourced.com/townsourced/townsourced/app"
	"git.townsourced.com/townsourced/townsourced/data"
)

var (
	batchSize = 100
)

func main() {
	cfg, err := config.LoadOrCreate("data.cfg")
	if err != nil {
		logrus.Fatalf("Error loading data.cfg.  Make sure data.cfg is in the running dir. ERROR: %v", err)
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

	skip := 0
	reindexed := 0

	total, err := data.TownAllCount()
	if err != nil {
		logrus.Fatalf("Error getting total town count. ERROR: %v", err)
	}

	progress := pb.StartNew(total)
	for {
		var towns []app.Town
		err = data.TownGetAll(&towns, skip, batchSize)
		if err != nil {
			logrus.Fatalf("Error getting towns.  ERROR: %v", err)
		}
		for i := range towns {
			_ = data.TownRemoveIndex(towns[i].Key)
			// FIXME: IsNotFound check doesn't seem to actually work?
			//if err != nil && !elastic.IsNotFound(err) {
			//logrus.Fatalf("Error removing index for town %s ERROR: %v", towns[i].Key, err)
			//}

			if !towns[i].Private {
				err = data.TownIndex(towns[i], towns[i].Key)
				if err != nil {
					logrus.Fatalf("Error indexing town %s ERROR: %v", towns[i].Key, err)
				}
				reindexed++
			}
			skip++
			progress.Increment()
		}
		if len(towns) < batchSize {
			break
		}
	}

	progress.FinishPrint(fmt.Sprintf("Reindexing complete %d towns reindexed", reindexed))
}
