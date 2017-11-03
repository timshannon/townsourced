// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"errors"
	"time"

	rt "git.townsourced.com/townsourced/gorethink"
)

const imageDatabase = "image"

func init() {
	tables = append(tables, tblImage)
}

var tblImage = &table{
	name:     "image",
	database: imageDatabase,
	indexes: []index{
		index{name: "InUse"},
	},
}

// ImageGet retrieves an image by it's key
func ImageGet(result interface{}, key UUID, thumb, placeholder bool) error {
	trm := tblImage.Get(key)

	if placeholder {
		trm = trm.Without("Data", "ThumbData")
	} else if thumb {
		trm = trm.Without("Data", "PlaceholderData")
	} else {
		trm = trm.Without("ThumbData", "PlaceholderData")
	}

	c, err := trm.Default(nil).Run(session)
	if err != nil {
		return err
	}

	if c.IsNil() {
		return ErrNotFound
	}

	return c.One(result)
}

// ImageInsert inserts a new image into the database
func ImageInsert(image interface{}) (UUID, error) {
	w, err := tblImage.Insert(image).RunWrite(session)
	err = wErr(w, err)
	if err != nil {
		return EmptyUUID, err
	}

	if len(w.GeneratedKeys) != 1 {
		return EmptyUUID, errors.New("No new key generated for this image")
	}

	return UUID(w.GeneratedKeys[0]), nil
}

// ImageUpdate updates an existing image
func ImageUpdate(image interface{}, key UUID) error {
	return tryUpdateVersion(tblImage.Get(key), image)
}

// ImageDelete deletes an image
func ImageDelete(key UUID) error {
	return wErr(tblImage.Get(key).Delete().RunWrite(session))
}

// ImageDeleteOrphans deletes all images that aren't currently in use
// and haven't been updated since the passed in time
func ImageDeleteOrphans(updatedSince time.Time) error {
	return wErr(tblImage.GetAllByIndex("InUse", false).Filter(rt.Row.Field("Updated").Le(updatedSince)).
		Delete(rt.DeleteOpts{Durability: "soft"}).RunWrite(session))
}
