// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"fmt"
	"math"
	"net/mail"
	"path"
	"regexp"
	"strconv"
	"time"

	rt "git.townsourced.com/townsourced/gorethink/types"
	"github.com/timshannon/townsourced/app/email"
	"github.com/timshannon/townsourced/data"
	"github.com/timshannon/townsourced/fail"
)

// Town is a community in townsourced
type Town struct {
	Key data.Key `json:"key,omitempty"` //Unique Url Name, case insensitive

	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Information string      `json:"information,omitempty"` //markdown page for detailed information
	Moderators  []Moderator `json:"moderators,omitempty"`
	CreatorKey  data.Key    `json:"creatorKey,omitempty" gorethink:",omitempty"`

	HeaderImage data.UUID `json:"headerImage,omitempty"`
	Color       string    `json:"color,omitempty"`
	Location    rt.Point  `json:"location,omitempty" gorethink:",omitempty"`

	//Private towns are where only invitees can post or join the town, and only members can view
	// Note that you may be a member, and not an invitee, as in the announcements town
	Private        bool            `json:"private,omitempty"`
	Invites        []data.Key      `json:"invites,omitempty"`
	InviteRequests []InviteRequest `json:"inviteRequests,omitempty"`
	AutoModerator  struct {        //blacklists - what isn't allowed instead of what is
		Categories   []string       `json:"categories,omitempty"`   //automod posts in these categories
		MinUserDays  uint           `json:"minUserDays,omitempty"`  //automod if posted by user younger than x days MaxNumLinks  uint           `json:"maxNumLinks,omitempty"`  //automod if has more than x links in post
		MaxNumLinks  uint           `json:"maxNumLinks,omitempty"`  //automod if has more than x links in post
		Users        []data.Key     `json:"users,omitempty"`        //automod if post is submitted by one of these users
		RegexpReject []RegexpReason `json:"regexpReject,omitempty"` //automod if regexp matches with given reason
		//TODO: minUserRating
	} `json:"autoModerator,omitempty"`

	Population int `json:"population,omitempty" gorethink:",omitempty"` //Calculated field - not pulled from document

	data.Version
}

