// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"strconv"
	"time"

	log "git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/townsourced/app"
	"git.townsourced.com/townsourced/townsourced/data"
	"git.townsourced.com/townsourced/townsourced/fail"
)

// rate limit the number of new users that can be created by the same ip address
var userNewRequestType = app.RequestType{
	Type:         "userNew",
	FreeAttempts: 2,
	Scale:        15 * time.Second,
	Range:        15 * time.Minute,
	MaxWait:      1 * time.Minute,
}

type userInput struct {
	Username      *data.Key `json:"username,omitempty"`
	Email         *string   `json:"email,omitempty"`
	Password      *string   `json:"password,omitempty"`
	NewPassword   *string   `json:"newPassword,omitempty"`
	Name          *string   `json:"name,omitempty"`
	NotifyPost    *bool     `json:"notifyPost,omitempty"`
	NotifyComment *bool     `json:"notifyComment,omitempty"`
	VerTag        *string   `json:"verTag,omitempty"`

	EmailPrivateMsg     *bool `json:"emailPrivateMsg,omitempty"`
	EmailPostMention    *bool `json:"emailPostMention,omitempty"`
	EmailCommentMention *bool `json:"emailCommentMention,omitempty"`
	EmailCommentReply   *bool `json:"emailCommentReply,omitempty"`
	EmailPostComment    *bool `json:"emailPostComment,omitempty"`
}

func userTemplate(w http.ResponseWriter, r *http.Request, c context) {
	u, self := userHandleGet(w, r, c)
	if u == nil {
		return
	}

	err := w.(*templateWriter).execute("USER", struct {
		*app.User
		Self bool `json:"self"`
	}{User: u, Self: self})
	if err != nil {
		log.Errorf("Error executing user template: %s", err)
	}
}

// userHandleGet retrieves a user based on the passed in request and context
// if user is nil, then an error has already been properly written to w
func userHandleGet(w http.ResponseWriter, r *http.Request, c context) (u *app.User, self bool) {
	var err error
	userKey := c.params.ByName("user")

	if c.session != nil {
		if userKey == string(c.session.UserKey) || userKey == app.UsernameSelf || userKey == "" {
			u, err = c.session.User()
			self = true

			if errHandled(err, w, r, c) {
				u = nil
				return
			}
			return
		}
	}

	if userKey == app.UsernameSelf || userKey == "" {
		// blank or me with no logged in user == unauthorized access
		u = nil
		unauthorized(w, r)
		return
	}

	u, err = app.UserGet(data.NewKey(userKey))
	if err == app.ErrUserNotFound {
		u = nil
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		u = nil
		return
	}
	u.ClearPrivate()

	return
}

func userGet(w http.ResponseWriter, r *http.Request, c context) {
	u, _ := userHandleGet(w, r, c)
	if u == nil {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   u,
	})
}

func userPost(w http.ResponseWriter, r *http.Request, c context) {
	if errHandled(app.AttemptRequest(ipAddress(r), app.RequestType{}), w, r, c) {
		return
	}

	if demoMode {
		if c.session != nil {
			u, err := c.session.User()
			if errHandled(err, w, r, c) {
				return
			}
			if !u.Admin {
				// in demo mode, only admin accounts can create new users
				errHandled(fail.New("Signups are currently disabled.  Contact info@townsourced.com "+
					"for more information"), w, r, c)
				return
			}
		} else {
			errHandled(fail.New("Signups are currently disabled.  Contact info@townsourced.com "+
				"for more information"), w, r, c)
			return
		}

	}

	input := &userInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	if input.Username == nil {
		errHandled(fail.New("A username is required", input), w, r, c)
		return
	}
	if input.Email == nil {
		errHandled(fail.New("An email address is required", input), w, r, c)
		return
	}

	empty := ""
	if input.Password == nil {
		input.Password = &empty
	}

	u, err := app.UserNew(*input.Username, *input.Email, *input.Password)
	if errHandled(err, w, r, c) {
		return
	}

	if errHandled(app.AttemptRequest(ipAddress(r), userNewRequestType), w, r, c) {
		return
	}

	if errHandled(setSessionCookie(w, r, u, false), w, r, c) {
		return
	}
	respondJsendCode(w, &JSend{
		Status: statusSuccess,
		Data:   u,
	}, http.StatusCreated)
}

