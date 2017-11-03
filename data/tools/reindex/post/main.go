// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

// This tool runs through all posts in the configured database and makes sure that they are properly indexed in
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
		logrus.Fatalf("Error loading data.cfg.  Make sure data.cfg is in the running dir. ERROR: %s", err)
	}

	dataCfg := data.DefaultConfig()

	err = cfg.ValueToType("data", dataCfg)
	if err != nil {
		logrus.Fatalf("Error reading data config values: %s", err)
	}

	err = data.Init(dataCfg)
	if err != nil {
		logrus.Fatalf("Error initializing Data layer. ERROR: %v", err)
	}
	skip := 0
	published := 0

	total, err := data.PostAllCount()
	if err != nil {
		logrus.Fatalf("Error getting total post count")
	}

	progress := pb.StartNew(total)
	for {
		var posts []app.Post
		err = data.PostGetAll(&posts, skip, batchSize)
		if err != nil {
			logrus.Fatalf("Error getting posts.  ERROR: %s", err)
		}
		for i := range posts {
			key := posts[i].Key
			_ = data.PostRemoveIndex(key)
			// FIXME: IsNotFound check doesn't seem to actually work?
			//if err != nil && !elastic.IsNotFound(err) {
			//logrus.Fatalf("Error removing index for post %s ERROR: %s", key, err)
			//}

			if posts[i].Status == data.PostStatusPublished {
				err = data.PostIndex(posts[i], key)
				if err != nil {
					logrus.Fatalf("Error indexing post %s ERROR: %s", key, err)
				}
				published++
			}
			skip++
			progress.Increment()
		}
		if len(posts) < batchSize {
			break
		}
	}

	progress.FinishPrint(fmt.Sprintf("Reindexing complete %d posts published", published))
}