// RegexpReason is a reason for moderation tied to a regular expression
type RegexpReason struct {
	Regexp string `json:"regexp,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// Moderator is the record of moderator's start and end dates moderating a town
// used to determine any profit sharing for the town
type Moderator struct {
	Start      time.Time `json:"start,omitempty" gorethink:",omitempty"`
	End        time.Time `json:"end,omitempty" gorethink:",omitempty"`
	Username   data.Key  `json:"username,omitempty" gorethink:",omitempty"`
	InviteSent time.Time `json:"inviteSent,omitempty" gorethink:",omitempty"`
}

// InviteRequest contains a request for invitation to a Private town
type InviteRequest struct {
	Who  data.Key  `json:"who,omitempty"`
	When time.Time `json:"when,omitempty"`
}

var (
	//ErrTownExists is the error returned when a town already exists
	ErrTownExists = fail.New("A town already exists with the given url path, please choose another.")
	//ErrTownNotFound is the error returned when a town isn't found with the given key
	ErrTownNotFound = fail.New("Town not found")
	//ErrTownNotMod is the error returned when a user is trying to update a town, but they aren't a moderator
	ErrTownNotMod = fail.New("You cannot update this town because you are not currently a moderator of it.")
	//ErrTownDescriptionMax is the error when a town's description is too long
	ErrTownDescriptionMax = fail.New("A town's description can only be " + strconv.Itoa(townDescriptionMax) + " characters long.")
	//ErrTownNameMax is the error when a town's name is too long
	ErrTownNameMax = fail.New("A town's name can only be " + strconv.Itoa(townNameMax) + " characters long.")

	//ErrTownNotPrivate is the error when a Private update is made to a non-private town
	ErrTownNotPrivate = fail.New("You cannot make invite changes to a non-private town.")
	//ErrTownNoInvite is the error when a user tries to join a private town without an invite
	ErrTownNoInvite = fail.New("You cannot join this town because it is marked as private, and you do not currently have an invite.")
)

const (
	townHeaderImageWidth    = 1482
	townHeaderImageHeight   = 380
	townHeaderExpectedRatio = float64(townHeaderImageWidth) / float64(townHeaderImageHeight)
	townDescriptionMax      = 500
	townNameMax             = 200
	townKeyMax              = data.MaxKeyLength
	townMaxLinkDefault      = 5
	townModMaxRegexpLength  = 500
	townModMaxRexexpCount   = 100
	townSearchMaxRetrieve   = 1000
	townSearchMinDistance   = 0.5
	townSearchMaxDistance   = 1000.0
	townLocationUnit        = data.LocationUnitMile

	townInviteTokenExpire = 14 * 24 * time.Hour // two weeks

	//TownInviteAcceptPath is the url path used in building the forgot password url
	TownInviteAcceptPath = "towninvite"
)

// AnnouncementTown is the key of the townsourced announcements town
const AnnouncementTown = data.AnnouncementTown

var (
	townColorTest = regexp.MustCompile("^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$")
	rxMDLink      = regexp.MustCompile(`\[([^\[]+)\]\(([^\)]+)\)|((([A-Za-z]{3,9}:(?:\/\/)?)(?:[\-;:&=\+\$,\w]+@)?[A-Za-z0-9\.\-]+|(?:www\.|[\-;:&=\+\$,\w]+@)[A-Za-z0-9\.\-]+)((?:\/[\+~%\/\.\w\-_]*)?\??(?:[\-\+=&;%@\.\w_]*)#?(?:[\.\!\/\\\w]*))?)`) //markdown links & other urls
)

// TownNew creates a new town
func TownNew(key data.Key, name, description string, creator *User, longitude, latitude float64, private bool) (*Town, error) {
	town := &Town{
		Key:         key,
		Name:        name,
		Description: description,
		CreatorKey:  creator.Username,
		Moderators: []Moderator{Moderator{
			Start:    time.Now(),
			Username: creator.Username,
		}},
		Private: private,
	}

	town.AutoModerator.MaxNumLinks = townMaxLinkDefault

	err := town.validate()
	if err != nil {
		return nil, err
	}

	err = town.setLocation(creator, longitude, latitude)
	if err != nil {
		return nil, err
	}

	town.Rev()

	err = data.TownInsert(town, town.Key)
	if err != nil {
		return nil, err
	}

	err = town.index()
	if err != nil {
		return nil, err
	}

	err = creator.JoinTown(town)
	if err != nil {
		return nil, err
	}
	err = creator.Update()
	if err != nil {
		return nil, err
	}

	sub, msg, err := messages.use("msgTownNew").Execute(town)
	if err != nil {
		return nil, err
	}

	err = notificationNew(data.EmptyKey, creator.Username, sub, msg)
	if err != nil {
		return nil, err
	}

	//TODO: Task to notify any nearby users of the new town

	return town, nil
}

func (t *Town) index() error {
	if t.Private {
		// don't index private towns
		return nil
	}
	return data.TownIndex(t, t.Key)
}

func (t *Town) validate() error {
	if t.Name == "" {
		return fail.New("You must enter a town name")
	}
	if len(t.Name) > townNameMax {
		return ErrTownNameMax
	}

	if len(t.Description) > townDescriptionMax {
		return ErrTownDescriptionMax
	}

	if !urlify(string(t.Key)).is() {
		return fail.New("A town key can only contain letters, numbers and dashes")
	}

	_, err := TownGet(t.Key)
	if err == nil {
		return fail.NewFromErr(ErrTownExists, t.Key)
	}
	if err != ErrTownNotFound {
		return err
	}

	return nil
}

// TownGet retrieves a town
func TownGet(key data.Key) (*Town, error) {
	t := &Town{}

	err := data.TownGet(t, key)
	if err == data.ErrNotFound {
		return nil, ErrTownNotFound
	}
	if err != nil {
		return nil, err
	}

	err = t.setPopulation()
	if err != nil {
		return nil, err
	}

	return t, nil
}

// TownsGet retrieves a list of towns from the passed in keys
func TownsGet(keys ...data.Key) ([]Town, error) {
	var towns []Town
	if len(keys) == 0 {
		return towns, nil
	}

	err := data.Towns(&towns, keys...)
	if err != nil {
		return nil, err
	}

	for i := range towns {
		err = towns[i].setPopulation()
		if err != nil {
			return nil, err
		}
	}

	return towns, nil
}

// TownSearchDistance retrieves a town by a passed in location and distance from that location to search
func TownSearchDistance(longitude, latitude, milesDistant float64, from, limit int) ([]Town, error) {
	var towns []Town

	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(townSearchMaxRetrieve)))
	milesDistant = math.Min(math.Max(townSearchMinDistance, milesDistant), townSearchMaxDistance)

	latLng, err := data.NewLatLng(latitude, longitude)
	if err != nil {
		return towns, fail.NewFromErr(err, latLng, latLng)
	}

	qry := data.NewDistanceSearch(latLng, milesDistant, townLocationUnit)

	err = data.TownGetByLocation(&towns, qry, from, limit)
	if err == data.ErrNotFound {
		return towns, nil
	}
	if err != nil {
		return nil, err
	}

	for i := range towns {
		err = towns[i].setPopulation()
		if err != nil {
			return nil, err
		}
	}

	return towns, nil
}