func userPut(w http.ResponseWriter, r *http.Request, c context) {
	if c.params.ByName("user") != app.UsernameSelf {
		four04(w, r)
		return
	}

	if c.session == nil {
		unauthorized(w, r)
		return
	}
	input := &userInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	vertag := ""
	if input.VerTag != nil {
		vertag = *input.VerTag
	}

	u.SetVer(vertag)

	if input.Name != nil {
		u.SetName(*input.Name)
	}

	if input.Email != nil {
		if input.Password == nil {
			errHandled(fail.New("A password is required to change your email address", input), w, r, c)
			return
		}
		err = u.SetEmail(*input.Email, *input.Password)
		if errHandled(err, w, r, c) {
			return
		}
	}

	if input.NewPassword != nil && input.Password != nil {
		err = u.SetPassword(*input.Password, *input.NewPassword)
		if errHandled(err, w, r, c) {
			return
		}
	}

	if input.NotifyComment != nil {
		u.NotifyComment = *input.NotifyComment
	}
	if input.NotifyPost != nil {
		u.NotifyPost = *input.NotifyPost
	}

	if input.EmailCommentMention != nil {
		u.EmailCommentMention = *input.EmailCommentMention
	}

	if input.EmailPrivateMsg != nil {
		u.EmailPrivateMsg = *input.EmailPrivateMsg
	}
	if input.EmailPostMention != nil {
		u.EmailPostMention = *input.EmailPostMention
	}
	if input.EmailCommentReply != nil {
		u.EmailCommentReply = *input.EmailCommentReply
	}
	if input.EmailPostComment != nil {
		u.EmailPostComment = *input.EmailPostComment
	}

	err = u.Update()
	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func emailGet(w http.ResponseWriter, r *http.Request, c context) {
	//TODO: ratelimit different from normal?
	email := r.URL.Query().Get("email")
	if errHandled(app.UserEmailExists(email), w, r, c) {
		return
	}
	respondJsend(w, &JSend{
		Status: statusSuccess,
	})

}

type imageInput struct {
	ImageKey *data.UUID `json:"imageKey,omitempty"`
	X0       float64    `json:"x0"`
	Y0       float64    `json:"y0"`
	X1       float64    `json:"x1"`
	Y1       float64    `json:"y1"`
	VerTag   *string    `json:"verTag,omitempty"`
}

func userPutImage(w http.ResponseWriter, r *http.Request, c context) {
	if c.params.ByName("user") != app.UsernameSelf {
		four04(w, r)
		return
	}
	if c.session == nil {
		unauthorized(w, r)
		return
	}
	input := &imageInput{}
	err := parseInput(r, input)
	if errHandled(err, w, r, c) {
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	vertag := ""
	if input.VerTag != nil {
		vertag = *input.VerTag
	}

	if input.ImageKey == nil {
		errHandled(fail.New("An imageKey is required", input), w, r, c)
		return
	}

	img, err := app.ImageGet(*input.ImageKey)
	if errHandled(err, w, r, c) {
		return
	}

	u.SetVer(vertag)

	if errHandled(u.SetProfileImage(img, input.X0, input.Y0, input.X1, input.Y1), w, r, c) {
		return
	}

	if errHandled(u.Update(), w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func userGetImage(w http.ResponseWriter, r *http.Request, c context) {
	// ?icon
	values := r.URL.Query()

	userKey := c.params.ByName("user")

	u, err := app.UserGet(data.NewKey(userKey))

	if err == app.ErrUserNotFound {
		four04(w, r)
		return
	}

	if errHandled(err, w, r, c) {
		return
	}

	var i *app.Image

	if _, ok := values["icon"]; ok {
		i, err = app.ImageGet(u.ProfileIcon)
		if errHandled(err, w, r, c) {
			return
		}
	} else {
		i, err = app.ImageGet(u.ProfileImage)
		if errHandled(err, w, r, c) {
			return
		}
	}

	serveImage(w, r, i)
}

func userGetTowns(w http.ResponseWriter, r *http.Request, c context) {
	u, _ := userHandleGet(w, r, c)
	if u == nil {
		return
	}

	towns, err := u.Towns()
	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   towns,
	})
}

func userGetPosts(w http.ResponseWriter, r *http.Request, c context) {
	//?since=<time>&limit=100
	// ?status=<status>
	// ?saved
	userKey := c.params.ByName("user")

	if userKey == app.UsernameSelf {
		if c.session == nil {
			unauthorized(w, r)
			return
		}
		userKey = string(c.session.UserKey)
	}

	u, err := app.UserGet(data.NewKey(userKey))
	if err == app.ErrUserNotFound {
		four04(w, r)
		return
	}

	if errHandled(err, w, r, c) {
		return
	}

	var who *app.User

	if c.session != nil {
		if u.Username == c.session.UserKey {
			who = u
		} else {

			who, err = c.session.User()
			if errHandled(err, w, r, c) {
				return
			}
		}
	}

	values := r.URL.Query()
	since, limit, err := sinceLimitValues(values, postPageLimit)
	if errHandled(err, w, r, c) {
		return
	}

	posts, err := u.Posts(who, values.Get("status"), since, limit)

	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   posts,
	})
}

func userPostTown(w http.ResponseWriter, r *http.Request, c context) {
	if c.params.ByName("user") != app.UsernameSelf {
		four04(w, r)
		return
	}

	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}
	t, err := app.TownGet(data.NewKey(c.params.ByName("town")))
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	u.SetVer(u.Ver())

	if errHandled(u.JoinTown(t), w, r, c) {
		return
	}

	err = u.Update()
	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func userDeleteTown(w http.ResponseWriter, r *http.Request, c context) {
	if c.params.ByName("user") != app.UsernameSelf {
		four04(w, r)
		return
	}

	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}
	t, err := app.TownGet(data.NewKey(c.params.ByName("town")))
	if err == app.ErrTownNotFound {
		four04(w, r)
		return
	}
	if errHandled(err, w, r, c) {
		return
	}

	u.SetVer(u.Ver())

	u.LeaveTown(t)

	err = u.Update()
	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})
}

