// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package data

import (
	"errors"
	"strings"
	"time"

	"git.townsourced.com/townsourced/elastic"
	rt "git.townsourced.com/townsourced/gorethink"
)

const (
	// PostStatusDraft not public visible yet
	PostStatusDraft = "draft"
	// PostStatusPublished publicly visible
	PostStatusPublished = "published"
	//PostStatusClosed not searchable but link is still valid
	PostStatusClosed = "closed"
)

/* PostSort sets the sort order on post search results */
const (
	PostSearchSortNone           = ""
	PostSearchPriceSortHighToLow = "pricehigh"
	PostSearchPriceSortLowToHigh = "pricelow"
	PostSearchSortNew            = "new"
	PostSearchSortOld            = "old"
)

func init() {
	tables = append(tables, tblPost)
	searchTypes = append(searchTypes, srcPost)
}

var tblPost = &table{
	name: "post",
	indexes: []index{
		index{
			name: "Creator",
			indexFunc: func(row rt.Term) interface{} {
				return []interface{}{row.Field("Creator"), row.Field("Updated")}
			},
		},
		index{name: "Published"},
	},
}

var srcPost = &searchType{
	name: "post",
	properties: map[string]interface{}{
		//"_all": map[string]interface{}{ // issue with elasticsearch 2.2?
		//"enabled": false,
		//},
		"title": map[string]interface{}{
			"type":     "string",
			"analyzer": "snowball",
		},
		"content": map[string]interface{}{
			"type":     "string",
			"analyzer": "snowball",
		},
		"category": map[string]interface{}{
			"type":  "string",
			"index": "not_analyzed",
		},
		"townKeys": map[string]interface{}{
			"type":  "string",
			"index": "not_analyzed",
		},
		"hashTags": map[string]interface{}{
			"type":  "string",
			"index": "not_analyzed",
		},
		"prices": map[string]interface{}{
			"type":  "double",
			"index": "not_analyzed",
		},
		"moderation": map[string]interface{}{
			"type": "nested",
			"properties": map[string]interface{}{
				"town": map[string]interface{}{
					"type":  "string",
					"index": "not_analyzed",
				},
			},
		},
	},
}

// fields to pluck for when posts are gathered in a listing
//	(i.e. full content and images aren't needed unless someone is viewing the full document
// and I have to do this twice because elastic and rethink have different APIs
var postListPluck = []interface{}{"Key", "Title", "Category", "TownKeys", "Status", "Moderation",
	"Creator", "Created", "Updated", "Published", "FeaturedImage", "HashTags", "Prices"}

var postListFields = []string{"key", "title", "category", "townKeys", "status", "moderation",
	"creator", "created", "updated", "published", "featuredImage", "hashTags", "prices"}

// PostInsert inserts a new post into the database
func PostInsert(post interface{}) (UUID, error) {
	w, err := tblPost.Insert(post).RunWrite(session)
	err = wErr(w, err)
	if err != nil {
		return EmptyUUID, err
	}

	if len(w.GeneratedKeys) != 1 {
		return EmptyUUID, errors.New("No new key generated for this post")
	}

	return UUID(w.GeneratedKeys[0]), nil
}

// PostIndex indexes the post for full text searching
func PostIndex(post interface{}, key UUID) error {
	return srcPost.index(string(key), post)
}

// PostRemoveIndex removes the given post from the full text search index
func PostRemoveIndex(key UUID) error {
	return srcPost.delete(string(key))
}

// PostUpdate updates an existing post
func PostUpdate(post interface{}, key UUID) error {
	return tryUpdateVersion(tblPost.Get(key), post)
}

// PostGet retrieves a post by a specific post key
func PostGet(result interface{}, key UUID) error {
	c, err := tblPost.Get(key).Run(session)
	if err != nil {
		return err
	}

	if c.IsNil() {
		return ErrNotFound
	}

	return c.One(result)
}