// TownSearchArea retrieves all towns within the passed in area
func TownSearchArea(northBounds, southBounds, eastBounds, westBounds float64, from, limit int) ([]Town, error) {
	var towns []Town

	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(townSearchMaxRetrieve)))

	qry, err := data.NewAreaSearch(northBounds, southBounds, eastBounds, westBounds)
	if err != nil {
		return towns, fail.NewFromErr(err)
	}

	err = data.TownGetByLocation(&towns, qry, from, limit)
	if err == data.ErrNotFound {
		return towns, nil
	}
	if err != nil {
		return nil, err
	}

	for i := range towns {
		err = towns[i].setPopulation()
		if err != nil {
			return nil, err
		}
	}

	return towns, nil
}

// TownSearch searches for towns by their name or description
func TownSearch(search string, from, limit int) ([]Town, error) {
	var towns []Town

	limit = int(math.Min(math.Max(float64(1), float64(limit)), float64(townSearchMaxRetrieve)))

	result, err := data.TownGetBySearch(search, from, limit)

	if err == data.ErrNotFound {
		return towns, nil
	}

	if err != nil {
		return nil, err
	}

	towns = make([]Town, result.Count())

	for i := range towns {
		err = result.Next(&towns[i])
		if err != nil {
			return nil, err
		}

		err = towns[i].setPopulation()
		if err != nil {
			return nil, err
		}
	}

	return towns, nil
}

// SetVer  prepares the user data for an update
// based on the passed in vertag, if the vertag doesn't
// match the current record, then the update won't complete
func (t *Town) SetVer(verTag string) {
	t.VerTag = verTag
}

// Update updates the town
func (t *Town) Update() error {
	err := data.TownUpdate(t, t.Key)
	if err != nil {
		return err
	}
	return t.index()
}

// SetDescription sets the town's description
func (t *Town) SetDescription(who *User, newDescription string) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}
	if len(newDescription) > townDescriptionMax {
		return ErrTownDescriptionMax
	}

	t.Description = newDescription
	return nil

}

func (t *Town) setPopulation() error {
	pop, err := data.TownGetPopulation(t.Key)
	if err != nil {
		return err
	}

	t.Population = pop
	return nil
}

// SetInformation sets the town's information panel
func (t *Town) SetInformation(who *User, newInformation string) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	t.Information = newInformation
	return nil
}

