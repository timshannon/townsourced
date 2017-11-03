// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "git.townsourced.com/townsourced/logrus"
	"github.com/timshannon/townsourced/app"
	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/data/private"
	"github.com/timshannon/townsourced/fail"
)

const (
	townDefaultMilesDistant  = 3.0
	townDefaultRetrieveLImit = 50
)

type townInput struct {
	Key         *data.Key `json:"key,omitempty"`
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Information *string   `json:"information,omitempty"`
	VerTag      *string   `json:"verTag,omitempty"`
	Longitude   *float64  `json:"longitude,omitempty"`
	Latitude    *float64  `json:"latitude,omitempty"`
	Color       *string   `json:"color,omitempty"`
	Moderator   *string   `json:"moderator,omitempty"`
	Private     *bool     `json:"private,omitempty"`
	Invitee     *string   `json:"invitee,omitempty"`
	Email       *string   `json:"email,omitempty"`
}

// rate limit the number of new towns that can be created by the same user
var townNewRequestType = app.RequestType{
	Type:         "townNew",
	FreeAttempts: 5,
	Scale:        5 * time.Second,
	Range:        10 * time.Minute,
	MaxWait:      1 * time.Minute,
}

func townTemplate(w http.ResponseWriter, r *http.Request, c context) {
	townKey := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(townKey)
	if err == app.ErrTownNotFound {
		//redirect to town search
		val := url.Values{}
		val.Add("search", c.params.ByName("town"))

		http.Redirect(w, r, "/town/?"+val.Encode(), http.StatusTemporaryRedirect)
		return
	}
	if errHandledPage(err, w, r, c) {
		return
	}

	var u *app.User
	if c.session != nil {
		u, err = c.session.User()
		if errHandledPage(err, w, r, c) {
			return
		}

		_, join := r.URL.Query()["join"]
		if join && town.Invited(u) {
			if errHandledPage(u.JoinTown(town), w, r, c) {
				return
			}
			if errHandledPage(u.Update(), w, r, c) {
				return
			}
			http.Redirect(w, r, "/town/"+c.params.ByName("town"), http.StatusTemporaryRedirect)
			return
		}

	}
	privateTown := false
	var posts []app.Post

	if town.CanSearch(u) {
		posts, err = town.Posts(u, "", time.Time{}, postPageLimit, false)
		if errHandledPage(err, w, r, c) {
			return
		}
	} else {
		privateTown = true
	}

	err = w.(*templateWriter).execute("TOWN", struct {
		Town          *app.Town
		Posts         []app.Post
		User          *app.User
		Private       bool
		FacebookAppID string
	}{
		Town:          town,
		Posts:         posts,
		User:          u,
		Private:       privateTown,
		FacebookAppID: private.FacebookAppID,
	})
	if err != nil {
		log.Errorf("Error executing town template: %s", err)
	}
}

func townSettingsTemplate(w http.ResponseWriter, r *http.Request, c context) {
	townKey := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(townKey)
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}

	if errHandledPage(err, w, r, c) {
		return
	}

	if c.session == nil {
		unauthorized(w, r)
		return
	}

	user, err := c.session.User()
	if errHandledPage(err, w, r, c) {
		return
	}

	if !town.IsMod(user) {
		four04(w, r)
		return
	}

	err = w.(*templateWriter).execute("TOWN", town)
	if err != nil {
		log.Errorf("Error executing townSettings template: %s", err)
	}

}

func townGet(w http.ResponseWriter, r *http.Request, c context) {
	key := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(key)
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	//Note: information on private towns is public
	// but the posts inside the town are not
	// which is why we don't check for privacy here
	// I may reconsider this

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   town,
	})
}

