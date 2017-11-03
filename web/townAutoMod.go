package web

import (
	"net/http"

	"github.com/timshannon/townsourced/app"
	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/fail"
)

type townAutoModInput struct {
	Category    *string   `json:"category,omitempty"`
	MinUserDays *uint     `json:"minUserDays,omitempty"`
	MaxNumLinks *uint     `json:"maxNumLinks,omitempty"`
	User        *data.Key `json:"user,omitempty"`
	Regexp      *string   `json:"regexp,omitempty"`
	Reason      *string   `json:"reason,omitempty"`
}

func townPostAutoModCategory(w http.ResponseWriter, r *http.Request, c context) {
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

	input := &townAutoModInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.Category == nil {
		errHandled(fail.New("The field category is required", input), w, r, c)
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(town.AddAutoModCategory(who, *input.Category), w, r, c) {
		return
	}

	if errHandled(town.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func townDeleteAutoModCategory(w http.ResponseWriter, r *http.Request, c context) {
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

	input := &townAutoModInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.Category == nil {
		errHandled(fail.New("The field category is required", input), w, r, c)
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(town.RemoveAutoModCategory(who, *input.Category), w, r, c) {
		return
	}

	if errHandled(town.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func townPutAutoMod(w http.ResponseWriter, r *http.Request, c context) {
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

	input := &townAutoModInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	if input.MinUserDays != nil {
		if errHandled(town.SetAutoModMinUserDays(who, *input.MinUserDays), w, r, c) {
			return
		}
	}

	if input.MaxNumLinks != nil {
		if errHandled(town.SetAutoModMaxNumLinks(who, *input.MaxNumLinks), w, r, c) {
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

func townPostAutoModUser(w http.ResponseWriter, r *http.Request, c context) {
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

	input := &townAutoModInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.User == nil {
		errHandled(fail.New("The field user is required", input), w, r, c)
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(town.AddAutoModUser(who, *input.User), w, r, c) {
		return
	}

	if errHandled(town.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func townDeleteAutoModUser(w http.ResponseWriter, r *http.Request, c context) {
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

	input := &townAutoModInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.User == nil {
		errHandled(fail.New("The field user is required", input), w, r, c)
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(town.RemoveAutoModUser(who, *input.User), w, r, c) {
		return
	}

	if errHandled(town.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func townPostAutoModRegexp(w http.ResponseWriter, r *http.Request, c context) {
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

	input := &townAutoModInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.Regexp == nil || input.Reason == nil {
		errHandled(fail.New("The fields regexp and reason are required", input), w, r, c)
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(town.AddAutoModRegexp(who, *input.Regexp, *input.Reason), w, r, c) {
		return
	}

	if errHandled(town.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func townDeleteAutoModRegexp(w http.ResponseWriter, r *http.Request, c context) {
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

	input := &townAutoModInput{}
	err = parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.Regexp == nil {
		errHandled(fail.New("The field regexp is required", input), w, r, c)
		return
	}

	who, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(town.RemoveAutoModRegexp(who, *input.Regexp), w, r, c) {
		return
	}

	if errHandled(town.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}