// SetName sets the town's name
func (t *Town) SetName(who *User, newName string) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}
	if newName == "" {
		return fail.New("You must enter a town name")
	}
	if len(newName) > townNameMax {
		return ErrTownNameMax
	}

	t.Name = newName
	return nil
}

func (t *Town) setLocation(who *User, longitude, latitude float64) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	ll, err := data.NewLatLng(latitude, longitude)
	if err != nil {
		return fail.New(fmt.Sprintf("Error setting town location for town: %s", t.Name), latitude, longitude)
	}
	t.Location = rt.Point(ll)
	return nil
}

// SetColor sets the towns color theme
func (t *Town) SetColor(who *User, newColor string) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}
	if !townColorTest.MatchString(newColor) {
		return fail.New("Invalid color.  Color must be in hexidecimal format with a leading #")
	}

	t.Color = newColor
	return nil
}

// InviteModerator sends an invitation to a user that they need to accept before they become a moderator
func (t *Town) InviteModerator(who *User, newMod *User) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}
	mod := t.mod(newMod)
	if mod.active() {
		return fail.New("User is already a moderator", newMod.Username)
	}
	if mod.invited() {
		return fail.New("User has already been invited", newMod.Username)
	}
	t.Moderators = append(t.Moderators, Moderator{
		InviteSent: time.Now(),
		Username:   newMod.Username,
	})

	sub, msg, err := messages.use("msgTownModInvite").Execute(t)
	if err != nil {
		return err
	}

	err = who.SendMessage(newMod, sub, msg)
	if err != nil {
		return err
	}

	return nil
}

// AcceptModeratorInvite adds a new moderator to the town
func (t *Town) AcceptModeratorInvite(who *User) error {
	mod := t.mod(who)
	if mod.active() {
		//nothing to do
		return nil
	}

	if !mod.invited() {
		return fail.New("You were not invited to moderate this town")
	}

	for i := range t.Moderators {
		mod := t.Moderators[i]
		if mod.Username == who.Username {
			t.Moderators[i].Start = time.Now()
			return nil
		}
	}

	return fail.New("Invite not found")
}

// mod invited but not active yet
func (m *Moderator) invited() bool {
	return !m.InviteSent.IsZero() && m.Start.IsZero()
}

func (m *Moderator) active() bool {
	if m.Start.IsZero() {
		return false
	}
	return m.Start.Before(time.Now()) && (m.End.After(time.Now()) || m.End.IsZero())
}

// RemoveModerator removes a moderator from a town
func (t *Town) RemoveModerator(who *User, removeMod *User) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	if who.Username != removeMod.Username {
		return fail.New("Moderators can only remove themselves from being a moderator.")
	}

	for i := range t.Moderators {
		mod := t.Moderators[i]
		if mod.Username == removeMod.Username {
			t.Moderators[i].End = time.Now()
			return nil
		}
	}

	return nil
}

// IsMod is whether or not the passed in user is an active moderator
func (t *Town) IsMod(user *User) bool {
	return t.mod(user).active()
}

// returns  a mod entry for the passed in user
// if the user isn't a mod, then an empty mod is returned
// active mods will be returned instead of old inactive entries
// Admins are automatically mods in each town
func (t *Town) mod(user *User) *Moderator {
	var entries []*Moderator
	if user == nil {
		return &Moderator{}
	}

	if user.Admin {
		return &Moderator{
			Start:    user.Created,
			Username: user.Username,
		}
	}

	for i := range t.Moderators {
		mod := t.Moderators[i]
		if mod.Username == user.Username {
			if mod.active() {
				return &mod
			}
			entries = append(entries, &mod)
		}
	}

	if len(entries) == 0 {
		return &Moderator{}
	}

	//Find any pending invites
	for i := range entries {
		if entries[i].invited() {
			return entries[i]
		}
	}

	return entries[0]
}

// IsMember tests if the passed in user is a member of the town
func (t *Town) IsMember(user *User) bool {
	if user == nil {
		return false
	}

	for i := range user.TownKeys {
		if t.Key == user.TownKeys[i].Key {
			return true
		}
	}

	return false
}