func townPostNew(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	input := &townInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	if input.Key == nil {
		errHandled(fail.New("A town key is required", input), w, r, c)
		return
	}

	if input.Name == nil {
		errHandled(fail.New("A town Name is required", input), w, r, c)
		return
	}

	if input.Description == nil {
		errHandled(fail.New("A town Description is required", input), w, r, c)
		return
	}

	if input.Longitude == nil {
		errHandled(fail.New("A town location longitude is required", input), w, r, c)
		return
	}

	if input.Latitude == nil {
		errHandled(fail.New("A town location latitude is required", input), w, r, c)
		return
	}

	private := false
	if input.Private != nil {
		private = *input.Private
	}

	//Rate limit new towns
	if errHandled(app.AttemptRequest(string(c.session.UserKey), townNewRequestType), w, r, c) {
		return
	}

	t, err := app.TownNew(*input.Key, *input.Name, *input.Description, u, *input.Longitude, *input.Latitude, private)
	if errHandled(err, w, r, c) {
		return
	}

	respondJsendCode(w, &JSend{
		Status: statusSuccess,
		Data:   t,
	}, http.StatusCreated)
}

func townPut(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	key := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(key)
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	input := &townInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	vertag := ""
	if input.VerTag != nil {
		vertag = *input.VerTag
	}

	town.SetVer(vertag)

	//update based on input

	if input.Name != nil {
		if errHandled(town.SetName(who, *input.Name), w, r, c) {
			return
		}
	}

	if input.Description != nil {
		if errHandled(town.SetDescription(who, *input.Description), w, r, c) {
			return
		}
	}

	if input.Latitude != nil || input.Longitude != nil {
		errHandled(fail.New("Sorry, location cannot be changed on existing towns."), w, r, c)
		return
	}

	if input.Color != nil {
		if errHandled(town.SetColor(who, *input.Color), w, r, c) {
			return
		}
	}

	if input.Information != nil {
		if errHandled(town.SetInformation(who, *input.Information), w, r, c) {
			return
		}
	}

	if input.Private != nil {
		if errHandled(town.SetPrivate(who, *input.Private), w, r, c) {
			return
		}
	}

	if errHandled(town.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func townPutImage(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	key := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(key)
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	input := &imageInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	vertag := ""
	if input.VerTag != nil {
		vertag = *input.VerTag
	}

	town.SetVer(vertag)

	if input.ImageKey != nil {
		img, err := app.ImageGet(*input.ImageKey)
		if errHandled(err, w, r, c) {
			return
		}
		if errHandled(town.SetHeaderImage(who, img, input.X0, input.Y0, input.X1, input.Y1), w, r, c) {
			return
		}
	}

	if errHandled(town.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func townDeleteImage(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	key := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(key)
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	input := &imageInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	vertag := ""
	if input.VerTag != nil {
		vertag = *input.VerTag
	}

	town.SetVer(vertag)
	if errHandled(town.RemoveHeaderImage(who), w, r, c) {
		return
	}

	if errHandled(town.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})

}

// Moderators

func townDeleteMod(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	key := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(key)
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	input := &townInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	//if they dont' specify the vertag, then make the update
	// regardless of what the current version they are looking at is
	vertag := town.Ver()
	if input.VerTag != nil {
		vertag = *input.VerTag
	}
	town.SetVer(vertag)

	if errHandled(town.RemoveModerator(who, who), w, r, c) {
		return
	}

	if errHandled(town.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func townPostMod(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	key := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(key)
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	input := &townInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.Moderator == nil {
		errHandled(fail.New("The field moderator is required", input), w, r, c)
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	//if they dont' specify the vertag, then make the update
	// regardless of what the current version they are looking at is
	vertag := town.Ver()
	if input.VerTag != nil {
		vertag = *input.VerTag
	}
	town.SetVer(vertag)

	newMod, err := app.UserGet(data.NewKey(*input.Moderator))
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(town.InviteModerator(who, newMod), w, r, c) {
		return
	}

	if errHandled(town.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})

}

func townPutMod(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	key := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(key)
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	vertag := town.Ver()

	town.SetVer(vertag)

	if errHandled(town.AcceptModeratorInvite(who), w, r, c) {
		return
	}

	if errHandled(town.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

// Private town invites
func townPostInvite(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	key := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(key)
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	input := &townInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.Invitee == nil && input.Email == nil {
		errHandled(fail.New("You must private a username or email to invite", input), w, r, c)
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	//if they dont' specify the vertag, then make the update
	// regardless of what the current version they are looking at is
	vertag := town.Ver()
	if input.VerTag != nil {
		vertag = *input.VerTag
	}
	town.SetVer(vertag)

	if input.Invitee != nil {
		invitee, err := app.UserGet(data.NewKey(*input.Invitee))
		if errHandled(err, w, r, c) {
			return
		}

		if errHandled(town.AddInvite(who, invitee), w, r, c) {
			return
		}

		if errHandled(town.Update(), w, r, c) {
			return
		}
		respondJsend(w, &JSend{
			Status: statusSuccess,
		})
		return
	}

	if input.Email != nil {
		if errHandled(town.AddInviteByEmail(who, *input.Email), w, r, c) {
			return
		}
		respondJsend(w, &JSend{
			Status: statusSuccess,
		})
		return
	}
}

// accept email invite
func townAcceptInvite(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	token := c.params.ByName("token")

	who, err := c.session.User()
	if errHandledPage(err, w, r, c) {
		return
	}

	town, err := app.TownAcceptEmailInvite(who, token)
	if errHandledPage(err, w, r, c) {
		return
	}

	http.Redirect(w, r, "/town/"+string(town.Key), http.StatusTemporaryRedirect)
}

func townDeleteInvite(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	key := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(key)
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	input := &townInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.Invitee == nil {
		errHandled(fail.New("The field invitee is required", input), w, r, c)
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	//if they dont' specify the vertag, then make the update
	// regardless of what the current version they are looking at is
	vertag := town.Ver()
	if input.VerTag != nil {
		vertag = *input.VerTag
	}
	town.SetVer(vertag)

	invitee, err := app.UserGet(data.NewKey(*input.Invitee))
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(town.RemoveInvite(who, invitee), w, r, c) {
		return
	}

	invitee.LeaveTown(town)

	if errHandled(town.Update(), w, r, c) {
		return
	}

	if errHandled(invitee.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func townSearch(w http.ResponseWriter, r *http.Request, c context) {
	// ?latitude=<latitude>
	// ?longitude=<longitude>
	// ?milesDistant=<miles>
	// ?northBounds=<northBounds>
	// ?southBounds=<southBounds>
	// ?westBounds=<westBounds>
	// ?eastBounds=<eastBounds>
	// ?search=<searchstring>
	// ?from=<index>&limit=100

	values := r.URL.Query()

	limit, err := strconv.Atoi(values.Get("limit"))
	if err != nil {
		limit = townDefaultRetrieveLImit
	}
	from, err := strconv.Atoi(values.Get("from"))
	if err != nil {
		from = 0
	}

	search := strings.TrimSpace(values.Get("search"))

	latitude := strings.TrimSpace(values.Get("latitude"))
	longitude := strings.TrimSpace(values.Get("longitude"))
	milesDistant := strings.TrimSpace(values.Get("milesDistant"))
	northBounds := strings.TrimSpace(values.Get("northBounds"))
	southBounds := strings.TrimSpace(values.Get("southBounds"))
	westBounds := strings.TrimSpace(values.Get("westBounds"))
	eastBounds := strings.TrimSpace(values.Get("eastBounds"))

	var towns []app.Town

	if search != "" {
		towns, err = app.TownSearch(search, from, limit)
		if errHandled(err, w, r, c) {
			return
		}
	} else if latitude != "" || longitude != "" {
		lat, err := strconv.ParseFloat(latitude, 64)
		if err != nil {
			errHandled(fail.New("Invalid Latitude", latitude), w, r, c)
			return
		}
		lng, err := strconv.ParseFloat(longitude, 64)
		if err != nil {
			errHandled(fail.New("Invalid Longitude", longitude), w, r, c)
			return
		}
		miles, err := strconv.ParseFloat(milesDistant, 64)
		if err != nil {
			miles = townDefaultMilesDistant
		}
		towns, err = app.TownSearchDistance(lng, lat, miles, from, limit)
		if errHandled(err, w, r, c) {
			return
		}

	} else if northBounds != "" && southBounds != "" && westBounds != "" && eastBounds != "" {
		north, err := strconv.ParseFloat(northBounds, 64)
		if err != nil {
			errHandled(fail.New("Invalid north bounds", northBounds), w, r, c)
			return
		}
		south, err := strconv.ParseFloat(southBounds, 64)
		if err != nil {
			errHandled(fail.New("Invalid south bounds", southBounds), w, r, c)
			return
		}
		west, err := strconv.ParseFloat(westBounds, 64)
		if err != nil {
			errHandled(fail.New("Invalid west bounds", westBounds), w, r, c)
			return
		}
		east, err := strconv.ParseFloat(eastBounds, 64)
		if err != nil {
			errHandled(fail.New("Invalid east bounds", eastBounds), w, r, c)
			return
		}
		towns, err = app.TownSearchArea(north, south, east, west, from, limit)
		if errHandled(err, w, r, c) {
			return
		}

	} else {
		errHandled(fail.New("You must specify either a Search value, a Latitude and Longitude "+
			"or a bounded area (NSEW) for searching"),
			w, r, c)
		return
	}
	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   towns,
	})

}

func newtownTemplate(w http.ResponseWriter, r *http.Request, c context) {
	var u *app.User
	var err error
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err = c.session.User()
	if errHandledPage(err, w, r, c) {
		return
	}

	err = w.(*templateWriter).execute("NEWTOWN", struct {
		GoogleMapsAPIKey string
		User             *app.User
	}{
		GoogleMapsAPIKey: private.GoogleMapsAPIKey,
		User:             u,
	})
	if err != nil {
		log.Errorf("Error executing newtown template: %s", err)
	}
}

func townsearchTemplate(w http.ResponseWriter, r *http.Request, c context) {
	var u *app.User
	var err error

	if c.session != nil {
		u, err = c.session.User()
		if errHandledPage(err, w, r, c) {
			return
		}
	}

	location, err := app.IPToLocation(ipAddress(r))
	if errHandledPage(err, w, r, c) {
		return
	}

	err = w.(*templateWriter).execute("TOWNSEARCH", struct {
		GoogleMapsAPIKey string
		User             *app.User
		Location         *app.IPLocation
	}{
		GoogleMapsAPIKey: private.GoogleMapsAPIKey,
		User:             u,
		Location:         location,
	})
	if err != nil {
		log.Errorf("Error executing townSearch template: %s", err)
	}

}

// invite requests

// request private town invite
func townPostInviteRequest(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	key := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(key)
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(town.RequestInvite(who), w, r, c) {
		return
	}

	respondJsendCode(w, &JSend{
		Status: statusSuccess,
	}, http.StatusCreated)
}

// accept private town invite
func townPutInviteRequest(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	input := &townInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.Invitee == nil {
		errHandled(fail.New("The field invitee is required", input), w, r, c)
		return
	}
	invitee, err := app.UserGet(data.NewKey(*input.Invitee))
	if errHandled(err, w, r, c) {
		return
	}

	key := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(key)
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	//if they dont' specify the vertag, then make the update
	// regardless of what the current version they are looking at is
	vertag := town.Ver()
	if input.VerTag != nil {
		vertag = *input.VerTag
	}
	town.SetVer(vertag)

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(town.AcceptInviteRequest(who, invitee), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})

}

// reject private town invite
func townDeleteInviteRequest(w http.ResponseWriter, r *http.Request, c context) {
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	input := &townInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.Invitee == nil {
		errHandled(fail.New("The field invitee is required", input), w, r, c)
		return
	}
	invitee, err := app.UserGet(data.NewKey(*input.Invitee))
	if errHandled(err, w, r, c) {
		return
	}

	key := data.NewKey(c.params.ByName("town"))
	town, err := app.TownGet(key)
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	//if they dont' specify the vertag, then make the update
	// regardless of what the current version they are looking at is
	vertag := town.Ver()
	if input.VerTag != nil {
		vertag = *input.VerTag
	}
	town.SetVer(vertag)

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(town.RejectInviteRequest(who, invitee), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}
