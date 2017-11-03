// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"strings"

	rt "git.townsourced.com/townsourced/gorethink"
)

func init() {
	tables = append(tables, tblUser)
}

var tblUser = &table{
	name: "user",
	indexes: []index{
		index{name: "EmailSearch"},
		index{name: "GoogleID"},
		index{name: "FacebookID"},
		index{name: "TwitterID"},
		index{name: "Created"},
		index{
			name: "Name",
			indexFunc: func(row rt.Term) interface{} {
				return row.Field("Name").Downcase()
			},
		},
	},
	TableCreateOpts: rt.TableCreateOpts{
		PrimaryKey: "Username",
	},
}

// UserGet gets a user
func UserGet(result interface{}, username Key) error {
	c, err := tblUser.Get(username).Run(session)
	if err != nil {
		return err
	}

	if c.IsNil() {
		return ErrNotFound
	}

	return c.One(result)
}

// UserGetEmail gets a user with an email address
func UserGetEmail(result interface{}, email string) error {
	return userGetBy(result, "EmailSearch", strings.ToLower(email))
}

func userGetBy(result interface{}, index, key string) error {
	c, err := tblUser.GetAllByIndex(index, key).Run(session)
	if err != nil {
		return err
	}
	if c.IsNil() {
		return ErrNotFound
	}
	return c.One(result)
}

// UserGetMatching retrieves all users who's username starts with the passed in string
func UserGetMatching(result interface{}, match string, limit int) error {
	match = strings.ToLower(match)
	c, err := tblUser.Between(match, rt.MaxVal).
		Filter(func(row rt.Term) rt.Term {
			return row.Field("Username").Match("(?i)^" + match)
		}).
		Union(tblUser.Between(match, rt.MaxVal, rt.BetweenOpts{Index: "Name"}).
			Filter(func(row rt.Term) rt.Term {
				return row.Field("Name").Match("(?i)^" + match)
			})).OrderBy("Username").Limit(limit).Pluck("Username", "Name", "ProfileIcon").Run(session)

	if err != nil {
		return err
	}
	defer func() {
		if cerr := c.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	if c.IsNil() {
		return ErrNotFound
	}
	return c.All(result)
}

// UserGetGoogle gets a user with their GoogleID
func UserGetGoogle(result interface{}, googleID string) error {
	return userGetBy(result, "GoogleID", googleID)
}

// UserGetTwitter gets a user with their TwitterID
func UserGetTwitter(result interface{}, twitterID string) error {
	return userGetBy(result, "TwitterID", twitterID)
}

// UserGetFacebook gets a user with their FacebookID
func UserGetFacebook(result interface{}, facebookID string) error {
	return userGetBy(result, "FacebookID", facebookID)
}

// UserInsert inserts a new user into the database
func UserInsert(user interface{}) error {
	return wErr(tblUser.Insert(user).RunWrite(session))
}

// UserUpdate updates an existing user
func UserUpdate(user interface{}, username Key) error {
	return tryUpdateVersion(tblUser.Get(username), user)
}

// UserAllCount returns the count of the total number of users
func UserAllCount() (int, error) {
	c, err := tblUser.Count().Run(session)
	if err != nil {
		return -1, err
	}
	defer func() {
		if cerr := c.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	result := 0
	err = c.One(&result)
	if err != nil {
		return -1, err
	}
	return result, nil
}