// SetHeaderImage sets the header image for a given town
func (t *Town) SetHeaderImage(who *User, image *Image, x0, y0, x1, y1 float64) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	err := image.decode()
	if err != nil {
		return err
	}

	if x0 < 0 || y0 < 0 || x1 < 0 || y1 < 0 {
		x0 = 0
		x1 = float64(image.image.Bounds().Dx())
		y0 = 0
		y1 = float64(image.image.Bounds().Dy())
	}

	if (x1 - x0) < 250 {
		x0 = 0
		x1 = float64(image.image.Bounds().Dx())
	}

	if (y1 - y0) < 64 {
		y0 = 0
		y1 = float64(image.image.Bounds().Dy())
	}

	//delete any existing header image
	if t.HeaderImage != data.EmptyUUID {
		err := data.ImageDelete(t.HeaderImage)
		if err != nil {
			return err
		}
	}

	ratio := (x1 - x0) / (y1 - y0)

	//give a little wiggle room for rounding
	if ratio < townHeaderExpectedRatio-.5 || ratio > townHeaderExpectedRatio+.5 {
		height := image.image.Bounds().Dy()
		width := int(float64(height) * townHeaderExpectedRatio)

		err = image.cropCenter(who, width, height)
		if err != nil {
			return err
		}
	} else {
		err = image.crop(who, round(x0), round(y0), round(x1), round(y1))
		if err != nil {
			return err
		}

	}

	err = image.resize(who, townHeaderImageWidth, townHeaderImageHeight)
	if err != nil {
		return err
	}

	image.InUse = true
	err = image.update()
	if err != nil {
		return err
	}

	t.HeaderImage = image.Key

	err = image.encode()
	if err != nil {
		return err
	}

	return nil
}

// RemoveHeaderImage removes the header image for a given town
func (t *Town) RemoveHeaderImage(who *User) error {
	if t.HeaderImage == data.EmptyUUID {
		// nothing to do
		return nil
	}

	err := data.ImageDelete(t.HeaderImage)
	if err != nil {
		return err
	}

	t.HeaderImage = data.EmptyUUID

	return nil
}

func (t *Town) autoModerate(p *Post) (string, error) {
	// categories
	for _, category := range t.AutoModerator.Categories {
		if p.Category == category {
			return "This town does not allow posts from the " + category + " category.", nil
		}
	}

	//Users

	for _, u := range t.AutoModerator.Users {
		if p.Creator == u {
			return "You are not allowed to post to this town.  Contact the town moderator(s) for more information.", nil
		}
	}

	postUser, err := p.creator()
	if err != nil {
		return "", err
	}

	if postUser.Created.After(time.Now().AddDate(0, 0, int(t.AutoModerator.MinUserDays)*-1)) {
		return fmt.Sprintf("This town does not allow posts by users whose accounts are younger than %d day(s)", t.AutoModerator.MinUserDays), nil
	}

	//links
	// TODO: I'm betting there are ways around my rx with commonmark, and it may end up being safer and faster
	// to simply use the commonmark implementation to count links.  We'll start with this for now though.
	if len(rxMDLink.FindAll([]byte(p.Content), int(t.AutoModerator.MaxNumLinks+1))) > int(t.AutoModerator.MaxNumLinks) {
		return fmt.Sprintf("This town does not allow posts with more than %d links", t.AutoModerator.MaxNumLinks), nil
	}

	//regexp
	for _, rr := range t.AutoModerator.RegexpReject {
		rxFind, err := regexp.Compile(rr.Regexp)
		if err != nil {
			//skip the bad regexp
			continue
		}

		if rxFind.Find([]byte(p.Title)) != nil || rxFind.Find([]byte(p.Content)) != nil {
			return rr.Reason, nil
		}
	}

	return "", nil
}

