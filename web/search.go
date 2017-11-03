// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"strconv"

	log "git.townsourced.com/townsourced/logrus"
	"github.com/timshannon/townsourced/app"
	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/data/private"
	"github.com/timshannon/townsourced/fail"
)

var errTownRequired = fail.New("You must specify at least one town, or be logged in when searching posts")

func searchTemplate(w http.ResponseWriter, r *http.Request, c context) {
	result := struct {
		Posts []app.Post
		User  *app.User
		Town  *app.Town
		Error string
	}{}

	townKey := c.params.ByName("town")
	var err error

	if townKey != "" {
		//specific town search
		result.Town, err = app.TownGet(data.NewKey(townKey))
		if err == app.ErrTownNotFound {
			four04(w, r)
			return
		}
		if errHandledPage(err, w, r, c) {
			return
		}
	}

	result.User, result.Posts, err = postSearch(r, c, result.Town)
	if err == errTownRequired {
		http.Redirect(w, r, "/search/location", http.StatusTemporaryRedirect)
		return
	}

	if err != nil {
		if !fail.IsFail(err) {
			errHandledPage(err, w, r, c)
			return
		}
		result.Error = err.Error()
	}

	err = w.(*templateWriter).execute("SEARCH", result)
	if err != nil {
		log.Errorf("Error executing SEARCH template: %s", err)
	}
}

func searchLocationTemplate(w http.ResponseWriter, r *http.Request, c context) {
	result := struct {
		Posts            []app.Post
		User             *app.User
		GoogleMapsAPIKey string
		Error            string
	}{}

	var err error

	result.GoogleMapsAPIKey = private.GoogleMapsAPIKey

	result.User, result.Posts, err = postSearch(r, c, nil)

	if err != nil && err != errTownRequired {
		if !fail.IsFail(err) {
			errHandledPage(err, w, r, c)
			return
		}
		result.Error = err.Error()
	}

	err = w.(*templateWriter).execute("SEARCHLOCATION", result)
	if err != nil {
		log.Errorf("Error executing SEARCHLOCATION template: %s", err)
	}
}

func postsGet(w http.ResponseWriter, r *http.Request, c context) {
	// ?town=<town>&town=<town>
	// ?category=<category>
	// ?since=<since>&limit=100
	// ?showModerated

	var u *app.User
	var err error

	if c.session != nil {
		u, err = c.session.User()
		if errHandled(err, w, r, c) {
			return
		}
	}

	values := r.URL.Query()
	since, limit, err := sinceLimitValues(values, postPageLimit)
	if errHandled(err, w, r, c) {
		return
	}

	showModerated, err := strconv.ParseBool(values.Get("showModerated"))
	if err != nil {
		showModerated = false
	}

	towns := values["town"]

	var townKeys []data.Key
	if len(towns) == 0 {
		if u == nil {
			unauthorized(w, r)
			return
		}
		townKeys = data.KeyWhenSlice(u.TownKeys).Keys()
	} else {
		townKeys = data.NewKeySlice(towns)
	}

	posts, err := app.PostGetByTowns(u, townKeys, values.Get("category"), since, limit, showModerated)
	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   posts,
	})
}

func postSearchGet(w http.ResponseWriter, r *http.Request, c context) {
	_, posts, err := postSearch(r, c, nil)
	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   posts,
	})
}

func postSearch(r *http.Request, c context, town *app.Town) (user *app.User, posts []app.Post, err error) {
	// ?town=<town>&town=<town>
	// ?tag=<tag>&tag=<tag>
	// ?category=<category>
	// ?search=<searchstring>
	// ?latitude=<latitude>
	// ?longitude=<longitude>
	// ?milesDistant=<miles>
	// ?minPrice=<price>
	// ?maxPrice=<price>
	// ?sort=<none|priceHigh|priceLow>
	// ?from=<index>&limit=100
	// ?showModerated

	values := r.URL.Query()

	limit, err := strconv.Atoi(values.Get("limit"))
	if err != nil {
		limit = postPageLimit
	}
	from, err := strconv.Atoi(values.Get("from"))
	if err != nil {
		from = 0
	}

	showModerated, err := strconv.ParseBool(values.Get("showModerated"))
	if err != nil {
		showModerated = false
	}

	townsInput := values["town"]
	tags := values["tag"]
	search := values.Get("search")
	category := values.Get("category")

	latitude := values.Get("latitude")
	longitude := values.Get("longitude")
	milesDistant := values.Get("milesDistant")

	sort := values.Get("sort")

	if c.session != nil {
		user, err = c.session.User()
		if err != nil {
			return nil, nil, err
		}
	}

	var towns []app.Town

	if town == nil {
		if len(townsInput) == 0 {
			if latitude == "" || longitude == "" {
				if user == nil {
					return nil, nil, errTownRequired
				}

				towns, err = user.Towns()
				if err != nil {
					return nil, nil, err
				}
			} else {
				lat, err := strconv.ParseFloat(latitude, 64)
				if err != nil {
					return nil, nil, errTownRequired
				}
				lng, err := strconv.ParseFloat(longitude, 64)
				if err != nil {
					return nil, nil, errTownRequired
				}
				miles, err := strconv.ParseFloat(milesDistant, 64)
				if err != nil {
					miles = townDefaultMilesDistant
				}
				towns, err = app.TownSearchDistance(lng, lat, miles, 0, 1000)
				if err != nil {
					return nil, nil, err
				}
			}
		} else {
			townKeys := data.NewKeySlice(townsInput)
			towns, err = app.TownsGet(townKeys...)
			if err != nil {
				return nil, nil, err
			}
		}
	} else {
		towns = []app.Town{*town}
	}

	if search == "" && len(tags) == 0 {
		//no tags or search string specified
		return user, posts, nil
	}

	minPrice := -1.0
	maxPrice := -1.0

	minPrice, err = strconv.ParseFloat(values.Get("minPrice"), 64)
	if err != nil {
		minPrice = -1
	}
	maxPrice, err = strconv.ParseFloat(values.Get("maxPrice"), 64)
	if err != nil {
		maxPrice = -1
	}

	posts, err = app.PostSearch(user, search, tags, towns, category, from, limit, sort, minPrice, maxPrice, showModerated)
	if err != nil {
		return nil, nil, err
	}
	return user, posts, nil
}