func userGetSavedPosts(w http.ResponseWriter, r *http.Request, c context) {
	// ?status=<status>
	// ?from=<from>
	// ?limit=<limit>
	userKey := c.params.ByName("user")

	if userKey == app.UsernameSelf {
		if c.session == nil {
			unauthorized(w, r)
			return
		}
		userKey = string(c.session.UserKey)
	}

	u, err := app.UserGet(data.NewKey(userKey))
	if err == app.ErrUserNotFound {
		four04(w, r)
		return
	}

	if errHandled(err, w, r, c) {
		return
	}

	var who *app.User

	if c.session != nil {
		if u.Username == c.session.UserKey {
			who = u
		} else {

			who, err = c.session.User()
			if errHandled(err, w, r, c) {
				return
			}
		}
	}

	values := r.URL.Query()

	limit, err := strconv.Atoi(values.Get("limit"))
	if err != nil {
		limit = 20
	}

	from, err := strconv.Atoi(values.Get("from"))
	if err != nil {
		from = 0
	}

	posts, err := u.GetSavedPosts(who, values.Get("status"), from, limit)

	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   posts,
	})
}

func userPostSavedPost(w http.ResponseWriter, r *http.Request, c context) {
	if c.params.ByName("user") != app.UsernameSelf {
		four04(w, r)
		return
	}

	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	post, err := app.PostGet(data.ToUUID(c.params.ByName("post")))

	if errHandled(err, w, r, c) {
		return
	}

	u.SetVer(u.Ver())

	u.SavePost(post)

	err = u.Update()
	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})

}

func userDeleteSavedPost(w http.ResponseWriter, r *http.Request, c context) {
	if c.params.ByName("user") != app.UsernameSelf {
		four04(w, r)
		return
	}

	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	post, err := app.PostGet(data.ToUUID(c.params.ByName("post")))

	if errHandled(err, w, r, c) {
		return
	}

	u.SetVer(u.Ver())

	u.RemoveSavedPost(post)

	err = u.Update()
	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
	})

}

func userGetComments(w http.ResponseWriter, r *http.Request, c context) {
	// ?since=<since>
	// ?limit=<limit>

	var who *app.User
	var err error

	if c.session != nil {
		who, err = c.session.User()
		if errHandled(err, w, r, c) {
			return
		}
	}

	userKey := c.params.ByName("user")

	u, err := app.UserGet(data.NewKey(userKey))
	if err == app.ErrUserNotFound {
		four04(w, r)
		return
	}

	if errHandled(err, w, r, c) {
		return
	}

	values := r.URL.Query()

	since, limit, err := sinceLimitValues(values, commentListSizeDefault)
	if errHandled(err, w, r, c) {
		return
	}

	comments, err := u.Comments(who, since, limit)

	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   comments,
	})
}

func userGetMatch(w http.ResponseWriter, r *http.Request, c context) {
	// ?match=<match>
	// ?limit=<limit>
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	values := r.URL.Query()

	limit, err := strconv.Atoi(values.Get("limit"))
	if err != nil {
		limit = 10
	}

	users, err := app.UserGetMatch(values.Get("match"), limit)
	if errHandled(err, w, r, c) {
		return
	}

	respondJsend(w, &JSend{
		Status: statusSuccess,
		Data:   users,
	})
}

func welcomeTemplate(w http.ResponseWriter, r *http.Request, c context) {
	// ?latitude=<latitude>
	// ?longitude=<longitude>
	if c.session == nil {
		unauthorized(w, r)
		return
	}

	u, err := c.session.User()
	if errHandledPage(err, w, r, c) {
		return
	}

	values := r.URL.Query()
	lat, err := strconv.ParseFloat(values.Get("latitude"), 64)
	if err != nil {
		lat = 0
	}
	lng, err := strconv.ParseFloat(values.Get("longitude"), 64)
	if err != nil {
		lng = 0
	}

	var location *app.IPLocation

	if lat == 0 && lng == 0 {
		location, err = app.IPToLocation(ipAddress(r))
		if errHandledPage(err, w, r, c) {
			return
		}

		lat = location.Latitude
		lng = location.Longitude
	}

	var towns []app.Town

	towns, err = app.TownSearchDistance(lng, lat, 1000, 0, 90)

	if errHandledPage(err, w, r, c) {
		return
	}

	err = w.(*templateWriter).execute("WELCOME", struct {
		User     *app.User
		Towns    []app.Town
		Location *app.IPLocation
		CSRF     string
	}{
		User:     u,
		Towns:    towns,
		Location: location,
		CSRF:     c.session.CSRFToken,
	})

	if err != nil {
		log.Errorf("Error executing WELCOME template: %s", err)
	}
}