// ensureAnnoucementTown checks if the announcement town has been created yet
// and if not, creates it
func ensureAnnouncementTown() error {
	_, err := TownGet(AnnouncementTown)
	if err == nil {
		return nil
	}

	if err != ErrTownNotFound {
		return err
	}
	town := &Town{
		Key:  data.NewKey(AnnouncementTown),
		Name: "Townsourced Announcements",
		Description: "In this town you will find announcements from townsourced about new features, and changes in the site. " +
			"All new users automatically join this Town and you can leave at any time from the town page",
		Private: false,
	}

	town.AutoModerator.MaxNumLinks = townMaxLinkDefault

	town.Rev() // generate a vertag

	// The announcments town is private, so no one can post to it, but everyone is automatically a member, so they can view posts

	return data.TownInsert(town, town.Key)
}

// SetPrivate sets the town to Private, or removes the private setting
func (t *Town) SetPrivate(who *User, private bool) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	t.Private = private
	return nil
}

// AddInvite adds an invite to a private town
func (t *Town) AddInvite(who *User, invitee *User) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	if !t.Private {
		return ErrTownNotPrivate
	}

	if t.Invited(invitee) {
		return fail.New("User has already been invited")
	}

	for i := range t.InviteRequests {
		if t.InviteRequests[i].Who == invitee.Username {
			// remove from invited requested
			return t.AcceptInviteRequest(who, invitee)
		}
	}

	t.Invites = append(t.Invites, invitee.Username)

	sub, msg, err := messages.use("msgTownPrivateInvite").Execute(t)
	if err != nil {
		return err
	}

	err = who.SendMessage(invitee, sub, msg)
	if err != nil {
		return err
	}

	return nil
}

// AddInviteByEmail adds an invite to a private town by email
func (t *Town) AddInviteByEmail(who *User, emailAddr string) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	if !t.Private {
		return ErrTownNotPrivate
	}

	if !emailTest.MatchString(emailAddr) {
		return fail.New("Invalid email address format.", emailAddr)
	}

	tokenURL := ""
	token := Random(256)
	tokenURL = baseURL + "/" + path.Join(TownInviteAcceptPath, token)

	err := data.TempTokenSet(t.Key, token, townInviteTokenExpire)
	if err != nil {
		return err
	}

	sub, body, err := messages.use("emailTownInvite").Execute(struct {
		URL  string
		Town *Town
		Who  string
	}{
		URL:  tokenURL,
		Town: t,
		Who:  who.DisplayName(),
	})

	if err != nil {
		return err
	}

	err = email.Send(email.DefaultFrom, &mail.Address{
		Address: emailAddr,
	}, sub, body)

	if err != nil {
		return fmt.Errorf("Error sending town email invite: %s", err)
	}

	return nil
}

// TownAcceptEmailInvite associates an email invite to a user and adds them to the invite list
func TownAcceptEmailInvite(invitee *User, token string) (*Town, error) {
	townKey := data.EmptyKey
	err := data.TempTokenGet(&townKey, token)

	if err != nil {
		return nil, err
	}

	t, err := TownGet(townKey)
	if err != nil {
		return nil, err
	}

	if !t.Private {
		return t, nil
	}

	if t.Invited(invitee) {
		return t, nil
	}

	t.Invites = append(t.Invites, invitee.Username)

	err = t.Update()
	if err != nil {
		return nil, err
	}

	err = invitee.JoinTown(t)
	if err != nil {
		return nil, err
	}

	err = invitee.Update()
	if err != nil {
		return nil, err
	}

	err = data.TempTokenExpire(token)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// RemoveInvite removes an invite from a private town
func (t *Town) RemoveInvite(who *User, toRemove *User) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	if !t.Private {
		return ErrTownNotPrivate
	}

	for i := range t.Invites {
		if t.Invites[i] == toRemove.Username {
			//remove user from invites
			t.Invites = append(t.Invites[:i], t.Invites[i+1:]...)
			return nil
		}
	}

	return ErrUserNotFound
}