// PostGetByUser retrieves posts by a specific user
func PostGetByUser(result interface{}, username Key, status string, public bool, since time.Time, limit int) (err error) {
	var sinceOp interface{} = rt.MaxVal

	if !since.IsZero() {
		sinceOp = since
	}

	trm := tblPost.Between([]interface{}{username, rt.MinVal}, []interface{}{username, sinceOp},
		rt.BetweenOpts{
			Index:     "Creator",
			LeftBound: "open",
		}).OrderBy(rt.OrderByOpts{
		Index: rt.Desc("Creator"),
	}).Pluck(postListPluck...)

	if public {
		if status != PostStatusPublished {
			status = PostStatusPublished
		}
		trm = postPublic(trm)
	}

	if status != "" {
		if public && status != PostStatusPublished {
			status = PostStatusPublished
		}
		trm = trm.Filter(rt.Row.Field("Status").Eq(status))
	}

	c, err := trm.Limit(limit).Run(session)

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

// PostGetUserSaved retrieves posts saved by a specific user in the order in which they were saved
func PostGetUserSaved(result interface{}, username Key, status string, from, limit int) (err error) {

	trm := tblUser.Get(username).Field("SavedPosts").EqJoin("Key", tblPost.Term).
		Without(map[string]interface{}{"right": "When"}).Zip() // drop When on right to make sure we sort by KeyWhen

	if status != "" {
		trm = trm.Filter(rt.Row.Field("Status").Eq(status))
	}

	c, err := trm.Skip(from).Limit(limit).OrderBy(rt.Desc("When")).Pluck(postListPluck...).Run(session)

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

// PostGetByTowns retrieves posts by the passed in list of townkeys
func PostGetByTowns(result interface{}, towns []Key, category string, since time.Time, limit int,
	showModerated bool) (err error) {
	var sinceOp interface{} = rt.MaxVal

	if !since.IsZero() {
		sinceOp = since
	}

	trm := tblPost.Between(rt.MinVal, sinceOp, rt.BetweenOpts{
		Index:     "Published",
		LeftBound: "open",
	}).OrderBy(rt.OrderByOpts{
		Index: rt.Desc("Published"),
	})

	trm = postFilterByTowns(trm, towns, showModerated)
	if category != "" {
		trm = trm.Filter(rt.Row.Field("Category").Eq(category))
	}

	c, err := trm.Filter(rt.Row.Field("Status").Eq(PostStatusPublished)).Pluck(postListPluck...).
		Limit(limit).Run(session)

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

// PostSearch retrieves posts in relevant order by the search text or tags, for the given towns
// use PostSort* enumeration for sorting
func PostSearch(searchText string, tags []string, towns []Key, category string, from, limit int, postSort string,
	minPrice, maxPrice float64, showModerated bool) (*SearchResult, error) {
	qry := elastic.NewBoolQuery()

	if searchText != "" {
		qry = qry.Must(elastic.NewMultiMatchQuery(searchText, "title", "content"))
	}

	if minPrice > -1 && maxPrice > -1 {
		qry = qry.Must(elastic.NewRangeQuery("prices").Gte(minPrice).Lte(maxPrice))
	} else if minPrice > -1 {
		qry = qry.Must(elastic.NewRangeQuery("prices").Gte(minPrice))
	} else if maxPrice > -1 {
		qry = qry.Must(elastic.NewRangeQuery("prices").Lte(maxPrice))
	}

	for i := range tags {
		qry = qry.Must(elastic.NewMatchQuery("hashTags", tags[i]))
	}

	if category != "" {
		qry = qry.Must(elastic.NewTermQuery("category", category))
	}

	var townQ []elastic.Query
	for i := range towns {
		if showModerated {
			townQ = append(townQ, elastic.NewTermQuery("townKeys", towns[i]))
		} else {
			townQ = append(townQ, elastic.NewBoolQuery().
				Must(elastic.NewTermQuery("townKeys", towns[i])).
				MustNot(elastic.NewNestedQuery("moderation",
					elastic.NewTermQuery("moderation.town", towns[i]))))
		}

	}

	qry = qry.Should(townQ...).MinimumNumberShouldMatch(1)

	src := srcPost.search(qry)

	switch strings.ToLower(postSort) {
	case PostSearchPriceSortHighToLow:
		src = src.SortBy(elastic.NewFieldSort("prices").SortMode("avg").Desc())
	case PostSearchPriceSortLowToHigh:
		src = src.SortBy(elastic.NewFieldSort("prices").SortMode("avg").Asc())
	case PostSearchSortNew:
		src = src.SortBy(elastic.NewFieldSort("published").Desc())
	case PostSearchSortOld:
		src = src.SortBy(elastic.NewFieldSort("published").Asc())
	}

	src = src.SortBy(elastic.NewScoreSort().Desc()).
		Sort("published", false)

	result, err := src.
		From(from).
		Size(limit).
		FetchSourceContext(elastic.NewFetchSourceContext(true).Include(postListFields...)).
		Do()

	//TODO: Use Fields to limit result set, rather than source filtering, as it's quicker
	// however, it breaks how we're currently doing result sets
	if err != nil {
		return nil, err
	}

	if result.TotalHits() == 0 {
		return nil, ErrNotFound
	}

	return &SearchResult{
		result: result,
		index:  0,
	}, nil
}

func postFilterByTowns(trm rt.Term, towns []Key, showModerated bool) rt.Term {
	return trm.Filter(func(post rt.Term) rt.Term {
		return post.Field("TownKeys").Contains(func(townKey rt.Term) rt.Term {
			if showModerated {
				return rt.Expr(towns).Contains(townKey)
			}
			//check for moderation
			return rt.Branch(rt.Expr(towns).Contains(townKey),
				post.Field("Moderation").Map(func(mod rt.Term) rt.Term {
					return mod.Field("Town")
				}).Contains(townKey).Not(),
				false)

		})
	})
}

// PostAllCount returns the count of the total number of posts, usually used by maintenance and not the frontend
func PostAllCount() (int, error) {
	c, err := tblPost.Count().Run(session)
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

// PostGetAll retrieves all posts
// This likely shouldn't be used for the actual website, and should only be used for maintenance / tasks
func PostGetAll(result interface{}, from, limit int) error {
	c, err := tblPost.Skip(from).Limit(limit).Run(session)
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

// post public checks if a given post is visible to the public
func postPublic(trm rt.Term) rt.Term {
	return trm.Filter(func(post rt.Term) rt.Term {
		return postNotModeratedInAllTowns(post)
	}).Filter(func(post rt.Term) rt.Term {
		return postIsPublic(post)
	})
}

func postNotModeratedInAllTowns(post rt.Term) rt.Term {
	//check if moderated in all towns
	return post.Field("TownKeys").Contains(func(townKey rt.Term) rt.Term {
		return post.Field("Moderation").Map(func(mod rt.Term) rt.Term {
			return mod.Field("Town")
		}).Contains(townKey).Not()
	})
}

func postIsPublic(post rt.Term) rt.Term {
	return post.Field("TownKeys").Contains(func(key rt.Term) rt.Term {
		return tblTown.Get(key).Field("Private").Eq(false)
	})
}