// RequestInvite requests an invite to be a member of a private town. It sends a PM to
// all the moderators of the town
func (t *Town) RequestInvite(who *User) error {
	if !t.Private {
		// no need for an invite if the town isn't private
		return nil
	}

	if t.Invited(who) {
		//already invited no need to send the request
		return nil
	}

	if t.IsMember(who) {
		//already a member, no need to send the request
		return nil
	}

	for i := range t.InviteRequests {
		if t.InviteRequests[i].Who == who.Username {
			// already requested
			return nil
		}
	}

	t.InviteRequests = append(t.InviteRequests, InviteRequest{
		Who:  who.Username,
		When: time.Now(),
	})

	err := t.Update()
	if err != nil {
		return err
	}

	sub, msg, err := messages.use("msgTownPrivateRequestInvite").Execute(t)
	if err != nil {
		return err
	}

	//send PM to all active moderators of the town
	for i := range t.Moderators {
		if t.Moderators[i].active() {
			mod, err := UserGet(t.Moderators[i].Username)
			if err != nil {
				return err
			}
			err = who.SendMessage(mod, sub, msg)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

// AcceptInviteRequest accepts a request for invitation to a private town,
// removes the request and adds the user to the town
func (t *Town) AcceptInviteRequest(who *User, invitee *User) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	if !t.Private {
		return ErrTownNotPrivate
	}

	found := false
	for i := range t.InviteRequests {
		if t.InviteRequests[i].Who == invitee.Username {
			found = true
			// remove invite request
			t.InviteRequests = append(t.InviteRequests[:i], t.InviteRequests[i+1:]...)
			break
		}
	}

	if !found {
		if t.Invited(invitee) {
			//request already completed
			return nil
		}
		return fail.New("User " + invitee.DisplayName() + " has not requested an invite to this town")
	}

	t.Invites = append(t.Invites, invitee.Username)
	err := invitee.JoinTown(t)
	if err != nil {
		return err
	}

	err = t.Update()
	if err != nil {
		return err
	}

	err = invitee.Update()
	if err != nil {
		return err
	}

	sub, msg, err := messages.use("msgTownPrivateAcceptInviteRequest").Execute(t)
	if err != nil {
		return err
	}

	err = who.SendMessage(invitee, sub, msg)
	if err != nil {
		return err
	}

	return nil
}

// RejectInviteRequest rejects a request for invitation to a private town
func (t *Town) RejectInviteRequest(who *User, invitee *User) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	if !t.Private {
		return ErrTownNotPrivate
	}

	found := false
	for i := range t.InviteRequests {
		if t.InviteRequests[i].Who == invitee.Username {
			found = true
			// remove invite request
			t.InviteRequests = append(t.InviteRequests[:i], t.InviteRequests[i+1:]...)
			break
		}
	}

	if !found {
		return nil
	}

	err := t.Update()
	if err != nil {
		return err
	}

	sub, msg, err := messages.use("msgTownPrivateRejectInviteRequest").Execute(t)
	if err != nil {
		return err
	}

	err = who.SendMessage(invitee, sub, msg)
	if err != nil {
		return err
	}

	return nil
}

// Invited checks if the passed in user has an invite to this town
// Note: this should only be used to check if a user is invited or not
// Not whether or not posts are or aren't visible in a town
// FOr the Announcement's town a user is a member, but not invited.
// They become a member automatically
func (t *Town) Invited(who *User) bool {
	// mods get auto invited
	if t.mod(who).active() {
		return true
	}

	for i := range t.Invites {
		if t.Invites[i] == who.Username {
			return true
		}
	}

	return false
}

// AddAutoModCategory adds a category to the auto moderator
func (t *Town) AddAutoModCategory(who *User, category string) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	found := false
	for i := range postCategories {
		if postCategories[i] == category {
			found = true
			break
		}
	}

	if !found {
		return fail.New("Invalid Auto Moderation post category")
	}

	// remove if already added
	err := t.RemoveAutoModCategory(who, category)
	if err != nil {
		return err
	}

	t.AutoModerator.Categories = append(t.AutoModerator.Categories, category)

	return nil
}

// RemoveAutoModCategory removes a category from the auto moderator
func (t *Town) RemoveAutoModCategory(who *User, category string) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	for i := range t.AutoModerator.Categories {
		if t.AutoModerator.Categories[i] == category {
			t.AutoModerator.Categories = append(t.AutoModerator.Categories[:i], t.AutoModerator.Categories[i+1:]...)
			return nil
		}
	}

	return nil
}

// SetAutoModMinUserDays sets the minimum days old a user must be to post to this town
func (t *Town) SetAutoModMinUserDays(who *User, minUserDays uint) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}
	t.AutoModerator.MinUserDays = minUserDays
	return nil
}

// SetAutoModMaxNumLinks sets the maxiumum number of links allowed in a post to this town
func (t *Town) SetAutoModMaxNumLinks(who *User, maxNumLinks uint) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}
	t.AutoModerator.MaxNumLinks = maxNumLinks
	return nil
}

// AddAutoModUser adds a user to the auto moderator
func (t *Town) AddAutoModUser(who *User, username data.Key) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	_, err := UserGet(username)
	if err != nil {
		return err
	}

	//remove if exists already
	err = t.RemoveAutoModUser(who, username)
	if err != nil {
		return err
	}

	t.AutoModerator.Users = append(t.AutoModerator.Users, username)

	return nil
}

// RemoveAutoModUser removes a user from the auto moderator
func (t *Town) RemoveAutoModUser(who *User, username data.Key) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	for i := range t.AutoModerator.Users {
		if t.AutoModerator.Users[i] == username {
			t.AutoModerator.Users = append(t.AutoModerator.Users[:i], t.AutoModerator.Users[i+1:]...)
			return nil
		}
	}

	return nil
}

// AddAutoModRegexp adds a new regular expression to auto moderator
func (t *Town) AddAutoModRegexp(who *User, expr, reason string) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	if len(expr) > townModMaxRegexpLength {
		return fail.New(fmt.Sprintf("The expression is too long.  The max length is %d", townModMaxRegexpLength))
	}

	if len(t.AutoModerator.RegexpReject) >= townModMaxRexexpCount {
		return fail.New(fmt.Sprintf("This town has too many expressions added, you'll need to remove some before "+
			"you can add more.  The max number of expressions is %d", townModMaxRexexpCount))
	}

	_, err := regexp.Compile(expr)
	if err != nil {
		return fail.NewFromErr(err)
	}

	for _, rr := range t.AutoModerator.RegexpReject {
		if rr.Regexp == expr {
			return fail.New("An auto moderator reason already exists for this expression.", expr)
		}
	}

	t.AutoModerator.RegexpReject = append(t.AutoModerator.RegexpReject, RegexpReason{
		Regexp: expr,
		Reason: reason,
	})

	return nil
}

// RemoveAutoModRegexp removes a regular expression from the auto moderator
func (t *Town) RemoveAutoModRegexp(who *User, expr string) error {
	if !t.mod(who).active() {
		return ErrTownNotMod
	}

	for i := range t.AutoModerator.RegexpReject {
		if t.AutoModerator.RegexpReject[i].Regexp == expr {
			t.AutoModerator.RegexpReject = append(t.AutoModerator.RegexpReject[:i], t.AutoModerator.RegexpReject[i+1:]...)
			return nil
		}
	}

	return fail.New("No auto moderator reason exists for this expression.", expr)
}

// CanSearch is whether or not the passed in user can run searches against this town
// e.g. town isn't private or they are a member
func (t *Town) CanSearch(who *User) bool {
	if !t.Private {
		return true
	}

	return t.IsMember(who)
}

// Posts returns the posts for the given town
func (t *Town) Posts(who *User, category string, since time.Time, limit int, showModerated bool) ([]Post, error) {
	return PostGetByTowns(who, []data.Key{t.Key}, category, since, limit, showModerated)
}
